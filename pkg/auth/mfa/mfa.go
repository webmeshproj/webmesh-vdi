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

package mfa

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

// Manager is an object for tracking users and their OTP secrets. It uses
// the configured secrets backend for storage.
type Manager struct {
	secrets *secrets.SecretEngine
}

// NewManager returns a new MFA manager with the given secrets engine.
func NewManager(secrets *secrets.SecretEngine) *Manager {
	return &Manager{secrets: secrets}
}

// GetMFAUsers will return a map of all MFA user names and whether they
// have been verified.
func (m *Manager) GetMFAUsers() (map[string]bool, error) {
	users, err := m.secrets.ReadSecret(v1.OTPUsersSecretKey, false)
	if err != nil {
		if errors.IsSecretNotFoundError(err) {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(users))
	mfaUsers := make(map[string]bool)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		spl := strings.Split(text, ":")
		if len(spl) < 3 {
			continue
		}
		mfaUsers[spl[0]] = parseBool(spl[2])
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return mfaUsers, nil
}

// GetUserMFAStatus will retrieve the OTP secret for the given user, and
// whether it has been verified. If there is no secret for this user, a
// UserNotFound error is returned.
func (m *Manager) GetUserMFAStatus(name string) (string, bool, error) {
	users, err := m.secrets.ReadSecret(v1.OTPUsersSecretKey, false)
	if err != nil {
		if errors.IsSecretNotFoundError(err) {
			return "", false, errors.NewUserNotFoundError(name)
		}
		return "", false, err
	}

	return m.getUserStatusFromReader(name, bytes.NewReader(users))
}

// SetUserMFAStatus sets the value of the user's OTP secret and whether it
// is verified.
func (m *Manager) SetUserMFAStatus(name, secret string, verified bool) error {
	if err := m.secrets.Lock(15); err != nil {
		return err
	}
	defer m.secrets.Release()
	users, err := m.secrets.ReadSecret(v1.OTPUsersSecretKey, false)
	if err != nil && !errors.IsSecretNotFoundError(err) {
		return err
	} else if errors.IsSecretNotFoundError(err) {
		users = make([]byte, 0)
	}
	newData, err := m.updateUserStatusInReader(name, secret, verified, bytes.NewReader(users))
	if err != nil {
		return err
	}
	return m.secrets.WriteSecret(v1.OTPUsersSecretKey, newData)
}

// DeleteUserSecret will remove OTP data for the given username.
func (m *Manager) DeleteUserSecret(name string) error {
	if err := m.secrets.Lock(15); err != nil {
		return err
	}
	defer m.secrets.Release()
	users, err := m.secrets.ReadSecret(v1.OTPUsersSecretKey, false)
	if err != nil && !errors.IsSecretNotFoundError(err) {
		return err
	} else if errors.IsSecretNotFoundError(err) {
		return nil
	}
	newData, err := m.deleteUserFromReader(name, bytes.NewReader(users))
	if err != nil {
		return err
	}
	return m.secrets.WriteSecret(v1.OTPUsersSecretKey, newData)
}

// getUserStatusFromReader will scan a given Reader interface for the provided
// username and return the OTP secret and verification status if found, or a
// UserNotFound error if the end of the data is reached first.
func (m *Manager) getUserStatusFromReader(name string, rdr io.Reader) (string, bool, error) {
	scanner := bufio.NewScanner(rdr)

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, name) {
			fields := strings.Split(text, ":")
			if len(fields) < 3 {
				return "", false, errors.New("User OTP data is malformed")
			}
			return fields[1], parseBool(fields[2]), nil
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return "", false, err
	}

	return "", false, errors.NewUserNotFoundError(name)
}

// updateUserStatusInReader will iterate the given reader, replacing the user
// data with the new values and producing a new secret for all users. If the user
// is not found in the secret, it's appended to the end.
func (m *Manager) updateUserStatusInReader(name, secret string, verified bool, rdr io.Reader) ([]byte, error) {
	scanner := bufio.NewScanner(rdr)
	var newData bytes.Buffer
	var updated bool

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		// If it's the same user, write the new secret to the buffer
		if strings.HasPrefix(text, name) {
			if _, err := newData.WriteString(fmt.Sprintf("%s:%s:%t\n", name, secret, verified)); err != nil {
				return nil, err
			}
			updated = true
			continue
		}
		// Copy the text to the new buffer
		if _, err := newData.WriteString(text + "\n"); err != nil {
			return nil, err
		}
	}

	// check if the scanner errored out
	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	// If we didn't update anything, append the user info now
	if !updated {
		if _, err := newData.WriteString(fmt.Sprintf("%s:%s:%t\n", name, secret, verified)); err != nil {
			return nil, err
		}
	}

	return newData.Bytes(), nil
}

// deleteUserFromReader will iterate the given reader, writing all data to a new
// buffer unless the given user matches, in which case the line is skipped.
func (m *Manager) deleteUserFromReader(name string, rdr io.Reader) ([]byte, error) {
	scanner := bufio.NewScanner(rdr)
	var newData bytes.Buffer

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		// Only write to the new buffer if username does not match
		if !strings.HasPrefix(text, name) {
			if _, err := newData.WriteString(text + "\n"); err != nil {
				return nil, err
			}
		}
	}

	// check if the scanner errored out
	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return newData.Bytes(), nil
}

func parseBool(bs string) bool {
	b, err := strconv.ParseBool(bs)
	if err != nil {
		// assume verified is false so when the next update happens
		// we can fix the bad value
		return false
	}
	return b
}
