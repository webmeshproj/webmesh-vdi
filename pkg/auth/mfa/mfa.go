package mfa

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
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

// GetMFAUsers will return a string slice of all MFA user names.
func (m *Manager) GetMFAUsers() ([]string, error) {
	users, err := m.secrets.ReadSecret(v1alpha1.OTPUsersSecretKey, false)
	if err != nil {
		if errors.IsSecretNotFoundError(err) {
			return []string{}, nil
		}
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(users))
	mfaUsers := make([]string, 0)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		spl := strings.Split(text, ":")
		if len(spl) < 2 {
			continue
		}
		mfaUsers = append(mfaUsers, spl[0])
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return mfaUsers, nil
}

// GetUserSecret will retrieve the OTP secret for the given user. If there is
// no secret for this user, a UserNotFound error is returned.
func (m *Manager) GetUserSecret(name string) (string, error) {
	users, err := m.secrets.ReadSecret(v1alpha1.OTPUsersSecretKey, false)
	if err != nil {
		if errors.IsSecretNotFoundError(err) {
			return "", errors.NewUserNotFoundError(name)
		}
		return "", err
	}

	return m.getUserSecretFromReader(name, bytes.NewReader(users))
}

// SetUserSecret sets the value of the user's OTP secret.
func (m *Manager) SetUserSecret(name, secret string) error {
	if err := m.secrets.Lock(); err != nil {
		return err
	}
	defer m.secrets.Release()
	users, err := m.secrets.ReadSecret(v1alpha1.OTPUsersSecretKey, false)
	if err != nil && !errors.IsSecretNotFoundError(err) {
		return err
	} else if errors.IsSecretNotFoundError(err) {
		users = make([]byte, 0)
	}
	newData, err := m.updateUserSecretInReader(name, secret, bytes.NewReader(users))
	if err != nil {
		return err
	}
	return m.secrets.WriteSecret(v1alpha1.OTPUsersSecretKey, newData)
}

// DeleteUserSecret will remove OTP data for the given username.
func (m *Manager) DeleteUserSecret(name string) error {
	if err := m.secrets.Lock(); err != nil {
		return err
	}
	defer m.secrets.Release()
	users, err := m.secrets.ReadSecret(v1alpha1.OTPUsersSecretKey, false)
	if err != nil && !errors.IsSecretNotFoundError(err) {
		return err
	} else if errors.IsSecretNotFoundError(err) {
		return nil
	}
	newData, err := m.deleteUserFromReader(name, bytes.NewReader(users))
	if err != nil {
		return err
	}
	return m.secrets.WriteSecret(v1alpha1.OTPUsersSecretKey, newData)
}

// getUserSecretFromReader will scan a given Reader interface for the provided
// username and return the OTP secret if found, or a UserNotFound error if
// the end of the data is reached.
func (m *Manager) getUserSecretFromReader(name string, rdr io.Reader) (string, error) {
	scanner := bufio.NewScanner(rdr)

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, name) {
			fields := strings.Split(text, ":")
			if len(fields) < 2 {
				return "", errors.New("User OTP data is malformed")
			}
			return strings.Join(fields[1:], ":"), nil
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return "", err
	}

	return "", errors.NewUserNotFoundError(name)
}

// updateUserSecretInReader will iterate the given reader, replacing the user
// data with the new value and producing a new secret for all users. If the user
// is not found in the secret, it's appended to the end.
func (m *Manager) updateUserSecretInReader(name, secret string, rdr io.Reader) ([]byte, error) {
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
			if _, err := newData.WriteString(fmt.Sprintf("%s:%s\n", name, secret)); err != nil {
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
		if _, err := newData.WriteString(fmt.Sprintf("%s:%s\n", name, secret)); err != nil {
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
		// Only write to the new buffer if usernamee does not match
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
