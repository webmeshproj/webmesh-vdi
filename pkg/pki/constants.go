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
