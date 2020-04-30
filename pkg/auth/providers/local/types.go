package local

import (
	"errors"
	"fmt"
	"strings"

	"github.com/tinyzimmer/kvdi/pkg/util/common"
)

type LocalUser struct {
	Username     string
	Groups       []string
	PasswordHash string
}

func (u *LocalUser) PasswordMatchesHash(passw string) bool {
	return common.PasswordMatchesHash(passw, u.PasswordHash)
}

func (u *LocalUser) Encode() []byte {
	return []byte(fmt.Sprintf("%s:%s:%s\n", u.Username, strings.Join(u.Groups, ","), u.PasswordHash))
}

func ParseUser(text string) (*LocalUser, error) {
	fields := strings.Split(text, ":")
	if len(fields) < 3 {
		return nil, errors.New("Not enough fields to parse in text")
	}
	user := &LocalUser{
		Username:     fields[0],
		Groups:       strings.Split(fields[1], ","),
		PasswordHash: strings.Join(fields[2:], ":"),
	}
	return user, nil
}
