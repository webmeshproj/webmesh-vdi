package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// setAccessToken sets the access token for the http client.
func (c *Client) setAccessToken(token string) { c.accessToken = token }

// getAccessToken retrieves the access token for the http client.
func (c *Client) getAccessToken() string { return c.accessToken }

// getEndpoint returns the full URL to the given API endpoint.
func (c *Client) getEndpoint(ep string) string {
	return fmt.Sprintf("%s/api/%s", strings.TrimSuffix(c.opts.URL, "/"), ep)
}

// returnAPIError converts the given response body into an API error and returns it.
// If the body cannot be decoded, an error containing its contents is returned.
func (c *Client) returnAPIError(body []byte) error {
	err := &errors.APIError{}
	if decodeerr := json.Unmarshal(body, err); decodeerr != nil {
		return errors.New(string(body))
	}
	return err
}

// do is a helper function for a generic request flow with the API.
func (c *Client) do(method, endpoint string, req, resp interface{}) error {
	var reqBody []byte
	var err error

	if req != nil {
		reqBody, err = json.Marshal(req)
		if err != nil {
			return err
		}
	}

	r, err := http.NewRequest(method, c.getEndpoint(endpoint), bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	r.Header.Add("X-Session-Token", c.getAccessToken())
	r.Header.Add("Content-Type", "application/json")

	rawRes, err := c.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer rawRes.Body.Close()

	body, err := ioutil.ReadAll(rawRes.Body)
	if err != nil {
		return err
	}
	if rawRes.StatusCode != http.StatusOK {
		return c.returnAPIError(body)
	}

	if resp != nil {
		return json.Unmarshal(body, resp)
	}

	return nil
}
