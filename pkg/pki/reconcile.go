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

package pki

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	appv1 "github.com/tinyzimmer/kvdi/apis/app/v1"
	desktopsv1 "github.com/tinyzimmer/kvdi/apis/desktops/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
	"github.com/tinyzimmer/kvdi/pkg/util/k8sutil"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reconcile reconciles the base PKI infrastructure for the VDICluster.
func (m *Manager) Reconcile(reqLogger logr.Logger) error {
	caCert, caKey, err := m.reconcileCA(reqLogger)
	if err != nil {
		return err
	}
	if err := m.reconcileAppCertificates(reqLogger, caCert, caKey); err != nil {
		return err
	}
	return nil
}

// ReconcileDesktop reconciles the mTLS server certificate for a desktop instance.
func (m *Manager) ReconcileDesktop(reqLogger logr.Logger, desktop *desktopsv1.Session, serviceIP string) error {
	// reconcile the CA to retrieve it
	caCert, caKey, err := m.reconcileCA(reqLogger)
	if err != nil {
		return err
	}

	nn := types.NamespacedName{
		Name:      desktop.GetName(),
		Namespace: desktop.GetNamespace(),
	}
	secret := &corev1.Secret{}
	if err := m.client.Get(context.TODO(), nn, secret); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		// We need to create the certificate
		desktopCert := newDesktopProxyCertificate(m.cluster, desktop, serviceIP)
		privKey, err := newKey()
		if err != nil {
			return err
		}
		desktopCertBytes, err := x509.CreateCertificate(rand.Reader, desktopCert, caCert, &privKey.PublicKey, caKey)
		if err != nil {
			return err
		}
		certData, err := encodeTLSKeyPair(caCert.Raw, desktopCertBytes, privKey)
		if err != nil {
			return err
		}
		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            nn.Name,
				Namespace:       nn.Namespace,
				Labels:          k8sutil.GetDesktopLabels(m.cluster, desktop),
				Annotations:     desktop.GetAnnotations(),
				OwnerReferences: desktop.OwnerReferences(),
			},
			Data: certData,
		}
		return m.client.Create(context.TODO(), newSecret)
	}

	// TODO: since these are shortlived I can postpone doing verification
	// but it should be done

	return nil
}

// reconcileCA will ensure the presence and validity of a CA certificate and return
// its contents or any error.
func (m *Manager) reconcileCA(reqLogger logr.Logger) (*x509.Certificate, *rsa.PrivateKey, error) {
	caCert, err := m.secrets.ReadSecretMap(m.cluster.GetCAName(), true)
	if err != nil {
		if !errors.IsSecretNotFoundError(err) {
			return nil, nil, err
		}
		reqLogger.Info("Generating new CA for the kVDI cluster")
		ca := newCACertificate(m.cluster)
		caPrivKey, err := newKey()
		if err != nil {
			return nil, nil, err
		}
		caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
		if err != nil {
			return nil, nil, err
		}
		caCert, err = encodeTLSKeyPair(caBytes, caBytes, caPrivKey)
		if err != nil {
			return nil, nil, err
		}
		if err := m.secrets.WriteSecretMap(m.cluster.GetCAName(), caCert); err != nil {
			return nil, nil, err
		}
		cert, err := x509.ParseCertificate(caBytes)
		return cert, caPrivKey, err
	}

	// Verify the existing CA

	// run this function if any error occurs during parsing or verification
	recreateFunc := func(msg string) error {
		reqLogger.Info("We have lost our CA. Will need to re-create the entire PKI", "Error", msg)
		if err := m.secrets.WriteSecretMap(m.cluster.GetCAName(), nil); err != nil {
			return err
		}
		return errors.NewRequeueError("Pre-existing CA was corrupted, recreating PKI", 1)
	}

	// make sure all keys are present
	for _, key := range allTLSKeys {
		if _, ok := caCert[key]; !ok {
			return nil, nil, recreateFunc("Key is missing from ca secret: " + key)
		}
	}

	// decode required values
	privKeyBlock, _ := pem.Decode(caCert[privateKeySecretKey])
	if privKeyBlock == nil {
		return nil, nil, recreateFunc("Could not decode CA private key")
	}
	certBlock, _ := pem.Decode(caCert[certificateSecretKey])
	if certBlock == nil {
		return nil, nil, recreateFunc("Could not decode CA certificate")
	}

	// parse the key and certificate
	privKey, err := x509.ParsePKCS1PrivateKey(privKeyBlock.Bytes)
	if err != nil {
		return nil, nil, recreateFunc("Could not parse CA private key: " + err.Error())
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, nil, recreateFunc("Could not parse CA certificate: " + err.Error())
	}

	// TODO: Check the key
	return cert, privKey, nil
}

