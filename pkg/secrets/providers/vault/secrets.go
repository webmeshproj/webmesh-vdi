package vault

import (
	"encoding/base64"
	"fmt"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// ReadSecret implements SecretsProvider and will retrieve the requsted secret
// from vault. Since it is assume that all secrets are []byte, when reading the
// secret we have to decode the base64 that vault returns it as.
func (p *Provider) ReadSecret(name string) ([]byte, error) {
	path := p.getSecretPath(name)
	res, err := p.client.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	if res == nil || res.Data == nil {
		vaultLogger.Info("Secret data is nil, assuming doesn't exist", "Path", path)
		return nil, errors.NewSecretNotFoundError(name)
	}
	contents, ok := res.Data["data"]
	if !ok {
		vaultLogger.Info("No 'data' key found in the secret", "Path", path)
		return nil, errors.NewSecretNotFoundError(name)
	}
	out, ok := contents.(string)
	if !ok {
		vaultLogger.Info("Could not assert secret data to string, probably empty", "Path", path)
		return nil, errors.NewSecretNotFoundError(name)
	}
	outBytes, err := base64.StdEncoding.DecodeString(out)
	if err != nil {
		vaultLogger.Info("Could not decode vault base64 data", "Path", path)
		return nil, errors.NewSecretNotFoundError(name)
	}
	return outBytes, nil
}

// WriteSecret implements SecretsProvider and will write the secret to the vault
// backend.
func (p *Provider) WriteSecret(name string, content []byte) error {
	_, err := p.client.Logical().Write(p.getSecretPath(name), map[string]interface{}{
		"data": content,
	})
	return err
}

// getSecretPath returns the path to a given secret name in vault.
func (p *Provider) getSecretPath(name string) string {
	return fmt.Sprintf("%s/%s", p.crConfig.GetSecretsPath(), name)
}
