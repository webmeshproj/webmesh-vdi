package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/tinyzimmer/kvdi/pkg/auth"
)

func (d *desktopAPI) buildRouter() error {
	r := mux.NewRouter()

	loginHandler := auth.GetAuthProvider(d.vdiCluster)
	if err := loginHandler.Setup(d.vdiCluster); err != nil {
		return err
	}
	r.PathPrefix("/api/login").HandlerFunc(loginHandler.Authenticate).Methods("POST")

	protected := r.PathPrefix("/api").Subrouter()
	protected.HandleFunc("/whoami", d.WhoAmI).Methods("GET")
	protected.HandleFunc("/users", d.GetUsers).Methods("GET")
	protected.HandleFunc("/users/{user}", d.GetUser).Methods("GET")
	protected.HandleFunc("/templates", d.GetDesktopTemplates).Methods("GET")
	protected.HandleFunc("/sessions", d.StartDesktopSession).Methods("POST")
	protected.HandleFunc("/sessions/{namespace}/{name}", d.GetSessionStatus).Methods("GET")
	protected.HandleFunc("/websockify/{endpoint}", mtlsWebsockify)

	protected.Use(d.ValidateUserSession)

	d.router = r
	return nil
}

func (d *desktopAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.router.ServeHTTP(w, r)
}