// reconcileAppCertificates reconciles certificates for the app pods.
func (m *Manager) reconcileAppCertificates(reqLogger logr.Logger, caCert *x509.Certificate, caPrivKey *rsa.PrivateKey) error {
	// a list of objects containing the namespaced name and cert
	// create function for a server and client certificate.
	//
	// the server certificate is used by the frontend to serve HTTPS
	// and should be adapted to allow a custom cert.
	//
	// the client certificate is for creating mTLS sessions with
	// desktop pods.
	appCertificates := []struct {
		namespacedName types.NamespacedName
		createCertFunc func(*appv1.VDICluster) *x509.Certificate
	}{
		{
			namespacedName: m.cluster.GetAppServerTLSNamespacedName(),
			createCertFunc: newAppServerCertificate,
		},
		{
			namespacedName: m.cluster.GetAppClientTLSNamespacedName(),
			createCertFunc: newAppClientCertificate,
		},
	}

	for _, appCertificate := range appCertificates {
		if m.cluster.AppIsUsingExternalServerTLS() && appCertificate.namespacedName.Name == m.cluster.GetAppServerTLSSecretName() {
			// skip app server certificate if using a user-supplied certificate
			continue
		}
		secret := &corev1.Secret{}
		if err := m.client.Get(context.TODO(), appCertificate.namespacedName, secret); err != nil {
			if client.IgnoreNotFound(err) != nil {
				return err
			}
			reqLogger.Info("Generating new app certificate/key-pair", "Certificate", appCertificate.namespacedName)
			// create a new keypair
			key, err := newKey()
			if err != nil {
				return err
			}
			// create a new signed certificate
			cert := appCertificate.createCertFunc(m.cluster)
			certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &key.PublicKey, caPrivKey)
			if err != nil {
				return err
			}
			// encode to PEM and save to k8s
			certData, err := encodeTLSKeyPair(caCert.Raw, certBytes, key)
			if err != nil {
				return err
			}
			newSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:            appCertificate.namespacedName.Name,
					Namespace:       appCertificate.namespacedName.Namespace,
					Labels:          m.cluster.GetComponentLabels("app"),
					Annotations:     m.cluster.GetAnnotations(),
					OwnerReferences: m.cluster.OwnerReferences(),
				},
				Data: certData,
			}
			if err := m.client.Create(context.TODO(), newSecret); err != nil {
				return err
			}
			continue
		}
		// we have a certificate, verify it
		certData := secret.Data
		recreateFunc := func(msg string) error {
			reqLogger.Info("Secret data is corrupted, deleting and requeueing", "Certificate", appCertificate.namespacedName, "Error", msg)
			if err := m.client.Delete(context.TODO(), secret); err != nil {
				return err
			}
			return errors.NewRequeueError(fmt.Sprintf("Need to recreate app certificate: %s", msg), 3)
		}
		if certData == nil {
			return recreateFunc("Secret data is nil")
		}

		// verify that the ca provided to the function matches the one in the secret
		existingCAPEM, ok := certData[caCertSecretKey]
		if !ok {
			return recreateFunc("No CA in secret data")
		}
		existingCABlock, _ := pem.Decode(existingCAPEM)
		if existingCABlock == nil {
			return recreateFunc("Could not PEM decode CA data")
		}
		existingCA, err := x509.ParseCertificate(existingCABlock.Bytes)
		if err != nil {
			return recreateFunc("Failed to parse PEM decoded data to certificate")
		}
		if !existingCA.Equal(caCert) {
			return recreateFunc("Provided CA certificate doesn't match that in the secret")
		}
		// verify the cert
		existingCert, ok := secret.Data[certificateSecretKey]
		if !ok {
			return recreateFunc("No certificate in secret data")
		}
		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(existingCAPEM); !ok {
			return recreateFunc("Failed to create cert pool from CA")
		}
		block, _ := pem.Decode([]byte(existingCert))
		if block == nil {
			return recreateFunc("Failed to parse certificate PEM")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return recreateFunc("Failed to parse certificate: " + err.Error())
		}
		opts := x509.VerifyOptions{
			DNSName: m.cluster.GetAppName(),
			Roots:   roots,
			// caExtUsages includes all
			KeyUsages: caExtUsages,
		}
		if _, err := cert.Verify(opts); err != nil {
			return recreateFunc("Failed to verify certificate: " + err.Error())
		}
		// TODO: check the key
	}

	return nil
}
