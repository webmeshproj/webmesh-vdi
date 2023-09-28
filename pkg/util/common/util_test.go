/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

package common

import (
	"os"
	"reflect"
	"testing"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var testLogger = logf.Log.WithName("test")

func TestBoolPointer(t *testing.T) {
	if !*BoolPointer(true) {
		t.Error("Expected pointer to true")
	}
	if *BoolPointer(false) {
		t.Error("Expected pointer to false")
	}
}

func TestInt32Ptr(t *testing.T) {
	if *Int32Ptr(10) != 10 {
		t.Error("Expected pointer to 10")
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

func TestAppendStringIfMissing(t *testing.T) {
	sl := []string{"a", "b", "c"}
	sl = AppendStringIfMissing(sl, "a", "b", "c")
	if !reflect.DeepEqual(sl, []string{"a", "b", "c"}) {
		t.Error("Did not get correct slice back:", sl)
	}
	sl = AppendStringIfMissing(sl, "d", "e")
	if !reflect.DeepEqual(sl, []string{"a", "b", "c", "d", "e"}) {
		t.Error("Did not get correct slice back:", sl)
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
	file, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	_, err = file.Write(fakeResolvConf)
	if err != nil {
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
	passw, err := GeneratePassword(16)
	if err != nil {
		t.Fatal(err)
	}
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

func TestPrintVersion(t *testing.T) {
	PrintVersion(testLogger)
}
