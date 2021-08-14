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

package lock

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	types "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// This would be a good test to have moved to envtest. Since then we could test
// the ownerreferences on the lock.

func getFakeClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	appv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme)
}

func setupLock(t *testing.T, timeoutSeconds int) (*Lock, client.Client) {
	t.Helper()
	os.Setenv("POD_NAME", "test-pod")
	os.Setenv("POD_NAMESPACE", "test-namespace")
	c := getFakeClient(t)
	c.Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
	})
	timeout := time.Duration(timeoutSeconds) * time.Second
	return New(c, "test-lock", timeout), c
}

func TestNew(t *testing.T) {
	l := New(getFakeClient(t), "test-lock", time.Duration(30)*time.Second)
	if l.GetName() != "test-lock" {
		t.Error("Expected name of new lock to be 'test-lock', got:", l.GetName())
	}
	if l.GetTimeout() != time.Duration(30)*time.Second {
		t.Error("Expected new lock to have a 30 second timeout, got:", l.GetTimeout())
	}
}

func TestAcquireLock(t *testing.T) {
	c := getFakeClient(t)
	timeout := time.Duration(30) * time.Second
	lock := New(c, "test-lock", timeout)

	// Test initialization error conditions
	if err := lock.Acquire(); err == nil {
		t.Error("Expected error trying to acquire lock with no environment initialization, got nil")
	} else if !strings.Contains(err.Error(), "POD_NAME") {
		t.Error("Expected error due to no POD_NAME in the environment, got:", err.Error())
	}
	os.Setenv("POD_NAME", "test-pod")
	if err := lock.Acquire(); err == nil {
		t.Error("Expected error trying to acquire lock with no environment initialization, got nil")
	} else if !strings.Contains(err.Error(), "POD_NAMESPACE") {
		t.Error("Expected error due to no POD_NAMESPACE in the environment, got:", err.Error())
	}
	os.Setenv("POD_NAMESPACE", "test-namespace")
	if err := lock.Acquire(); err == nil {
		t.Error("Expected error trying to acquire lock with no environment initialization, got nil")
	} else if !strings.Contains(err.Error(), "not found") {
		t.Error("Expected error due to pod not found in the environment, got:", err.Error())
	}

	// Create a fake pod where this would be running and lock should be acquired
	// successfully.
	c.Create(context.TODO(), &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
	})

	if err := lock.Acquire(); err != nil {
		t.Error("Expected to acquire lock without error, got:", err.Error())
	}

	// A configmap with the name of the lock in the pod namespace should exist
	cmName := types.NamespacedName{Name: "test-lock", Namespace: "test-namespace"}
	if err := c.Get(context.TODO(), cmName, &corev1.ConfigMap{}); err != nil {
		t.Error("Expected a configmap called 'test-lock' in 'test-namespace', got err:", err.Error())
	}

	if err := lock.Release(); err != nil {
		t.Error("Expected to be able to release lock, got:", err.Error())
	}

	// should be safe to call on an already deleted lock without error
	if err := lock.Release(); err != nil {
		t.Error("Expected to be able to release lock, got:", err.Error())
	}

	if err := c.Get(context.TODO(), cmName, &corev1.ConfigMap{}); err == nil {
		t.Error("Expected configmap to not exist anymore, got nil")
	} else if client.IgnoreNotFound(err) != nil {
		t.Error("Expected not found error, got:", err)
	}
}

func TestLockTimeout(t *testing.T) {
	// create a lock with a 3 second timeout
	l, c := setupLock(t, 3)

	var wg sync.WaitGroup
	if err := l.Acquire(); err != nil {
		t.Fatal("Could not acquire lock:", err.Error())
	}

	wg.Add(1)
	// Setup a "separate process" to try to grab the same lock
	go func() {
		defer wg.Done()
		nl := New(c, "test-lock", time.Duration(10)*time.Second)
		start := time.Now()
		if err := nl.Acquire(); err != nil {
			t.Error("Expected to eventually acquire the lock")
		}
		duration := time.Since(start)
		// With an original timeout of 5 seconds this routing should have taken
		// between 2-5 seconds. This is probably a terrible idea in a test.
		if duration.Seconds() < 2 || duration.Seconds() > 5 {
			t.Error("Expected stale lock acquisition to take 2-5 seconds.")
		}

	}()
	wg.Wait()
}

func TestLockNoTimeout(t *testing.T) {
	l, c := setupLock(t, -1)

	if err := l.Acquire(); err != nil {
		t.Fatal(err)
	}
	defer l.Release()

	cm := &corev1.ConfigMap{}
	nn := types.NamespacedName{Name: l.GetName(), Namespace: l.pod.GetNamespace()}
	if err := c.Get(context.TODO(), nn, cm); err != nil {
		t.Fatal(err)
	}
	if len(cm.Data) != 0 {
		t.Error("Expected CM with no data for no-timeout lock, got:", cm.Data)
	}
}

func TestLockWithLabels(t *testing.T) {
	l, c := setupLock(t, -1)

	l = l.WithLabels(map[string]string{"test-key": "test-value"})

	if err := l.Acquire(); err != nil {
		t.Fatal(err)
	}
	defer l.Release()

	cm := &corev1.ConfigMap{}
	nn := types.NamespacedName{Name: l.GetName(), Namespace: l.pod.GetNamespace()}
	if err := c.Get(context.TODO(), nn, cm); err != nil {
		t.Fatal(err)
	}
	if len(cm.GetLabels()) != 1 {
		t.Fatal("Expected CM labels map with one k/v pair, got:", cm.GetLabels())
	}
	if val, ok := cm.GetLabels()["test-key"]; !ok {
		t.Error("Expected 'test-key' in labels, got:", cm.GetLabels())
	} else if val != "test-value" {
		t.Error("Expected value of 'test-key' to be 'test-value', got:", val)
	}
}
