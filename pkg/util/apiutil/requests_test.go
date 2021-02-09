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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	v1 "github.com/tinyzimmer/kvdi/pkg/apis/meta/v1"
)

func mustNewRequest(t *testing.T, path string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func TestRequestUserSession(t *testing.T) {
	claims := &v1.JWTClaims{User: &v1.VDIUser{Name: "test-user"}}
	req := mustNewRequest(t, "/test")

	SetRequestUserSession(req, claims)

	reqClaims := GetRequestUserSession(req)
	if reqClaims == nil {
		t.Fatal("Claims in request context came back nil")
	}

	// pointers should be identical
	if claims != reqClaims {
		t.Error("Expected same claims ptr to be set and retrieved from request")
	}
}

func TestRequestObject(t *testing.T) {
	req := mustNewRequest(t, "/test")

	obj := &struct {
		val string
	}{val: "test"}

	SetRequestObject(req, obj)

	reqObj := GetRequestObject(req)
	if reqObj == nil {
		t.Fatal("Request object in context came back nil")
	}

	if reqObj != obj {
		t.Error("Expected same obj ptr to be set and retrieved from request")
	}
}

func TestGorillaHelpers(t *testing.T) {
	// Tests are executed inside router methods. Values expected configured below

	r := mux.NewRouter()

	r.PathPrefix("/user/{user}/test").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromRequest(r)
		if user != "helloworld" {
			t.Error("Expected user value to be helloworld, got:", user)
		}
		path := GetGorillaPath(r)
		if path != "/user/{user}/test" {
			t.Error("Gorilla path malformed, got:", path)
		}
	})

	r.PathPrefix("/role/{role}/test").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := GetRoleFromRequest(r)
		if role != "helloworld" {
			t.Error("Expected template value to be helloworld, got:", role)
		}
		path := GetGorillaPath(r)
		if path != "/role/{role}/test" {
			t.Error("Gorilla path malformed, got:", path)
		}
	})

	r.PathPrefix("/template/{template}/test").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tmpl := GetTemplateFromRequest(r)
		if tmpl != "helloworld" {
			t.Error("Expected template value to be helloworld, got:", tmpl)
		}
		path := GetGorillaPath(r)
		if path != "/template/{template}/test" {
			t.Error("Gorilla path malformed, got:", path)
		}
	})

	r.PathPrefix("/nn/{namespace}/{name}/test").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nn := GetNamespacedNameFromRequest(r)
		if nn.Namespace != "hello" {
			t.Error("Expected namespace value in request to be 'hello', got:", nn.Namespace)
		}
		if nn.Name != "world" {
			t.Error("Expected name value in request to be 'world', got:", nn.Name)
		}
		path := GetGorillaPath(r)
		if path != "/nn/{namespace}/{name}/test" {
			t.Error("Gorilla path malformed, got:", path)
		}
	})

	// namespaced name
	req := mustNewRequest(t, "/nn/hello/world/test")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	req = mustNewRequest(t, "/template/helloworld/test")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	req = mustNewRequest(t, "/role/helloworld/test")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	req = mustNewRequest(t, "/user/helloworld/test")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)
}
