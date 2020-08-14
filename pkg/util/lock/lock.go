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
	// labels to apply to the configmap
	labels map[string]string
	// the pod that owns this lock
	pod *corev1.Pod
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
	}
}

// WithLabels configures labels to add to the resources associated with this lock.
func (l *Lock) WithLabels(labels map[string]string) *Lock {
	l.labels = labels
	return l
}

// GetName returns the name of this lock.
func (l *Lock) GetName() string { return l.name }

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
	var err error

	l.pod, err = k8sutil.GetThisPod(l.client)
	if err != nil {
		lockLogger.Error(err, "Error retrieving current pod, could not acquire lock")
		return err
	}

	failTimeout := time.Now().Add(l.GetTimeout())

	var ctx context.Context
	var cancel func()
	if l.GetTimeout() > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), l.GetTimeout())
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}

	defer cancel()

	cm := newConfigMapForLock(l)

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
		nn := types.NamespacedName{Name: cm.GetName(), Namespace: cm.GetNamespace()}
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

	return nil
}

func (l *Lock) checkExistingLockExpiry(ctx context.Context, existingLock *corev1.ConfigMap) error {
	expiresAt, ok := existingLock.Data[expireKey]
	if !ok {
		if l.GetTimeout() > 0 {
			if err := l.releaseLock(ctx, existingLock); err != nil {
				lockLogger.Error(err, "Failed to release stale lock, could not acquire lock")
				return err
			}
		}
		return nil
	}
	if expireTime, err := strconv.ParseInt(expiresAt, 10, 64); err == nil {
		if time.Now().After(time.Unix(expireTime, 0)) {
			if err := l.releaseLock(ctx, existingLock); err != nil {
				lockLogger.Error(err, fmt.Sprintf("Failed to release stale lock, could not acquire lock %s", err.Error()))
				return err
			}
		}
	}
	return nil
}

// Release will delete the configmap, releasing the lock. If the found lock does not
// belong to the running pod, an error is returned.
func (l *Lock) Release() error {
	lockLogger.Info("Releasing lock", "Lock.Name", l.name)
	cm := &corev1.ConfigMap{}
	nn := types.NamespacedName{Name: l.GetName(), Namespace: l.pod.GetNamespace()}
	if err := l.client.Get(context.TODO(), nn, cm); err != nil {
		if !kerrors.IsNotFound(err) {
			lockLogger.Error(err, "Error looking up existing lock, could not release lock")
			return err
		}
		lockLogger.Info("Lock has already been released")
		return nil
	}
	ref := cm.GetOwnerReferences()
	if len(ref) != 1 {
		return fmt.Errorf("Owner references on found lock is malformed: %+v", ref)
	}
	if ref[0].UID != l.pod.GetUID() {
		return fmt.Errorf("Present lock is not owned by this pod, owned by: %s", ref[0].Name)
	}
	return l.releaseLock(context.Background(), cm)
}

// releaseLock removes a lock from kubernetes
func (l *Lock) releaseLock(ctx context.Context, cm *corev1.ConfigMap) error {
	lockLogger.Info("Releasing lock", "Owner", cm.OwnerReferences[0])
	if err := l.client.Delete(ctx, cm); err != nil {
		if !kerrors.IsNotFound(err) {
			lockLogger.Error(err, fmt.Sprintf("Error releasing lock: %s", err.Error()))
			return err
		}
	}
	return nil
}

// newConfigMapForLock returns a new configmap for locking.
func newConfigMapForLock(l *Lock) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      l.GetName(),
			Namespace: l.pod.GetNamespace(),
			Labels:    l.labels,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "v1",
					Kind:               "Pod",
					Name:               l.pod.GetName(),
					UID:                l.pod.GetUID(),
					Controller:         common.BoolPointer(true),
					BlockOwnerDeletion: common.BoolPointer(false),
				},
			},
		},
		Data: l.GetCMData(),
	}
}
