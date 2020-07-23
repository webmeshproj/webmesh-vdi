package tlsutil

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func writeTLSCerts(t *testing.T) (string, func(), error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", nil, err
	}
	clean := func() { os.RemoveAll(dir) }
	if err := ioutil.WriteFile(filepath.Join(dir, corev1.TLSCertKey), testCert, 0644); err != nil {
		clean()
		return "", nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(dir, corev1.TLSPrivateKeyKey), testKey, 0644); err != nil {
		clean()
		return "", nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "ca.crt"), testCA, 0644); err != nil {
		clean()
		return "", nil, err
	}
	return dir, clean, nil
}

func getFakeClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	return fake.NewFakeClientWithScheme(scheme)
}

func TestNewServerTLSConfig(t *testing.T) {
	var err error
	var clean func()
	// overwrite server cert dir
	serverCertMountPath, clean, err = writeTLSCerts(t)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewServerTLSConfig()
	if err != nil {
		t.Error("Expected no error, got:", err)
	}
	if config.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Error("Expected RequireAndVerifyClientCert in TLS config, got:", config.ClientAuth)
	}
	if config.MinVersion != tls.VersionTLS12 {
		t.Error("Expected Minimum VersionTLS12 in TLS config, got:", config.MinVersion)
	}

	clean()
	if _, err := NewServerTLSConfig(); err == nil {
		t.Error("Expected error for missing certs")
	}
}

func TestNewClientTLSConfig(t *testing.T) {
	var err error
	var clean func()
	// overwrite server cert dir
	clientCertMountPath, clean, err = writeTLSCerts(t)
	if err != nil {
		t.Fatal(err)
	}
	config, err := NewClientTLSConfig()
	if err != nil {
		t.Error("Expected no error, got:", err)
	}
	if config.MinVersion != tls.VersionTLS12 {
		t.Error("Expected Minimum VersionTLS12 in TLS config, got:", config.MinVersion)
	}

	// cause the CA error
	os.Remove(filepath.Join(clientCertMountPath, "ca.crt"))
	if _, err := NewClientTLSConfig(); err == nil {
		t.Error("Expected error for missing ca certs")
	}

	// clear the rest
	clean()
	if _, err := NewClientTLSConfig(); err == nil {
		t.Error("Expected error for missing client certs")
	}
}

func TestNewClientTLSConfigFromSecret(t *testing.T) {
	c := getFakeClient(t)
	secret := &corev1.Secret{}
	secret.Name = "test-secret"
	secret.Namespace = "test-namespace"
	secret.Data = map[string][]byte{
		"ca.crt":                testCA,
		corev1.TLSCertKey:       testCert,
		corev1.TLSPrivateKeyKey: testKey,
	}
	c.Create(context.TODO(), secret)

	config, err := NewClientTLSConfigFromSecret(c, "test-secret", "test-namespace")
	if err != nil {
		t.Error("Expected no error for valid secret, got:", err)
	}
	if config.MinVersion != tls.VersionTLS12 {
		t.Error("Expected Minimum VersionTLS12 in TLS config, got:", config.MinVersion)
	}

	secret.Data[corev1.TLSCertKey] = []byte("invalid")
	c.Update(context.TODO(), secret)
	if _, err := NewClientTLSConfigFromSecret(c, "test-secret", "test-namespace"); err == nil {
		t.Error("Expected error for invalid cert, got nil")
	}

	delete(secret.Data, "ca.crt")
	c.Update(context.TODO(), secret)
	if _, err := NewClientTLSConfigFromSecret(c, "test-secret", "test-namespace"); err == nil {
		t.Error("Expected error for missing secret key, got nil")
	}

	if _, err := NewClientTLSConfigFromSecret(c, "fake-secret", "test-namespace"); err == nil {
		t.Error("Expected error for non-existing secret, got nil")
	}
}

func TestServerKeypair(t *testing.T) {
	if cert, key := ServerKeypair(); cert != filepath.Join(serverCertMountPath, corev1.TLSCertKey) {
		t.Error("Got wrong cert path for server keypair:", cert)
	} else if key != filepath.Join(serverCertMountPath, corev1.TLSPrivateKeyKey) {
		t.Error("Got wrong key path for server keypair:", cert)
	}
}

func TestClientKeypair(t *testing.T) {
	if cert, key := ClientKeypair(); cert != filepath.Join(clientCertMountPath, corev1.TLSCertKey) {
		t.Error("Got wrong cert path for client keypair:", cert)
	} else if key != filepath.Join(clientCertMountPath, corev1.TLSPrivateKeyKey) {
		t.Error("Got wrong key path for client keypair:", cert)
	}
}
