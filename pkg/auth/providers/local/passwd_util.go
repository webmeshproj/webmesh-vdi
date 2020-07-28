package local

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

func addUserToBuffer(file io.Reader, newUser *User) (io.Reader, error) {
	buf := new(bytes.Buffer)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		user, err := ParseUser(text)
		if err != nil {
			continue
		}
		if user.Username == newUser.Username {
			return nil, fmt.Errorf("A user with the name %s already exists", user.Username)
		}
		if _, err := buf.Write(user.Encode()); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	if _, err := buf.Write(newUser.Encode()); err != nil {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func getAllUsersFromBuffer(file io.Reader) ([]*User, error) {
	scanner := bufio.NewScanner(file)

	out := make([]*User, 0)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		if user, err := ParseUser(text); err == nil {
			out = append(out, user)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return out, nil
}

func getUserFromBuffer(file io.Reader, username string) (*User, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		user, err := ParseUser(text)
		if err != nil {
			continue
		}
		if user.Username == username {
			return user, nil
		}
	}

	if err := scanner.Err(); err == io.EOF {
		return nil, errors.NewUserNotFoundError(username)
	} else if err != nil {
		return nil, err
	}

	return nil, errors.NewUserNotFoundError(username)
}

func updateUserInBuffer(file io.Reader, updated *User) (io.Reader, error) {
	buf := new(bytes.Buffer)

	var updatedUser bool
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		user, err := ParseUser(text)
		if err != nil {
			continue
		}
		if user.Username == updated.Username {
			if len(updated.Groups) == 0 {
				updated.Groups = user.Groups
			}
			if updated.PasswordHash == "" {
				updated.PasswordHash = user.PasswordHash
			}
			if _, err := buf.Write(updated.Encode()); err != nil {
				return nil, err
			}
			updatedUser = true
			continue
		}
		if _, err := buf.Write(user.Encode()); err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	if !updatedUser {
		return nil, errors.NewUserNotFoundError(updated.Username)
	}

	return bytes.NewReader(buf.Bytes()), nil
}

func deleteUserInBuffer(file io.Reader, username string) (io.Reader, error) {
	buf := new(bytes.Buffer)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		user, err := ParseUser(text)
		if err != nil {
			continue
		}
		if user.Username != username {
			if _, err := buf.Write(user.Encode()); err != nil {
				return nil, err
			}
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}

	return bytes.NewReader(buf.Bytes()), nil
}
