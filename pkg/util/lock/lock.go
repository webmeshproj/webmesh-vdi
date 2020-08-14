package lock

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

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
	// labels to apply to the configmap
	labels map[string]string
	// the cm object backing this lock
	cm *corev1.ConfigMap
}

// New returns a new lock. If timeout is a value less than zero, then no expiration
// is placed on the lock. A safeguard for deadlocks is still in place due to the
// OwnerReference to the pod that holds the lock on the configmap.
func New(c client.Client, name string, timeout time.Duration) *Lock {
	return &Lock{
		client:  c,
		name:    name,
		timeout: timeout,
		labels:  map[string]string{},
		cm:      &corev1.ConfigMap{},
	}
}

// WithLabels configures labels to add to the resources associated with this lock.
func (l *Lock) WithLabels(labels map[string]string) *Lock {
	l.labels = labels
	return l
}

// GetName returns the name of this lock.
func (l *Lock) GetName() string { return l.name }

// GetNamespace returns the namespace of this lock.
func (l *Lock) GetNamespace() string { return l.namespace }

// GetTimeout returns the timeout for this lock.
func (l *Lock) GetTimeout() time.Duration { return l.timeout }

// GetCMData returns the data to apply to the configmap for this lock.
func (l *Lock) GetCMData() map[string]string {
	if l.GetTimeout() > 0 {
		return map[string]string{expireKey: strconv.FormatInt(time.Now().Add(l.timeout).Unix(), 10)}
	}
	return map[string]string{}
}

// Acquire will attempt to acquire the lock, retrying until the lock is either
// acquired or the timeout is reached.
func (l *Lock) Acquire() error {
	lockLogger.Info("Acquiring lock", "Lock.Name", l.GetName())
	pod, err := k8sutil.GetThisPod(l.client)
	if err != nil {
		lockLogger.Error(err, "Error retrieving current pod, could not acquire lock")
		return err
	}

	l.namespace = pod.GetNamespace()

	failTimeout := time.Now().Add(l.GetTimeout())

	var ctx context.Context
	var cancel func()
	if l.GetTimeout() > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), l.GetTimeout())
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	defer cancel()

	cm := newConfigMapForLock(l, pod)
	nn := types.NamespacedName{Name: cm.GetName(), Namespace: cm.GetNamespace()}

	for {

		err := l.client.Create(ctx, cm)

		// the lock was acquired
		if err == nil {
			lockLogger.Info("Lock acquired", "Lock.Name", l.GetName())
			break
		}

		// we couldn't acquire the lock
		if !kerrors.IsAlreadyExists(err) {
			lockLogger.Error(err, "Error trying to create configmap, could not acquire lock")
			return err
		}

		lockLogger.Info("Lock is currently held, checking status of existing lock")
		existingLock := &corev1.ConfigMap{}
		if err := l.client.Get(context.TODO(), nn, existingLock); err != nil {
			if kerrors.IsNotFound(err) {
				continue
			}
			lockLogger.Error(err, "Error looking up existing lock, could not acquire lock")
			return err
		}

		if err := l.checkExistingLockExpiry(ctx, existingLock); err != nil {
			return err
		}

		if time.Now().After(failTimeout) {
			lockLogger.Info("Timeout reached before we could acquire a lock")
			return errors.New("Failed to acquire lock in the given time limit")
		}

		lockLogger.Info("Current lock is still active, trying again in 1 second...")
		time.Sleep(time.Duration(1) * time.Second)
	}

	// read the cm object into memory for quicker release
	return l.client.Get(context.TODO(), nn, l.cm)
}

func (l *Lock) checkExistingLockExpiry(ctx context.Context, existingLock *corev1.ConfigMap) error {
	expiresAt, ok := existingLock.Data[expireKey]
	if !ok {
		if l.GetTimeout() > 0 {
			if err := l.releaseStaleLock(ctx, existingLock); err != nil {
				lockLogger.Error(err, "Failed to release stale lock, could not acquire lock")
				return err
			}
		}
		return nil
	}
	if expireTime, err := strconv.ParseInt(expiresAt, 10, 64); err == nil {
		if time.Now().After(time.Unix(expireTime, 0)) {
			if err := l.releaseStaleLock(ctx, existingLock); err != nil {
				lockLogger.Error(err, fmt.Sprintf("Failed to release stale lock, could not acquire lock %s", err.Error()))
				return err
			}
		}
	}
	return nil
}

// Release will delete the configmap, releasing the lock.
func (l *Lock) Release() error {
	lockLogger.Info("Releasing lock", "Lock.Name", l.name)
	if err := l.client.Delete(context.TODO(), l.cm); err != nil {
		if !kerrors.IsNotFound(err) {
			lockLogger.Error(err, fmt.Sprintf("Error releasing lock: %s", err.Error()))
			return err
		}
	}
	return nil
}

// releaseStaleLock removes a stale lock from kubernetes
func (l *Lock) releaseStaleLock(ctx context.Context, cm *corev1.ConfigMap) error {
	lockLogger.Info("Releasing stale lock", "PreviousOwner", cm.OwnerReferences[0])
	if err := l.client.Delete(ctx, cm); err != nil {
		if !kerrors.IsNotFound(err) {
			lockLogger.Error(err, fmt.Sprintf("Error releasing lock: %s", err.Error()))
			return err
		}
	}
	return nil
}

// newConfigMapForLock returns a new configmap for locking.
func newConfigMapForLock(l *Lock, pod *corev1.Pod) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      l.GetName(),
			Namespace: pod.GetNamespace(),
			Labels:    l.labels,
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
		Data: l.GetCMData(),
	}
}
