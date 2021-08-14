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

package local

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/kvdi/kvdi/pkg/util/errors"
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
