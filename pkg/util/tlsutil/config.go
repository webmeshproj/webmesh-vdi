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

func NewServerTLSConfig() (*tls.Config, error) {
	caCertPool, err := getCACertPool(v1alpha1.ServerCertificateMountPath)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		ClientCAs:                caCertPool,
		ClientAuth:               tls.RequireAndVerifyClientCert,
		PreferServerCipherSuites: true,
		MinVersion:               tls.VersionTLS12,
	}
	return tlsConfig, nil
}

func NewClientTLSConfig() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(ClientKeypair())
	if err != nil {
		return nil, err
	}
	caCertPool, err := getCACertPool(v1alpha1.ClientCertificateMountPath)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
	}, nil
}

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
	}, nil
}

func ServerKeypair() (string, string) {
	return filepath.Join(v1alpha1.ServerCertificateMountPath, corev1.TLSCertKey),
		filepath.Join(v1alpha1.ServerCertificateMountPath, corev1.TLSPrivateKeyKey)
}

func ClientKeypair() (string, string) {
	return filepath.Join(v1alpha1.ClientCertificateMountPath, corev1.TLSCertKey),
		filepath.Join(v1alpha1.ClientCertificateMountPath, corev1.TLSPrivateKeyKey)
}

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

func CAUsages() []cm.KeyUsage {
	return append(ServerMTLSUsages(), cm.UsageSigning, cm.UsageCertSign, cm.UsageCRLSign, cm.UsageOCSPSigning)
}

func ServerMTLSUsages() []cm.KeyUsage {
	return append(ClientMTLSUsages(), cm.UsageServerAuth)
}

func ClientMTLSUsages() []cm.KeyUsage {
	return []cm.KeyUsage{
		cm.UsageKeyEncipherment,
		cm.UsageDigitalSignature,
		cm.UsageClientAuth,
	}
}
