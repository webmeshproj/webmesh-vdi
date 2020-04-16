package util

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func BoolPointer(b bool) *bool { return &b }

func Int64Ptr(i int64) *int64 { return &i }

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
	h.Write(spec)
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
	return fmt.Sprintf("%s.%s.%s", desktop.GetName(), desktop.GetName(), desktop.GetNamespace())
}
