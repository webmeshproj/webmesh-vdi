package client

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// authenticate retrieves an access token for the API and starts a goroutine
// to refresh the token as needed.
func (c *Client) authenticate() error {
	loginRequest := &v1.LoginRequest{
		Username: c.opts.Username,
		Password: c.opts.Password,
		State:    uuid.New().String(),
	}
	payload, err := json.Marshal(loginRequest)
	if err != nil {
		return err
	}
	res, err := c.httpClient.Post(c.getEndpoint("login"), "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return c.returnAPIError(body)
	}

	sessionResponse := &v1.SessionResponse{}

	if err := json.Unmarshal(body, sessionResponse); err != nil {
		return err
	}

	if sessionResponse.State != loginRequest.State {
		return errors.New("State was malformed during authentication flow, your request might have been intercepted")
	}

	c.setAccessToken(sessionResponse.Token)

	if sessionResponse.Renewable {
		c.stopCh = make(chan struct{})
		go c.runTokenRefreshLoop(sessionResponse)
	}

	return nil
}

// runTokenRefreshLoop is used as a goroutine to request a new access token when the
// current one is about to expire.
func (c *Client) runTokenRefreshLoop(session *v1.SessionResponse) {
	runIn := session.ExpiresAt - time.Now().Unix() - 10
	ticker := time.NewTicker(time.Duration(runIn) * time.Second)
	var err error

	for {
		select {
		case <-ticker.C:
			session, err = c.refreshToken()
			if err != nil {
				log.Println("Error refreshing client token, retrying in 2 seconds")
				ticker = time.NewTicker(time.Duration(2 * time.Second))
				continue
			}
			c.setAccessToken(session.Token)
			runIn = session.ExpiresAt - time.Now().Unix() - 10
			ticker = time.NewTicker(time.Duration(runIn) * time.Second)
		case <-c.stopCh:
			return
		}
	}
}

// refreshToken performs a refresh_token request and returns the response or any error.
func (c *Client) refreshToken() (*v1.SessionResponse, error) {
	res, err := c.httpClient.Get(c.getEndpoint("refresh_token"))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, c.returnAPIError(body)
	}

	sessionResponse := &v1.SessionResponse{}
	return sessionResponse, json.Unmarshal(body, sessionResponse)
}
