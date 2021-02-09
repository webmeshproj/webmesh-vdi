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
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/secrets"
)

const testUsername = "admin"
const testGroup = "test-group"
const testHash = "test-hash"

func getTestUser(t *testing.T, name string) *User {
	t.Helper()
	return &User{
		Username:     name,
		Groups:       []string{testGroup},
		PasswordHash: testHash,
	}
}

func providerSetUp(t *testing.T) *AuthProvider {
	t.Helper()
	client := getFakeClient(t)
	cluster := &v1alpha1.VDICluster{}
	cluster.Name = "test-cluster"
	cluster.Spec = v1alpha1.VDIClusterSpec{}
	provider := &AuthProvider{
		client:  client,
		cluster: cluster,
		secrets: secrets.GetSecretEngine(cluster),
	}
	provider.secrets.Setup(client, cluster)
	provider.Setup(client, cluster)
	return provider
}

func TestGetPasswdFile(t *testing.T) {
	provider := providerSetUp(t)

	_, err := provider.getPasswdFile()
	if err == nil {
		t.Error("Expected error because no key exists yet")
	}

	var buf bytes.Buffer
	buf.Write(getTestUser(t, testUsername).Encode())
	if err := provider.updatePasswdFile(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatal("Expceted no error updating passwd file, got", err)
	}

	rdr, err := provider.getPasswdFile()
	if err != nil {
		t.Fatal("Expected no error fetching passwd data, got", err)
	}
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(getTestUser(t, testUsername).Encode()) {
		t.Error("Data was malformed on return, got:", string(data))
	}
}

type deadBuffer struct{}

func (d *deadBuffer) Read([]byte) (int, error) {
	return 0, errors.New("")
}

func TestUpdatePasswdFile(t *testing.T) {
	provider := providerSetUp(t)

	if err := provider.updatePasswdFile(&deadBuffer{}); err == nil {
		t.Error("Expected error reading bad buffer")
	}

	var buf bytes.Buffer

	buf.Write(getTestUser(t, testUsername).Encode())

	if err := provider.updatePasswdFile(bytes.NewReader(buf.Bytes())); err != nil {
		t.Error("Expceted no error updating passwd file, got", err)
	}

	passwdFile, err := provider.getPasswdFile()
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(passwdFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != buf.String() {
		t.Error("Passwd body was malformed on update")
	}

	buf.Write(getTestUser(t, "anotherUser").Encode())
	if err := provider.updatePasswdFile(bytes.NewReader(buf.Bytes())); err != nil {
		t.Error("Expceted no error updating passwd file, got", err)
	}

	passwdFile, err = provider.getPasswdFile()
	if err != nil {
		t.Fatal(err)
	}

	body, err = ioutil.ReadAll(passwdFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != buf.String() {
		t.Error("Passwd body was malformed on update")
	}
	if len(strings.Split(strings.TrimSpace(string(body)), "\n")) != 2 {
		t.Error("There should be 2 lines in the file, got", string(body))
	}
}
