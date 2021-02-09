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
	"io/ioutil"
	"strings"
	"testing"
)

func TestAddUserToBuffer(t *testing.T) {
	var buf bytes.Buffer
	// add some bad data to the buffer to simulate check conditions
	buf.Write([]byte("\n"))
	buf.Write([]byte("# maybe a comment\n"))
	buf.Write(getTestUser(t, "user1").Encode())
	newUser := getTestUser(t, "user2")
	newBuf, err := addUserToBuffer(bytes.NewReader(buf.Bytes()), newUser)
	if err != nil {
		t.Fatal("Expected no error adding user to buffer, got", err)
	}

	body, err := ioutil.ReadAll(newBuf)
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.Split(strings.TrimSpace(string(body)), "\n")) != 2 {
		t.Error("Expected new buffer with two lines of users, got", string(body))
	}

	if _, err = addUserToBuffer(bytes.NewReader(body), newUser); err == nil {
		t.Error("Expected error for user already existing")
	}
}

func TestGetUserFromBuffer(t *testing.T)  {}
func TestUpdateUserInBuffer(t *testing.T) {}
func TestDeleteUserInBuffer(t *testing.T) {}
