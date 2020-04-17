package apiutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util/errors"
)

type AuthProvider interface {
	Setup(*v1alpha1.VDICluster) error
	Authenticate(w http.ResponseWriter, r *http.Request)
}

func WriteOrLogError(out []byte, w http.ResponseWriter) {
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(append(out, []byte("\n")...)); err != nil {
		fmt.Println("Failed to write API response:", string(out))
	}
}

func ReturnAPIError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	fmt.Println("API Error:", err.Error())
	WriteOrLogError(errors.ToAPIError(err).JSON(), w)
}

func WriteJSON(i interface{}, w http.ResponseWriter) {
	out, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		ReturnAPIError(err, w)
		return
	}
	WriteOrLogError(out, w)
}

func UnmarshalRequest(r *http.Request, in interface{}) error {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, in)
}
