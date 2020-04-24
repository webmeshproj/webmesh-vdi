package util

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"golang.org/x/crypto/bcrypt"

	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func BoolPointer(b bool) *bool { return &b }

func Int64Ptr(i int64) *int64 { return &i }

func StringSliceContains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

func StringSliceRemove(ss []string, s string) []string {
	newSlice := make([]string, 0)
	for _, x := range ss {
		if x != s {
			newSlice = append(newSlice, x)
		}
	}
	return newSlice
}

func ParseFlagsAndSetupLogging() {
	pflag.CommandLine.AddFlagSet(zap.FlagSet())
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	logf.SetLogger(zap.Logger())
}

func IsMarkedForDeletion(cr *v1alpha1.VDICluster) bool {
	return cr.GetDeletionTimestamp() != nil
}

func SetCreationSpecAnnotation(meta *metav1.ObjectMeta, obj runtime.Object) error {
	annotations := meta.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	spec, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	h := sha256.New()
	if _, err := h.Write(spec); err != nil {
		return err
	}
	annotations[v1alpha1.CreationSpecAnnotation] = fmt.Sprintf("%x", h.Sum(nil))
	meta.SetAnnotations(annotations)
	return nil
}

func CreationSpecsEqual(m1 metav1.ObjectMeta, m2 metav1.ObjectMeta) bool {
	m1ann := m1.GetAnnotations()
	m2ann := m2.GetAnnotations()
	spec1, ok := m1ann[v1alpha1.CreationSpecAnnotation]
	if !ok {
		return false
	}
	spec2, ok := m2ann[v1alpha1.CreationSpecAnnotation]
	if !ok {
		return false
	}
	return spec1 == spec2
}

func GetClusterSuffix() string {
	resolvconf, err := ioutil.ReadFile("/etc/resolv.conf")
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

func DNSNames(svcName, svcNamespace string) []string {
	return []string{
		svcName,
		fmt.Sprintf("%s.%s", svcName, svcNamespace),
		fmt.Sprintf("%s.%s.svc", svcName, svcNamespace),
		fmt.Sprintf("%s.%s.svc.%s", svcName, svcNamespace, GetClusterSuffix()),
	}
}

func HeadlessDNSNames(podName, svcName, svcNamespace string) []string {
	return append(DNSNames(svcName, svcNamespace), []string{
		fmt.Sprintf("%s.%s", podName, svcName),
		fmt.Sprintf("%s.%s.%s", podName, svcName, svcNamespace),
		fmt.Sprintf("%s.%s.%s.svc", podName, svcName, svcNamespace),
		fmt.Sprintf("%s.%s.%s.svc.%s", podName, svcName, svcNamespace, GetClusterSuffix()),
	}...)
}

func StatefulSetDNSNames(svcName, svcNamespace string, replicas int32) []string {
	dnsNames := DNSNames(svcName, svcNamespace)
	for i := int32(0); i < replicas; i++ {
		podName := fmt.Sprintf("%s-%d", svcName, i)
		dnsNames = append(dnsNames,
			fmt.Sprintf("%s.%s", podName, svcName),
			fmt.Sprintf("%s.%s.%s", podName, svcName, svcNamespace),
			fmt.Sprintf("%s.%s.%s.svc", podName, svcName, svcNamespace),
			fmt.Sprintf("%s.%s.%s.svc.%s", podName, svcName, svcNamespace, GetClusterSuffix()),
		)
	}
	return dnsNames
}

func DesktopShortURL(desktop *v1alpha1.Desktop) string {
	return fmt.Sprintf("%s.%s.svc", desktop.GetName(), desktop.GetNamespace())
}

func LookupClusterByName(c client.Client, name string) (*v1alpha1.VDICluster, error) {
	found := &v1alpha1.VDICluster{}
	return found, c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: metav1.NamespaceAll}, found)
}

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

func HashPassword(passw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(passw), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func PasswordMatchesHash(passw, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(passw)) == nil
}
