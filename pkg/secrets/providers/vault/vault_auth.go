package vault

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/vault/api"
)

// DefaultTokenPath is where the k8s serviceaccount token is mounted inside the
// container.
const DefaultTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"

// VaultAuthRequest represents a request for a vault token using the k8s JWT.
// There is probably a struct defined in the libary for this somewhere.
type VaultAuthRequest struct {
	JWT  string `json:"jwt"`
	Role string `json:"role"`
}

// getClientToken will read the k8s serviceaccount token and use it to request
// a vault login token.
func (p *Provider) getClientToken() (*api.Secret, error) {
	tokenBytes, err := ioutil.ReadFile(DefaultTokenPath)
	if err != nil {
		return nil, err
	}
	authURLStr := fmt.Sprintf("%s/v1/auth/kubernetes/login", p.vaultConfig.Address)
	body, err := json.Marshal(&VaultAuthRequest{JWT: string(tokenBytes), Role: p.crConfig.GetAuthRole()})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, authURLStr, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	res, err := p.vaultConfig.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.New(string(resBody))
	}
	authResponse := &api.Secret{}
	return authResponse, json.Unmarshal(resBody, authResponse)
}

// runTokenRefreshLoop waits for 60 seconds before token expiry and either renews
// or requests a new token.
func (p *Provider) runTokenRefreshLoop(authInfo *api.Secret) {
	var err error
	ticker := newAuthTicker(authInfo.Auth)
	for {
		select {
		case <-p.stopCh:
			vaultLogger.Info("Stopping token refresh loop")
			return
		case <-ticker.C:
			vaultLogger.Info("Refreshing client token")
			if authInfo != nil && authInfo.Auth.Renewable {
				authInfo, err = p.client.Auth().Token().RenewSelf(authInfo.Auth.LeaseDuration)
				if err == nil {
					p.client.SetToken(authInfo.Auth.ClientToken)
					ticker = newAuthTicker(authInfo.Auth)
					continue
				}
				vaultLogger.Error(err, "Failed to renew token, requesting a new one")
				// If there was an error we can try a full login
			}
			var err error
			authInfo, err = p.getClientToken()
			if err != nil {
				vaultLogger.Error(err, "Failed to acquire a new vault token, retrying in 10 seconds")
				ticker = time.NewTicker(time.Duration(10) * time.Second)
				continue
			}
			p.client.SetToken(authInfo.Auth.ClientToken)
			ticker = newAuthTicker(authInfo.Auth)
			continue
		}
	}
}

// newAuthTicker returns a ticker for 60 seconds before the expiry of the given
// token information.
func newAuthTicker(auth *api.SecretAuth) *time.Ticker {
	return time.NewTicker(time.Duration(auth.LeaseDuration-60) * time.Second)
}
