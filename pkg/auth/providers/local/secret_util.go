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
	"bytes"
	"io"
)

func (a *AuthProvider) getPasswdFile() (io.ReadWriter, error) {
	data, err := a.secrets.ReadSecret(passwdKey, false)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(data), nil
}

func (a *AuthProvider) updatePasswdFile(rdr io.Reader) error {
	body, err := io.ReadAll(rdr)
	if err != nil {
		return err
	}
	return a.secrets.WriteSecret(passwdKey, body)
}
