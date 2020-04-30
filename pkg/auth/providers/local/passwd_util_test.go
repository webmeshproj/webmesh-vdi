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
