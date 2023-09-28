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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/kvdi/kvdi/pkg/util/apiutil"
	"github.com/kvdi/kvdi/pkg/util/errors"
)

// setAccessToken sets the access token for the http client.
func (c *Client) setAccessToken(token string) { c.accessToken = token }

// getAccessToken retrieves the access token for the http client.
func (c *Client) getAccessToken() string { return c.accessToken }

// getEndpoint returns the full URL to the given API endpoint.
func (c *Client) getEndpoint(ep string) string {
	return fmt.Sprintf("%s/api/%s", strings.TrimSuffix(c.opts.URL, "/"), ep)
}

// getWebsocketEndpoint returns the full URL (token included) for a given websocket endpoint.
func (c *Client) getWebsocketEndpoint(ep string) string {
	u := strings.Replace(c.opts.URL, "http", "ws", 1)
	return fmt.Sprintf("%s/api/%s?token=%s", u, ep, c.getAccessToken())
}

// doWebsocket is a helper function for a generic websocket request flow with the API.
func (c *Client) doWebsocket(endpoint string) (io.ReadWriteCloser, error) {
	dialer := websocket.Dialer{
		TLSClientConfig: c.tlsConfig,
	}
	conn, _, err := dialer.Dial(c.getWebsocketEndpoint(endpoint), nil)
	if err != nil {
		return nil, err
	}
	return apiutil.NewGorillaReadWriter(conn), nil
}

// doRaw retrieves the raw response for the given endpoint and method.
func (c *Client) doRaw(method, endpoint string, req interface{}) (*http.Response, error) {
	var reqBody []byte
	var err error

	if req != nil {
		reqBody, err = json.Marshal(req)
		if err != nil {
			return nil, err
		}
	}

	r, err := http.NewRequest(method, c.getEndpoint(endpoint), bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	r.Header.Add("X-Session-Token", c.getAccessToken())
	r.Header.Add("Content-Type", "application/json")

	return c.httpClient.Do(r)
}

// do is a helper function for a generic request flow with the API.
func (c *Client) do(method, endpoint string, req, resp interface{}, retry ...bool) error {
	rawRes, err := c.doRaw(method, endpoint, req)
	if err != nil {
		return err
	}
	defer rawRes.Body.Close()

	if err := errors.CheckAPIError(rawRes); err != nil {
		return err
	}

	body, err := io.ReadAll(rawRes.Body)
	if err != nil {
		return err
	}

	if rawRes.StatusCode == http.StatusUnauthorized {
		if c.tokenRetry {
			if len(retry) == 0 || retry[0] {
				session, err := c.refreshToken()
				if err != nil {
					return err
				}
				c.setAccessToken(session.Token)
				return c.do(method, endpoint, req, resp, false)
			}
		}
	}

	if resp != nil {
		return json.Unmarshal(body, resp)
	}

	return nil
}
