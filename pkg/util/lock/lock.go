package lock

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/util/common"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// expireKey is the key in the configmap where we store the expiry data
const expireKey = "expiresAt"

// lockLogger is a logger interface for lock events
var lockLogger = logf.Log.WithName("lock")

// Lock implements a simple locking mechanism backed by a configmap. When the
// lock is acquired, a configmap is made with an owner reference to the running
// pod. Another lock attempt will block until it can create the configmap with the
// same name, or if the lock expires.
type Lock struct {
	// the k8s client
	client client.Client
	// the timeout for the lock
	timeout time.Duration
	// the name of the lock
	name string
	// the namespace for the lock
	namespace string
}

// New returns a new lock
func New(c client.Client, name string, timeout time.Duration) *Lock {
	return &Lock{
		client:  c,
		name:    name,
		timeout: timeout,
	}
}

// Acquire will attempt to acquire the lock, retrying until the lock is either
// acquired of the timeout is reached.
func (l *Lock) Acquire() error {
	lockLogger.Info("Acquiring lock", "Lock.Name", l.name)
	pod, err := l.getCurrentPod()
	if err != nil {
		lockLogger.Error(err, "Error retrieving current pod, could not acquire lock")
		return err
	}

	l.namespace = pod.GetNamespace()

	failTimeout := time.Now().Add(l.timeout)
	ctx, cancel := context.WithTimeout(context.Background(), l.timeout)
	defer cancel()

	for {

		lock := newConfigMapForLock(l, pod)

		if err := l.client.Create(ctx, lock); err != nil {

			if !kerrors.IsAlreadyExists(err) {
				lockLogger.Error(err, "Error trying to create configmap, could not acquire lock")
				return err
			}
			existingLock := &corev1.ConfigMap{}
			nn := types.NamespacedName{Name: lock.GetName(), Namespace: lock.GetNamespace()}
			if err := l.client.Get(context.TODO(), nn, existingLock); err != nil {
				if kerrors.IsNotFound(err) {
					continue
				}
				lockLogger.Error(err, "Error looking up existing lock, could not acquire lock")
				return err
			}
			if expiresAt, ok := existingLock.Data[expireKey]; !ok {
				if err := l.releaseStaleLock(ctx, existingLock); err != nil {
					lockLogger.Error(err, "Failed to release stale lock, could not acquire lock")
					return err
				}
			} else if expireTime, err := strconv.ParseInt(expiresAt, 10, 64); err != nil {
				if time.Now().After(time.Unix(expireTime, 0)) {
					if err := l.releaseStaleLock(ctx, existingLock); err != nil {
						lockLogger.Error(err, "Failed to release stale lock, could not acquire lock")
						return err
					}
				}
			}

			if time.Now().After(failTimeout) {
				lockLogger.Info("Timeout reached before we could acquire a lock")
				return errors.New("Failed to acquire lock in the given time limit")
			}

			lockLogger.Info("Lock is currently held, trying again in 2 seconds...")
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}

		lockLogger.Info("Lock acquired", "Lock.Name", l.name)
		break
	}
	return nil
}

// Release will delete the configmap, releasing the lock.
func (l *Lock) Release() error {
	nn := types.NamespacedName{Name: l.name, Namespace: l.namespace}
	found := &corev1.ConfigMap{}
	if err := l.client.Get(context.TODO(), nn, found); err != nil {
		if kerrors.IsNotFound(err) {
			lockLogger.Info("Someone already cleaned up this lock", "Lock.Name", l.name)
			return nil
		}
		return err
	}
	lockLogger.Info("Releasing lock", "Lock.Name", l.name)
	return l.client.Delete(context.TODO(), found)
}

// releaseStaleLock removes a stale lock from kubernetes
func (l *Lock) releaseStaleLock(ctx context.Context, lock *corev1.ConfigMap) error {
	lockLogger.Info("Releasing stale lock", "PreviousOwner", lock.OwnerReferences[0])
	return client.IgnoreNotFound(l.client.Delete(ctx, lock))
}

// getCurrentPod returns the curently running pod based off environment variables
// populated from the metadata of the instance.
func (l *Lock) getCurrentPod() (*corev1.Pod, error) {
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		return nil, errors.New("Cannot get lock, no POD_NAME in environment")
	}
	podNamespace := os.Getenv("POD_NAMESPACE")
	if podNamespace == "" {
		return nil, errors.New("Cannot get lock, no POD_NAMESPACE in environment")
	}
	nn := types.NamespacedName{Name: podName, Namespace: podNamespace}
	pod := &corev1.Pod{}
	return pod, l.client.Get(context.TODO(), nn, pod)
}

// newConfigMapForLock returns a new configmap for locking.
func newConfigMapForLock(l *Lock, pod *corev1.Pod) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      l.name,
			Namespace: pod.GetNamespace(),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "v1",
					Kind:               "Pod",
					Name:               pod.GetName(),
					UID:                pod.GetUID(),
					Controller:         common.BoolPointer(true),
					BlockOwnerDeletion: common.BoolPointer(false),
				},
			},
		},
		Data: map[string]string{
			expireKey: strconv.FormatInt(time.Now().Add(l.timeout).Unix(), 10),
		},
	}
}
