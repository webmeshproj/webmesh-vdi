package common

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

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

// PrintVersion will dump version info to the given log interface.
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

// StopRetry is returned to tell the Retry function to stop retrying.
type StopRetry struct{ Err error }

// Error implements the error interface
func (s *StopRetry) Error() string { return s.Err.Error() }

// Retry will retry the given function until either the maximum attempts is reached or
// a stop error is returned.
func Retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if stop, ok := err.(*StopRetry); ok {
			return stop.Err
		}
		// user can pass -1 to retry indefinitely
		if attempts--; attempts > 0 || attempts < 0 {
			// Add some randomness to prevent creating a Thundering Herd

			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, f)
		}
		return err
	}

	return nil
}

// TarDirectoryToTempFile will create a gzipped tarball of the given directory,
// write it to a tempfile, and return the path to the file.
func TarDirectoryToTempFile(srcPath string) (string, error) {
	targetDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	baseDir := filepath.Base(srcPath)
	outFile := filepath.Join(targetDir, fmt.Sprintf("%s.tar.gz", baseDir))

	var fwriter *os.File
	fwriter, err = os.Create(outFile)
	if err != nil {
		return "", err
	}
	defer fwriter.Close()

	gzw := gzip.NewWriter(fwriter)
	defer gzw.Close()

	tarball := tar.NewWriter(gzw)
	defer tarball.Close()

	fmt.Println("Archiving", srcPath, "to", outFile)

	err = filepath.Walk(srcPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, srcPath))

			if err := tarball.WriteHeader(header); err != nil {
				return err
			}

			if info.IsDir() {
				fmt.Println("Skipping dir:", path)
				return nil
			}

			// skip symlinks for now
			if !info.Mode().IsRegular() {
				fmt.Println("Skipping symlink or irregular file:", path)
				return nil
			}

			fmt.Println("Opening file:", path)
			file, err := os.Open(path)

			// in case a file gets deleted while we are in the middle of
			// traversing
			if err != nil && !os.IsNotExist(err) {
				fmt.Println("File open error", err, "path:", path)
				return err
			}

			defer file.Close()
			fmt.Println("Copying file:", path)
			_, err = io.Copy(tarball, file)
			return err
		})

	if err != nil {
		if cleanErr := os.RemoveAll(targetDir); cleanErr != nil {
			fmt.Println("Failed to clean up failed tar directory:", cleanErr)
		}
		return "", err
	}

	fmt.Println("Finished archive:", outFile)
	return outFile, nil
}
