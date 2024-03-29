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

package apiutil

import (
	"testing"
	"time"

	"github.com/kvdi/kvdi/pkg/types"
)

var secret = []byte("test-secret")

func TestGenerateJWT(t *testing.T) {
	authResult := &types.AuthResult{
		User: &types.VDIUser{
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
	_, token, err := GenerateJWT(secret, &types.AuthResult{
		User: &types.VDIUser{
			Name: "test-user",
		},
	}, authorized, duration)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func mustDecodeAndVerifyJWT(t *testing.T, token string) *types.JWTClaims {
	t.Helper()
	claims, err := DecodeAndVerifyJWT(secret, token)
	if err != nil {
		t.Fatal(err)
	}
	return claims
}

func TestDecodeAndVerifyJWT(t *testing.T) {
	var token string
	var claims *types.JWTClaims
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
