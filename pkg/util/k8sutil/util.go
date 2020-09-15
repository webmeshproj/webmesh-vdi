package k8sutil

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DefaultClient represents a client for performing raw CRUD operations against the
// Kubernetes API
var DefaultClient *kubernetes.Clientset

// init tries to create a DefaultClient for raw CRUD operations. If this fails, then any Manager
// would probably also fail to start anyway.
func init() {
	var err error
	if DefaultClient, err = getClientSet(); err != nil {
		fmt.Println("Unable to initialze in-cluster client, some functionality will be disabled")
	}
}

// LookupClusterByName fetches the VDICluster with the given name
func LookupClusterByName(c client.Client, name string) (*v1alpha1.VDICluster, error) {
	found := &v1alpha1.VDICluster{}
	return found, c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: metav1.NamespaceAll}, found)
}

// IsMarkedForDeletion returns true if the given cluster is marked for deletion.
func IsMarkedForDeletion(cr *v1alpha1.VDICluster) bool {
	return cr.GetDeletionTimestamp() != nil
}

// SetCreationSpecAnnotation sets an annotation with a checksum of the desired
// spec of the object.
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
	annotations[v1.CreationSpecAnnotation] = fmt.Sprintf("%x", h.Sum(nil))
	meta.SetAnnotations(annotations)
	return nil
}

// CreationSpecsEqual returns true if the two objects spec annotations are equal.
func CreationSpecsEqual(m1 metav1.ObjectMeta, m2 metav1.ObjectMeta) bool {
	m1ann := m1.GetAnnotations()
	m2ann := m2.GetAnnotations()
	spec1, ok := m1ann[v1.CreationSpecAnnotation]
	if !ok {
		return false
	}
	spec2, ok := m2ann[v1.CreationSpecAnnotation]
	if !ok {
		return false
	}
	return spec1 == spec2
}

// GetThisPodName attempts to return the name of the running pod from the environment.
func GetThisPodName() (string, error) {
	if podName := os.Getenv("POD_NAME"); podName != "" {
		return podName, nil
	}
	return "", errors.New("No POD_NAME in the environment")
}

// GetThisPodNamespace attempts to return the namespace of the running pod from the environment.
func GetThisPodNamespace() (string, error) {
	if podNS := os.Getenv("POD_NAMESPACE"); podNS != "" {
		return podNS, nil
	}
	return "", errors.New("No POD_NAMESPACE in the environment")
}

// GetThisPod attempts to return the full pod object of the requesting instance.
func GetThisPod(c client.Client) (*corev1.Pod, error) {
	podName, err := GetThisPodName()
	if err != nil {
		return nil, err
	}
	podNamespace, err := GetThisPodNamespace()
	if err != nil {
		return nil, err
	}
	nn := types.NamespacedName{Name: podName, Namespace: podNamespace}
	pod := &corev1.Pod{}
	return pod, c.Get(context.TODO(), nn, pod)
}

// LogFollower implements a ReadCloser for reading logs from a container in a pod.
type LogFollower struct {
	cancel func()
	buf    *bytes.Buffer
}

// Read reads data from the log buffer
func (l *LogFollower) Read(p []byte) (int, error) {
	return l.buf.Read(p)
}

// Close cancels the log stream
func (l *LogFollower) Close() error {
	l.cancel()
	return nil
}

// GetPodLogs attempts to return the logs for the given pod instance.
func GetPodLogs(pod *corev1.Pod, containerName string, follow bool) (io.ReadCloser, error) {
	if DefaultClient == nil {
		return nil, errors.New("There is no raw client configured for scraping logs")
	}

	// No matter what, first retrieve logs with no follow. Running with follow true will not return
	// previous logs.
	podLogOpts := corev1.PodLogOptions{Follow: false, Container: containerName}
	req := DefaultClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return nil, err
	}
	defer podLogs.Close()

	// Create a new buffer for the response
	follower := &LogFollower{
		buf:    new(bytes.Buffer),
		cancel: func() {},
	}

	// Copy the logs to the buffer
	if _, err = io.Copy(follower.buf, podLogs); err != nil {
		return nil, err
	}

	// If we aren't following, return here
	if !follow {
		return follower, nil
	}

	// create ctx and a cancel func
	var ctx context.Context
	ctx, follower.cancel = context.WithCancel(context.Background())

	// create a new stream
	podLogOpts = corev1.PodLogOptions{Follow: true, Container: containerName}
	req = DefaultClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	followLogs, err := req.Stream(ctx) // use the context tied to the cancel func
	if err != nil {
		follower.Close()
		return nil, err
	}

	// Spawn a copy in a gorouting
	go func() {
		defer followLogs.Close()
		if _, err = io.Copy(follower.buf, followLogs); err != nil {
			follower.Close()
			if !strings.Contains(err.Error(), "canceled") {
				fmt.Println("Error copying pod logs to buffer:", err)
			}
		}
	}()

	// return the buffer and the cancel function
	return follower, nil
}

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
