package local

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

func TestNew(t *testing.T) {
	if reflect.TypeOf(New()) != reflect.TypeOf(&LocalAuthProvider{}) {
		t.Error("Someone messed with New")
	}
}

func TestSetup(t *testing.T) {
	cluster := &v1alpha1.VDICluster{}
	cluster.Name = "test-cluster"
	cluster.Namespace = "test-namespace"
	provider := New()
	if err := provider.Setup(cluster); err != nil {
		t.Error("No error should haappen when setting up the local auth provider")
	}
}

func TestAuthenticate(t *testing.T) {
	provider := &LocalAuthProvider{
		getDB: func() (rethinkdb.RethinkDBSession, error) {
			return rethinkdb.NewMock(), nil
		},
		signToken: apiutil.GenerateJWT,
		getKey:    func() ([]byte, error) { return []byte("secret"), nil },
		compHash:  func(string, string) bool { return true },
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(fmt.Sprintf(`
{"username": "%s", "password": "password"}
`, rethinkdb.SuccessItem))))
	provider.Authenticate(rr, req)
	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Error("Expected good status code for valid auth")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(fmt.Sprintf(`
{"username": "%s", "password": "password"}
`, rethinkdb.ErrorItem))))
	provider.Authenticate(rr, req)
	res = rr.Result()
	if res.StatusCode != http.StatusForbidden {
		t.Error("Expected forbidden for unable to lookup user")
	}

	provider.compHash = func(string, string) bool { return false }
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(fmt.Sprintf(`
{"username": "%s", "password": "password"}
`, rethinkdb.SuccessItem))))
	provider.Authenticate(rr, req)
	res = rr.Result()
	if res.StatusCode != http.StatusForbidden {
		t.Error("Expected forbidden status code for bad password")
	}

	provider.getDB = func() (rethinkdb.RethinkDBSession, error) { return nil, errors.New("") }
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(fmt.Sprintf(`
{"username": "%s", "password": "password"}
`, rethinkdb.SuccessItem))))
	provider.Authenticate(rr, req)
	res = rr.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Error("Expected bad request for unable to connect to db")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/login", bytes.NewBuffer([]byte(`
invalid content
`)))
	provider.Authenticate(rr, req)
	res = rr.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Error("Expected bad request for invalid payload")
	}
}
