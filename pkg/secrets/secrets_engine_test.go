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

package secrets

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	appv1 "github.com/kvdi/kvdi/apis/app/v1"

	"github.com/kvdi/kvdi/pkg/secrets/providers/k8secret"
	"github.com/kvdi/kvdi/pkg/secrets/providers/vault"
	"github.com/kvdi/kvdi/pkg/util/errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestCluster(t *testing.T) *appv1.VDICluster {
	t.Helper()
	cluster := &appv1.VDICluster{}
	cluster.Name = "test-cluster"
	cluster.Spec = appv1.VDIClusterSpec{
		App: &appv1.AppConfig{
			Replicas: 2,
		},
	}
	return cluster
}

func mustSetupSecretEngine(t *testing.T) *SecretEngine {
	scheme := runtime.NewScheme()
	appv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)
	os.Setenv("POD_NAME", "test-pod")
	os.Setenv("POD_NAMESPACE", "test-namespace")
	c := fake.NewFakeClientWithScheme(scheme)
	p := &corev1.Pod{}
	p.Name = "test-pod"
	p.Namespace = "test-namespace"
	c.Create(context.TODO(), p)
	cluster := newTestCluster(t)
	se := GetSecretEngine(cluster)
	if err := se.Setup(c, cluster); err != nil {
		t.Fatal(err)
	}
	return se
}

func TestGetSecretEngine(t *testing.T) {
	cluster := newTestCluster(t)

	se := GetSecretEngine(cluster)
	if reflect.TypeOf(se.backend) != reflect.TypeOf(k8secret.New()) {
		t.Error("Expected secret engine with k8secret backend, got:", reflect.TypeOf(se.backend))
	}

	cluster.Spec = appv1.VDIClusterSpec{
		Secrets: &appv1.SecretsConfig{
			Vault: &appv1.VaultConfig{
				Address: "fake-vault",
			},
		},
	}

	se = GetSecretEngine(cluster)
	if reflect.TypeOf(se.backend) != reflect.TypeOf(vault.New()) {
		t.Error("Expected secret engine with vault backend, got:", reflect.TypeOf(se.backend))
	}
}

func TestReadAndWriteSecret(t *testing.T) {
	se := mustSetupSecretEngine(t)
	defer func() {
		if err := se.Close(); err != nil {
			t.Error("Expected no error closing secret engine, got:", err)
		}
	}()

	// Write a test secret
	if err := se.WriteSecret("test-secret", []byte("test-value")); err != nil {
		t.Fatal(err)
	}

	// Retrieve secret without cache
	val, err := se.ReadSecret("test-secret", false)
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "test-value" {
		t.Error("Secret value malformed on retrieval, got:", string(val))
	}

	// Retrieve secret with cache
	val, err = se.ReadSecret("test-secret", true)
	if err != nil {
		t.Fatal(err)
	}
	if string(val) != "test-value" {
		t.Error("Secret value malformed on retrieval, got:", string(val))
	}

	if _, err := se.ReadSecret("non-exist", true); err == nil {
		t.Fatal("Expected error reading non-existent secret, got nil")
	} else if !errors.IsSecretNotFoundError(err) {
		t.Error("Expected secret not found error, got:", err)
	}

}

func TestReadAndWriteMap(t *testing.T) {
	se := mustSetupSecretEngine(t)
	defer func() {
		if err := se.Close(); err != nil {
			t.Error("Expected no error closing secret engine, got:", err)
		}
	}()

	// Write a test secret
	if err := se.WriteSecretMap("test-secret", map[string][]byte{
		"test-key": []byte("test-value"),
	}); err != nil {
		t.Fatal(err)
	}

	// Retrieve secret without cache
	valMap, err := se.ReadSecretMap("test-secret", false)
	if err != nil {
		t.Fatal(err)
	}
	if val, ok := valMap["test-key"]; !ok {
		t.Fatal("Retrieved secret did not have 'test-key'")
	} else if string(val) != "test-value" {
		t.Error("Secret value malformed on retrieval, got:", string(val))
	}

	// Retrieve secret with cache
	valMap, err = se.ReadSecretMap("test-secret", true)
	if err != nil {
		t.Fatal(err)
	}
	if val, ok := valMap["test-key"]; !ok {
		t.Fatal("Retrieved secret did not have 'test-key'")
	} else if string(val) != "test-value" {
		t.Error("Secret value malformed on retrieval, got:", string(val))
	}

	if _, err := se.ReadSecretMap("non-exist", true); err == nil {
		t.Fatal("Expected error reading non-existent secret, got nil")
	} else if !errors.IsSecretNotFoundError(err) {
		t.Error("Expected secret not found error, got:", err)
	}
}

func TestAppendSecret(t *testing.T) {
	se := mustSetupSecretEngine(t)
	defer func() {
		if err := se.Close(); err != nil {
			t.Error("Expected no error closing secret engine, got:", err)
		}
	}()

	if err := se.AppendSecret("test-secret", []byte("line-one")); err != nil {
		t.Fatal(err)
	}

	if val, _ := se.ReadSecret("test-secret", true); val == nil {
		t.Fatal("Value came back nil")
	} else if string(val) != "line-one" {
		t.Error("Value maformed after append, got:", string(val))
	}

	if err := se.AppendSecret("test-secret", []byte("line-two")); err != nil {
		t.Fatal(err)
	}

	if val, _ := se.ReadSecret("test-secret", true); val == nil {
		t.Fatal("Value came back nil")
	} else if string(val) != "line-one\nline-two" {
		t.Error("Value maformed after append, got:", string(val))
	}
}

func TestCacheExpiry(t *testing.T) {
	se := mustSetupSecretEngine(t)
	defer func() {
		if err := se.Close(); err != nil {
			t.Error("Expected no error closing secret engine, got:", err)
		}
	}()

	se.cacheTTL = time.Duration(2) * time.Second

	// Write a test secret to populate the cache
	if err := se.WriteSecretMap("test-secret-map", map[string][]byte{
		"test-key": []byte("test-value"),
	}); err != nil {
		t.Fatal(err)
	}

	if val := se.readCacheMap("test-secret-map"); val == nil {
		t.Error("Expected cached item to be returned, got nil")
	}

	time.Sleep(time.Duration(2) * time.Second)

	if val := se.readCacheMap("test-secret-map"); val != nil {
		t.Error("Expected cached item to be expired, got return value")
	}

	// Write a test secret to populate the cache
	if err := se.WriteSecret("test-secret", []byte("test-value")); err != nil {
		t.Fatal(err)
	}

	if val := se.readCache("test-secret"); val == nil {
		t.Error("Expected cached item to be returned, got nil")
	}

	time.Sleep(time.Duration(2) * time.Second)

	if val := se.readCache("test-secret"); val != nil {
		t.Error("Expected cached item to be expired, got return value")
	}

}
