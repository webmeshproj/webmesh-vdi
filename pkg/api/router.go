package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (d *desktopAPI) buildRouter() {
	r := mux.NewRouter()
	r.Path("/api/templates").Methods("GET").HandlerFunc(d.GetDesktopTemplates)
	r.Path("/api/sessions").Methods("POST").HandlerFunc(d.StartDesktopSession)
	r.Path("/api/sessions/{namespace}/{name}").Methods("GET").HandlerFunc(d.GetSessionStatus)
	d.router = r
}

func (d *desktopAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.router.ServeHTTP(w, r)
}
