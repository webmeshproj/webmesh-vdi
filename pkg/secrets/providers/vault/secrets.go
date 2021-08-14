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

package vault

import (
	"encoding/base64"
	"fmt"

	"github.com/kvdi/kvdi/pkg/util/errors"
)

// ReadSecret implements SecretsProvider and will retrieve the requsted secret
// from vault. Since it is assume that all secrets are []byte, when reading the
// secret we have to decode the base64 that vault returns it as.
func (p *Provider) ReadSecret(name string) ([]byte, error) {
	secretMap, err := p.ReadSecretMap(name)
	if err != nil {
		return nil, err
	}
	data, ok := secretMap["data"]
	if !ok {
		return nil, errors.NewSecretNotFoundError(name)
	}
	return data, nil
}

// WriteSecret implements SecretsProvider and will write the secret to the vault
// backend.
func (p *Provider) WriteSecret(name string, content []byte) error {
	if len(content) == 0 {
		return p.WriteSecretMap(name, nil)
	}
	return p.WriteSecretMap(name, map[string][]byte{
		"data": content,
	})
}

// ReadSecretMap returns a map from the vault server.
func (p *Provider) ReadSecretMap(name string) (map[string][]byte, error) {
	path := p.getSecretPath(name)
	res, err := p.client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Data == nil {
		vaultLogger.Info("Secret data is nil, assuming doesn't exist", "Path", path)
		return nil, errors.NewSecretNotFoundError(name)
	}
	out := make(map[string][]byte)
	for k, v := range res.Data {
		data, ok := v.(string)
		if !ok {
			vaultLogger.Info("Could not assert secret data to string, probably empty", "Path", path)
			return nil, errors.NewSecretNotFoundError(name)
		}
		outBytes, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			vaultLogger.Info("Could not decode vault base64 data", "Path", path)
			return nil, errors.NewSecretNotFoundError(name)
		}
		out[k] = outBytes
	}
	return out, nil
}

// WriteSecretMap implements SecretsProvider and will write the key-value pair
// to the secrets backend. The secret can be read back in the same fashion.
// This will be the preferred function going forward.
func (p *Provider) WriteSecretMap(name string, content map[string][]byte) error {
	if len(content) == 0 {
		_, err := p.client.Logical().Delete(p.getSecretPath(name))
		return err
	}
	out := make(map[string]interface{})
	for k, v := range content {
		out[k] = v
	}
	_, err := p.client.Logical().Write(p.getSecretPath(name), out)
	return err
}

// getSecretPath returns the path to a given secret name in vault.
func (p *Provider) getSecretPath(name string) string {
	return fmt.Sprintf("%s/%s", p.crConfig.GetSecretsPath(), name)
}
