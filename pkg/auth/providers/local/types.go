package local

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/util/common"
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
