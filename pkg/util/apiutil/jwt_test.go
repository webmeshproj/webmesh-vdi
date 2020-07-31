package apiutil

import (
	"testing"
	"time"

	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
)

var secret = []byte("test-secret")

func TestGenerateJWT(t *testing.T) {
	authResult := &v1.AuthResult{
		User: &v1.VDIUser{
			Name: "test-user",
		},
	}
	claims, token, err := GenerateJWT(secret, authResult, true, time.Duration(30)*time.Second)
	if err != nil {
		t.Fatal("Expected no error generating JWT")
	}
	if claims.User.Name != "test-user" {
		t.Error("Username malformed in claims, got:", claims.User.Name)
	}

	// Validity of token is tested in TestDecodeAndVerifyJWT
	if len(token) == 0 {
		t.Error("Got an empty token back")
	}
}

func mustGenerateJWT(t *testing.T, authorized bool, duration time.Duration) string {
	t.Helper()
	_, token, err := GenerateJWT(secret, &v1.AuthResult{
		User: &v1.VDIUser{
			Name: "test-user",
		},
	}, authorized, duration)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func mustDecodeAndVerifyJWT(t *testing.T, token string) *v1.JWTClaims {
	t.Helper()
	claims, err := DecodeAndVerifyJWT(secret, token)
	if err != nil {
		t.Fatal(err)
	}
	return claims
}

func TestDecodeAndVerifyJWT(t *testing.T) {
	var token string
	var claims *v1.JWTClaims
	var err error

	// valid token test cases

	// authorized token
	token = mustGenerateJWT(t, true, time.Duration(10)*time.Second)
	claims = mustDecodeAndVerifyJWT(t, token)
	if !claims.Authorized {
		t.Error("Expected token to be authorized, got false")
	}
	if claims.User.Name != "test-user" {
		t.Error("Expected username to be 'test-user', got:", claims.User.Name)
	}

	// non-authorized token
	token = mustGenerateJWT(t, false, time.Duration(10)*time.Second)
	claims = mustDecodeAndVerifyJWT(t, token)
	if claims.Authorized {
		t.Error("Expected token to not be authorized, got true")
	}

	// invalid token test cases

	// something not even readable
	_, err = DecodeAndVerifyJWT(secret, "fuckeduptoken")
	if err == nil {
		t.Error("Expected error trying to parse a bad token, got nil")
	}

	// mess up the signature
	token = mustGenerateJWT(t, true, time.Duration(10)*time.Second)
	_, err = DecodeAndVerifyJWT(secret, token[:len(token)-5])
	if err == nil {
		t.Error("Expected error from bad signature, got nil")
	} else if err != errTokenSigInvalidError {
		t.Error("Expected bad signature error, got:", err)
	}

	// expired token
	token = mustGenerateJWT(t, true, time.Duration(1)*time.Second)
	time.Sleep(2 * time.Second)
	_, err = DecodeAndVerifyJWT(secret, token)
	if err == nil {
		t.Error("Expected error from expired token, got nil")
	} else if err != errTokenExpiredError {
		t.Error("Expected token expired error, got:", err)
	}

	// mess up the data
	token = mustGenerateJWT(t, true, time.Duration(10)*time.Second)
	_, err = DecodeAndVerifyJWT(secret, token[3:])
	if err == nil {
		t.Error("Expected error from malformed data, got nil")
	} else if err != errTokenMalformedError {
		t.Error("Expected malformed token error, got:", err)
	}
}
