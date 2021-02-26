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

// Package client provides a REST wrapper to the kVDI API.
package client

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
)

// Client provides a REST wrapper to the kVDI API.
type Client struct {
	// the options used to configure the client
	opts *Opts
	// the http client used for sending requests
	httpClient *http.Client
	// the current access token for performing requests
	accessToken string
	// whether to retry unauthorized requests with a refresh token.
	// only works for long-lived client usages.
	tokenRetry bool
}

// Opts are options to pass to New when creating a new client interface.
type Opts struct {
	// The full URL to the kVDI app server (e.g. https://kvdi.local)
	URL string
	// The username to use to authenticate.
	Username string
	// The password to use to authenticate.
	Password string
	// TODO: Allow for API keys tied to roles for auth providers that don't allow us
	// to independently verify credentials (e.g. OpenID).
	APIKey string
	// The PEM encoded CA certificate to use when validating the kVDI server certificate.
	// When using the generated certificate, this can be found in the kvdi-app
	// server TLS secret.
	TLSCACert []byte
	// Set to true to skip TLS verification.
	TLSInsecureSkipVerify bool
}

// New creates a new kVDI client.
func New(opts *Opts) (*Client, error) {
	cl := &Client{opts: opts, tokenRetry: true}

	// refresh tokens are supplied as httponly cookies
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	// initialize an http client with the cookie jar
	cl.httpClient = &http.Client{Jar: jar}

	if strings.HasPrefix(cl.opts.URL, "https") {
		// configure tls
		tlsConfig := &tls.Config{
			InsecureSkipVerify: cl.opts.TLSInsecureSkipVerify,
		}
		if len(cl.opts.TLSCACert) != 0 {
			certPool := x509.NewCertPool()
			certPool.AppendCertsFromPEM(cl.opts.TLSCACert)
			tlsConfig.RootCAs = certPool
		}
		cl.httpClient.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	return cl, cl.authenticate()
}

// SetAutoRefreshToken will set whether the client should try to auto refresh tokens after
// an unauthorized response. The default value is true. Use this method to disable the behavior.
func (c *Client) SetAutoRefreshToken(val bool) { c.tokenRetry = val }

// Close will stop the token refresh goroutine if it's running.
func (c *Client) Close() {
	if err := c.do(http.MethodPost, "logout", nil, nil, false); err != nil {
		log.Println("Error posting to /api/logout. Refresh token could not be revoked:", err)
	}
}
