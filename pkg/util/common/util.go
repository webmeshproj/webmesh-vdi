package common

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/tinyzimmer/kvdi/version"

	"github.com/go-logr/logr"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/bcrypt"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// BoolPointer returns a pointer to the given boolean
func BoolPointer(b bool) *bool { return &b }

// Int64Ptr returns a pointer to the given int64
func Int64Ptr(i int64) *int64 { return &i }

// Int32Ptr returns a pointer to the given int32
func Int32Ptr(i int32) *int32 { return &i }

// StringSliceContains returns true if the given string exists in the
// given slice.
func StringSliceContains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// StringSliceRemove returns a new slice with the given element removed.
func StringSliceRemove(ss []string, s string) []string {
	newSlice := make([]string, 0)
	for _, x := range ss {
		if x != s {
			newSlice = append(newSlice, x)
		}
	}
	return newSlice
}

// AppendStringIfMissing will append the given element(s) to the slice only if
// they are not already present.
func AppendStringIfMissing(sl []string, s ...string) []string {
ArgLoop:
	for _, x := range s {
		for _, ele := range sl {
			if ele == x {
				continue ArgLoop
			}
		}
		sl = append(sl, x)
	}
	return sl
}

func PrintVersion(log logr.Logger) {
	log.Info(fmt.Sprintf("kVDI Version: %s", version.Version))
	log.Info(fmt.Sprintf("kVDI Commit: %s", version.GitCommit))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

// ParseFlagsAndSetupLogging is a utility function to setup logging
// and parse any provided flags.
func ParseFlagsAndSetupLogging() {
	pflag.CommandLine.AddFlagSet(zap.FlagSet())
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	logf.SetLogger(zap.Logger())
}

// resolvConf is the path to the resolv config file when running inside a cluster.
var resolvConf = "/etc/resolv.conf"

// GetClusterSuffix returns the cluster suffix as parsed from the resolvconf.
// If we cannot read the file we return an empty string. This is a safeguard
// against irregular short-name resolution inside different cluster setups.
func GetClusterSuffix() string {
	resolvconf, err := ioutil.ReadFile(resolvConf)
	if err != nil {
		return ""
	}
	re := regexp.MustCompile("search.*")
	match := re.FindString(string(resolvconf))
	if strings.TrimSpace(match) == "" {
		return ""
	}
	fields := strings.Fields(match)
	return fields[len(fields)-1]
}

// GeneratePassword generates a password with the given length
func GeneratePassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	buf := make([]rune, length)
	for i := range buf {
		buf[i] = chars[rand.Intn(len(chars))]
	}
	return string(buf)
}

// hashCost is the cost to use for generating salts from passwords
var hashCost = bcrypt.MinCost

// HashPassword creates a salt from a password for storing in a database
func HashPassword(passw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(passw), hashCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// PasswordMatchesHash returns true if the given password matches the given salt.
func PasswordMatchesHash(passw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(passw)) == nil
}
