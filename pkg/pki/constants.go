package pki

import (
	"crypto/x509"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
)

// default keySize of 4096
const keySize = 4096

// Secrets key values redeclared locally.
const (
	privateKeySecretKey  = corev1.TLSPrivateKeyKey
	certificateSecretKey = corev1.TLSCertKey
	caCertSecretKey      = v1.CACertKey
)

// allTLSKeys used to check presence of all keys in a secret
var allTLSKeys = []string{privateKeySecretKey, certificateSecretKey, caCertSecretKey}

// Certificate Usages
var certificateUsages = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
var caUsages = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign

// Extended Certificate Usages
var clientExtUsages = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
var serverExtUsages = append(clientExtUsages, x509.ExtKeyUsageServerAuth)
var caExtUsages = append(serverExtUsages, clientExtUsages...)
