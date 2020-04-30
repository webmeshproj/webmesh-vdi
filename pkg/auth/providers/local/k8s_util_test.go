package local

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"

	corev1 "k8s.io/api/core/v1"
)

const testUsername = "admin"
const testGroup = "test-group"
const testHash = "test-hash"

func getTestUser(t *testing.T, name string) *LocalUser {
	t.Helper()
	return &LocalUser{
		Username:     name,
		Groups:       []string{testGroup},
		PasswordHash: testHash,
	}
}

func providerSetUp(t *testing.T) (*LocalAuthProvider, *corev1.Secret) {
	t.Helper()
	client := getFakeClient(t)
	cluster := &v1alpha1.VDICluster{}
	cluster.Name = "test-cluster"
	cluster.Spec = v1alpha1.VDIClusterSpec{}
	provider := &LocalAuthProvider{
		client:  client,
		cluster: cluster,
	}
	secret := &corev1.Secret{}
	secret.Name = cluster.GetAppSecretsName()
	secret.Namespace = cluster.GetCoreNamespace()
	return provider, secret
}

func TestGetSecret(t *testing.T) {

	provider, secret := providerSetUp(t)

	if _, err := provider.getSecret(); err == nil {
		t.Error("Expected error fetching non-exist secret")
	}

	provider.client.Create(context.TODO(), secret)
	if _, err := provider.getSecret(); err != nil {
		t.Error("Expected no error fetching secret, got", err)
	}
}

func TestGetPasswdFile(t *testing.T) {
	provider, secret := providerSetUp(t)

	if _, err := provider.getPasswdFile(); err == nil {
		t.Error("Expected error fetching non-exist secret")
	}

	provider.client.Create(context.TODO(), secret)
	if _, err := provider.getPasswdFile(); err == nil {
		t.Error("Expected error fetching secret with no data")
	}

	secret.Data = map[string][]byte{
		passwdKey: getTestUser(t, testUsername).Encode(),
	}
	provider.client.Update(context.TODO(), secret)

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
	provider, secret := providerSetUp(t)

	if err := provider.updatePasswdFile(&deadBuffer{}); err == nil {
		t.Error("Expected error reading bad buffer")
	}

	var buf bytes.Buffer

	if err := provider.updatePasswdFile(&buf); err == nil {
		t.Error("Expected error fetching non-exist secret")
	}

	provider.client.Create(context.TODO(), secret)

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
