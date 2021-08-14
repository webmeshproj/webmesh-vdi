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

	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	desktopsv1 "github.com/kvdi/kvdi/apis/desktops/v1"
	v1 "github.com/kvdi/kvdi/apis/meta/v1"

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
func LookupClusterByName(c client.Client, name string) (*appv1.VDICluster, error) {
	found := &appv1.VDICluster{}
	return found, c.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: metav1.NamespaceAll}, found)
}

// IsMarkedForDeletion returns true if the given cluster is marked for deletion.
func IsMarkedForDeletion(cr *appv1.VDICluster) bool {
	return cr.GetDeletionTimestamp() != nil
}

// GetDesktopLabels returns the labels to apply to components for a desktop.
func GetDesktopLabels(c *appv1.VDICluster, desktop *desktopsv1.Session) map[string]string {
	labels := desktop.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[v1.UserLabel] = desktop.GetUser()
	labels[v1.VDIClusterLabel] = c.GetName()
	labels[v1.ComponentLabel] = "desktop"
	labels[v1.DesktopNameLabel] = desktop.GetName()
	return labels
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
	ctx           context.Context
	cancel        func()
	buf           io.ReadWriter
	pod           *corev1.Pod
	containerName string
}

// NewLogFollower returns a new LogFollower for the given pod and container.
func NewLogFollower(pod *corev1.Pod, containerName string) *LogFollower {
	buf := new(bytes.Buffer)
	ctx, cancel := context.WithCancel(context.Background())
	return &LogFollower{
		ctx:           ctx,
		cancel:        cancel,
		buf:           buf,
		pod:           pod,
		containerName: containerName,
	}
}

// Stream will start the log stream
func (l *LogFollower) Stream(follow bool) error {
	if DefaultClient == nil {
		return errors.New("There is no raw client configured for scraping logs")
	}

	var err error
	defer func() {
		if err != nil {
			l.Close()
		}
	}()

	// No matter what, first retrieve logs with no follow. Running with follow true will not return
	// previous logs.
	podLogOpts := corev1.PodLogOptions{Follow: false, Container: l.containerName}
	req := DefaultClient.CoreV1().Pods(l.pod.Namespace).GetLogs(l.pod.Name, &podLogOpts)
	var podLogs io.ReadCloser
	podLogs, err = req.Stream(l.ctx)
	if err != nil {
		return err
	}
	defer podLogs.Close()

	// Copy the logs to the buffer
	if _, err = io.Copy(l, podLogs); err != nil {
		return err
	}

	if !follow {
		return nil
	}

	// create a new stream
	now := metav1.Now()
	podLogOpts = corev1.PodLogOptions{Follow: true, Container: l.containerName, SinceTime: &now}
	req = DefaultClient.CoreV1().Pods(l.pod.Namespace).GetLogs(l.pod.Name, &podLogOpts)
	var followLogs io.ReadCloser
	followLogs, err = req.Stream(l.ctx)
	if err != nil {
		return err
	}

	// Spawn a copy in a goroutine
	go func() {
		defer l.Close()
		if _, err = io.Copy(l, followLogs); err != nil {
			if !strings.Contains(err.Error(), "canceled") {
				fmt.Println("Error copying pod logs to buffer:", err)
			}
		}
	}()

	return nil
}

// Read reads data from the log buffer
func (l *LogFollower) Read(p []byte) (int, error) {
	return l.buf.Read(p)
}

// Write writes data to the log buffer
func (l *LogFollower) Write(p []byte) (int, error) {
	return l.buf.Write(p)
}

// Close cancels the log stream
func (l *LogFollower) Close() error {
	l.cancel()
	return nil
}

func getClientSet() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
