package common

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestBoolPointer(t *testing.T) {
	if !*BoolPointer(true) {
		t.Error("Expected pointer to true")
	}
	if *BoolPointer(false) {
		t.Error("Expected pointer to false")
	}
}

func TestInt64Ptr(t *testing.T) {
	if *Int64Ptr(10) != 10 {
		t.Error("Expected pointer to 10")
	}
}

func TestStringSliceContains(t *testing.T) {
	sl := []string{"a", "b", "c"}
	if StringSliceContains(sl, "d") {
		t.Error("String slice should not contain d")
	}
	if !StringSliceContains(sl, "c") {
		t.Error("String slice should contain c")
	}
}

func TestStringSliceRemove(t *testing.T) {
	sl := []string{"a", "b", "c"}
	if !reflect.DeepEqual(StringSliceRemove(sl, "a"), []string{"b", "c"}) {
		t.Error("New slice should be [b, c]")
	}
}

func TestParseFlagsAndSetupLogging(t *testing.T) {
	ParseFlagsAndSetupLogging()
}

func makeTempResolvConf(invalid bool) (string, error) {
	var fakeResolvConf []byte
	if invalid {
		fakeResolvConf = []byte(`
nameserver 8.8.8.8
nameserver 8.8.4.4
ndots ...
`)
	} else {
		fakeResolvConf = []byte(`
nameserver 8.8.8.8
nameserver 8.8.4.4
search default.svc.cluster.local svc.cluster.local cluster.local
ndots ...
`)
	}
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(file.Name(), fakeResolvConf, 0644); err != nil {
		return "", err
	}
	return file.Name(), nil
}

func TestGetClusterSuffix(t *testing.T) {
	// override the resolvConf value used by the function
	var err error
	if resolvConf, err = makeTempResolvConf(false); err != nil {
		t.Fatal(err)
	}

	if suffix := GetClusterSuffix(); suffix != "cluster.local" {
		t.Error("Expected cluster.local as the cluster suffix")
	}

	os.Remove(resolvConf)

	if suffix := GetClusterSuffix(); suffix != "" {
		t.Error("Expected empty string for cluster suffix")
	}

	// make an invalid one
	if resolvConf, err = makeTempResolvConf(true); err != nil {
		t.Fatal(err)
	}

	defer os.Remove(resolvConf)

	if suffix := GetClusterSuffix(); suffix != "" {
		t.Error("Expected empty string for cluster suffix")
	}
}

func TestPasswordFunctions(t *testing.T) {
	passw := GeneratePassword(16)
	if len(passw) != 16 {
		t.Error("Generated password is the wrong length")
	}

	hash, err := HashPassword(passw)
	if err != nil {
		t.Error("Unexpected error hashing password")
	}

	if !PasswordMatchesHash(passw, hash) {
		t.Error("Expected hash to match password")
	}

	// override hash cost to force a hashing error
	hashCost = 10000000
	if _, err = HashPassword(passw); err == nil {
		t.Error("Expected error for using invalid cost")
	}
}
