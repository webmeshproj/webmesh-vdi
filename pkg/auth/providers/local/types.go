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
	"errors"
	"fmt"
	"strings"

	"github.com/kvdi/kvdi/pkg/util/common"
)

// User is a struct implementation of a user as stored in the passwd file.
type User struct {
	Username     string
	Groups       []string
	PasswordHash string
}

// PasswordMatchesHash returns true if the supplied password matches the hash for this
// user.
func (u *User) PasswordMatchesHash(passw string) bool {
	return common.PasswordMatchesHash(passw, u.PasswordHash)
}

// Encode will return the string representation of this user for storage in the secret.
func (u *User) Encode() []byte {
	return []byte(fmt.Sprintf("%s:%s:%s\n", u.Username, strings.Join(u.Groups, ","), u.PasswordHash))
}

// ParseUser will parse a string representation of a user into a User object.
func ParseUser(text string) (*User, error) {
	fields := strings.Split(text, ":")
	if len(fields) < 3 {
		return nil, errors.New("Not enough fields to parse in text")
	}
	user := &User{
		Username:     fields[0],
		Groups:       strings.Split(fields[1], ","),
		PasswordHash: strings.Join(fields[2:], ":"),
	}
	return user, nil
}
