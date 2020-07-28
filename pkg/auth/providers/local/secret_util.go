package local

import (
	"bytes"
	"io"
	"io/ioutil"
)

func (a *AuthProvider) getPasswdFile() (io.ReadWriter, error) {
	data, err := a.secrets.ReadSecret(passwdKey, false)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(data), nil
}

func (a *AuthProvider) updatePasswdFile(rdr io.Reader) error {
	body, err := ioutil.ReadAll(rdr)
	if err != nil {
		return err
	}
	return a.secrets.WriteSecret(passwdKey, body)
}
