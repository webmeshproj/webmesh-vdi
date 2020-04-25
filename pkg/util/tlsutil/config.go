package tlsutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	cm "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1alpha3"
	cmmeta "github.com/jetstack/cert-manager/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// serverCertMountPath redeclared locally for mocking
var serverCertMountPath = v1alpha1.ServerCertificateMountPath

// clientCertMountPath redeclared locally for mocking
var clientCertMountPath = v1alpha1.ClientCertificateMountPath

// The minimum TLS version required for all mTLS traffic
var minTLSVersion = uint16(tls.VersionTLS12)

// NewServerTLSConfig returns a new server TLS configuration with client
// certificate verification enabled.
func NewServerTLSConfig() (*tls.Config, error) {
	caCertPool, err := getCACertPool(serverCertMountPath)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		ClientCAs:                caCertPool,
		ClientAuth:               tls.RequireAndVerifyClientCert,
		PreferServerCipherSuites: true,
		MinVersion:               minTLSVersion,
	}
	return tlsConfig, nil
}

// NewServerTLSConfig returns a new client TLS configuration for use with
// connecting to a server requiring mTLS.
func NewClientTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(ClientKeypair())
	if err != nil {
		return nil, err
	}
	caCertPool, err := getCACertPool(clientCertMountPath)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   minTLSVersion,
	}, nil
}

// NewClientTLSConfigFromSecret returns a client TLS config from a kubernetes
// certificate secret.
func NewClientTLSConfigFromSecret(c client.Client, name, namespace string) (*tls.Config, error) {
	nn := types.NamespacedName{Name: name, Namespace: namespace}
	secret := &corev1.Secret{}
	if err := c.Get(context.TODO(), nn, secret); err != nil {
		return nil, err
	}
	for _, key := range []string{cmmeta.TLSCAKey, corev1.TLSCertKey, corev1.TLSPrivateKeyKey} {
		if _, ok := secret.Data[key]; !ok {
			return nil, fmt.Errorf("%s missing from TLS secret", key)
		}
	}
	cert, err := tls.X509KeyPair(secret.Data[corev1.TLSCertKey], secret.Data[corev1.TLSPrivateKeyKey])
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(secret.Data[cmmeta.TLSCAKey])
	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   minTLSVersion,
	}, nil
}

// ServerKeypair returns the path to a server certificatee and key.
func ServerKeypair() (string, string) {
	return filepath.Join(serverCertMountPath, corev1.TLSCertKey),
		filepath.Join(serverCertMountPath, corev1.TLSPrivateKeyKey)
}

// ClientKeypair returns the path to a client certificatee and key.
func ClientKeypair() (string, string) {
	return filepath.Join(clientCertMountPath, corev1.TLSCertKey),
		filepath.Join(clientCertMountPath, corev1.TLSPrivateKeyKey)
}

// getCACertPool creates a CA Certificate pool from the CA found at the given
// mount point.
func getCACertPool(mountPath string) (*x509.CertPool, error) {
	caCertFile := filepath.Join(mountPath, cmmeta.TLSCAKey)
	caCert, err := ioutil.ReadFile(caCertFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool, nil
}

// CAUsages returns the key usages for a CA certificate. It must have every usage
// that will be given to a certificate signed from it.
func CAUsages() []cm.KeyUsage {
	return append(ServerMTLSUsages(), cm.UsageSigning, cm.UsageCertSign, cm.UsageCRLSign, cm.UsageOCSPSigning)
}

// ServerMTLSUsages returns the key usages to use for a server side TLS certificate.
func ServerMTLSUsages() []cm.KeyUsage {
	return append(ClientMTLSUsages(), cm.UsageServerAuth)
}

// ClientMTLSUsages returns the key usages for a client side TLS certificate.
func ClientMTLSUsages() []cm.KeyUsage {
	return []cm.KeyUsage{
		cm.UsageKeyEncipherment,
		cm.UsageDigitalSignature,
		cm.UsageClientAuth,
	}
}
