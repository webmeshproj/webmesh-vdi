package pki

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	mrand "math/rand"
	"net"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"
)

var ouName = []string{"kVDI"}

func newKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, keySize)
}

func newCACertificate(cluster *v1alpha1.VDICluster) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   cluster.GetCAName(),
			Organization: ouName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           caExtUsages,
		KeyUsage:              caUsages,
		BasicConstraintsValid: true,
		DNSNames:              []string{cluster.GetCAName()},
	}
}

func newAppServerCertificate(cluster *v1alpha1.VDICluster) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(int64(mrand.Intn(9999))),
		Subject: pkix.Name{
			CommonName:   cluster.GetAppName(),
			Organization: ouName,
		},
		DNSNames:     tlsutil.DNSNames(cluster.GetAppName(), cluster.GetCoreNamespace()),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		KeyUsage:     certificateUsages,
		ExtKeyUsage:  serverExtUsages,
	}
}

func newAppClientCertificate(cluster *v1alpha1.VDICluster) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject: pkix.Name{
			CommonName:   cluster.GetAppName(),
			Organization: ouName,
		},
		DNSNames:     tlsutil.DNSNames(cluster.GetAppName(), cluster.GetCoreNamespace()),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		KeyUsage:     certificateUsages,
		ExtKeyUsage:  clientExtUsages,
	}
}

func newDesktopProxyCertificate(cluster *v1alpha1.VDICluster, desktop *v1alpha1.Desktop, serviceIP string) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(4),
		Subject: pkix.Name{
			CommonName:   serviceIP,
			Organization: ouName,
		},
		IPAddresses:  []net.IP{net.ParseIP(serviceIP)},
		DNSNames:     append(tlsutil.DNSNames(desktop.GetName(), desktop.GetNamespace()), serviceIP),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		KeyUsage:     certificateUsages,
		ExtKeyUsage:  serverExtUsages,
	}
}

// encodeTLSKeyPair returns a map of PEM encoded values for the provided TLS key pair.
// The `ca` and `cert` are the raw asn1 data of the certificates.
func encodeTLSKeyPair(ca, cert []byte, key *rsa.PrivateKey) (certData map[string][]byte, err error) {
	caPEM := new(bytes.Buffer)
	if err := pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: ca,
	}); err != nil {
		return nil, err
	}
	certPEM := new(bytes.Buffer)
	if err := pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}); err != nil {
		return nil, err
	}
	caPrivKeyPEM := new(bytes.Buffer)
	if err := pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return nil, err
	}
	return map[string][]byte{
		caCertSecretKey:      caPEM.Bytes(),
		certificateSecretKey: certPEM.Bytes(),
		privateKeySecretKey:  caPrivKeyPEM.Bytes(),
	}, nil
}
