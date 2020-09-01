package apiutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func readResponseBody(res *http.Response) ([]byte, error) {
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}

// TestWriteOrLogError tests the generic response writer
func TestWriteOrLogError(t *testing.T) {
	content := []byte("fake response")
	w := httptest.NewRecorder()
	WriteOrLogError(content, w, http.StatusOK)
	res := w.Result()
	if res.Header.Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type application/json")
	}
	if body, err := readResponseBody(res); err != nil {
		t.Fatal("Could not read recorder body")
	} else if strings.TrimSpace(string(body)) != string(content) {
		t.Error("Body was malformed during write:", string(body))
	}
}

// TestReturnAPIError tests returning API errors
func TestAPIErrors(t *testing.T) {
	tests := []struct {
		Func     func(error, http.ResponseWriter)
		Expected int
	}{
		{
			Func:     ReturnAPIError,
			Expected: http.StatusBadRequest,
		},
		{
			Func:     ReturnAPINotFound,
			Expected: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()
		test.Func(errors.New("fake error"), rr)
		if res := rr.Result(); res.StatusCode != test.Expected {
			t.Error("Expected:", test.Expected, "Got:", res.StatusCode)
		}
	}

	rr := httptest.NewRecorder()
	errs := []error{
		errors.New("test-error-1"),
		errors.New("test-error-2"),
	}
	ReturnAPIErrors(errs, rr)
	if res := rr.Result(); res.StatusCode != http.StatusBadRequest {
		t.Error("Expected bad request response, got:", res.StatusCode)
	}

	rr = httptest.NewRecorder()
	ReturnAPIForbidden(nil, "forbidden", rr)
	if res := rr.Result(); res.StatusCode != http.StatusForbidden {
		t.Error("Expected forbidden response, got:", res.StatusCode)
	}

	rr = httptest.NewRecorder()
	ReturnAPIForbidden(nil, "forbidden", rr)
	if res := rr.Result(); res.StatusCode != http.StatusForbidden {
		t.Error("Expected forbidden response, got:", res.StatusCode)
	}

	rr = httptest.NewRecorder()
	ReturnAPIForbidden(errors.New("fake error"), "forbidden", rr)
	if res := rr.Result(); res.StatusCode != http.StatusForbidden {
		t.Error("Expected forbidden response, got:", res.StatusCode)
	}

}

func TestWriteJSON(t *testing.T) {
	testCases := []struct {
		In              interface{}
		ExpectedStatus  int
		ExpectedContent string
	}{
		{
			In:              "fake response",
			ExpectedStatus:  http.StatusOK,
			ExpectedContent: `"fake response"`,
		},
		{
			In:             make(chan int),
			ExpectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range testCases {
		rr := httptest.NewRecorder()
		WriteJSON(test.In, rr)
		res := rr.Result()
		if body, err := readResponseBody(res); err != nil {
			t.Fatal(err)
		} else if test.ExpectedContent != "" && strings.TrimSpace(string(body)) != test.ExpectedContent {
			t.Error("Expected:", test.ExpectedContent, "Got:", string(body))
		}
		if res.StatusCode != test.ExpectedStatus {
			t.Error("Expected status:", test.ExpectedStatus, "Got status:", res.StatusCode)
		}
	}
}

func TestUnmarshalRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "/", bytes.NewBuffer([]byte(`true`)))
	var res bool
	if err := UnmarshalRequest(req, &res); err != nil {
		t.Fatal(err)
	} else if res == false {
		t.Error("Expected result to be true")
	}

	req = httptest.NewRequest("GET", "/", bytes.NewBuffer([]byte(`bad json`)))
	if err := UnmarshalRequest(req, &res); err == nil {
		t.Error("Expected error unmarshaling request,")
	}
}

func TestWriteOK(t *testing.T) {
	w := httptest.NewRecorder()
	WriteOK(w)
	res := w.Result()
	body, err := readResponseBody(res)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Error("Expected status OK, got:", res.StatusCode)
	}
	resBody := make(map[string]bool)
	if err := json.Unmarshal(body, &resBody); err != nil {
		t.Fatal(err)
	}
	if val, ok := resBody["ok"]; !ok {
		t.Error("Expected 'ok' value in response")
	} else if !val {
		t.Error("Expected 'ok' value to be true")
	}
}

func TestFilterUserRolesByName(t *testing.T) {
	roles := []v1alpha1.VDIRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-role-one",
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-role-two",
			},
		},
	}

	filtered := FilterUserRolesByNames(roles, []string{"test-role-one"})
	if len(filtered) != 1 {
		t.Fatal("Expected one role returned")
	}
	if filtered[0].GetName() != "test-role-one" {
		t.Error("Expected name of returned role to be 'test-role-one', got:", filtered[0].GetName())
	}
}
