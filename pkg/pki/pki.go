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
	appv1 "github.com/kvdi/kvdi/apis/app/v1"
	"github.com/kvdi/kvdi/pkg/secrets"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Manager provides certificate generation, signing, and storage for
// mTLS communication in a VDICluster.
type Manager struct {
	cluster *appv1.VDICluster
	client  client.Client
	secrets *secrets.SecretEngine
}

// New returns a new PKI manager for the provided VDICluster.
func New(c client.Client, cluster *appv1.VDICluster, s *secrets.SecretEngine) *Manager {
	return &Manager{
		cluster: cluster,
		client:  c,
		secrets: s,
	}
}
