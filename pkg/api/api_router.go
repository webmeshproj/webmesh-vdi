package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth"

	"github.com/gorilla/mux"
)

func (d *desktopAPI) buildRouter() error {
	r := mux.NewRouter()

	loginHandler := auth.GetAuthProvider(d.vdiCluster)
	if err := loginHandler.Setup(d.vdiCluster); err != nil {
		return err
	}
	r.PathPrefix("/api/login").HandlerFunc(loginHandler.Authenticate).Methods("POST")

	// Main HTTP routes

	protected := r.PathPrefix("/api").Subrouter()

	// Subrouter assumes /api prefix
	protected.HandleFunc("/logout", d.Logout).Methods("POST")
	protected.HandleFunc("/whoami", d.WhoAmI).Methods("GET")

	protected.HandleFunc("/users", d.GetUsers).Methods("GET")
	protected.HandleFunc("/users", d.CreateUser).Methods("POST")
	protected.HandleFunc("/users/{user}", d.GetUser).Methods("GET")
	protected.HandleFunc("/users/{user}", d.UpdateUser).Methods("PUT")
	protected.HandleFunc("/users/{user}", d.DeleteUser).Methods("DELETE")

	protected.HandleFunc("/roles", d.GetRoles).Methods("GET")
	protected.HandleFunc("/roles", d.CreateRole).Methods("POST")
	protected.HandleFunc("/roles/{role}", d.GetRole).Methods("GET")
	protected.HandleFunc("/roles/{role}", d.UpdateRole).Methods("PUT")
	protected.HandleFunc("/roles/{role}", d.DeleteRole).Methods("DELETE")

	protected.HandleFunc("/templates", d.GetDesktopTemplates).Methods("GET")
	protected.HandleFunc("/sessions", d.StartDesktopSession).Methods("POST")
	protected.HandleFunc("/sessions/{namespace}/{name}", d.GetDesktopSessionStatus).Methods("GET")
	protected.HandleFunc("/sessions/{namespace}/{name}", d.DeleteDesktopSession).Methods("DELETE")

	protected.HandleFunc("/websockify/{namespace}/{name}", d.mtlsWebsockify)

	// Validate the user session on all requests followed by a grant check
	protected.Use(d.ValidateUserSession)
	protected.Use(d.ValidateUserGrants)

	d.router = r
	return nil
}

func (d *desktopAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.router.ServeHTTP(w, r)
}

func (d *desktopAPI) CreateUser(w http.ResponseWriter, r *http.Request) {
	return
}

func (d *desktopAPI) DeleteUser(w http.ResponseWriter, r *http.Request) {
	return
}

func (d *desktopAPI) UpdateUser(w http.ResponseWriter, r *http.Request) {
	return
}

func (d *desktopAPI) GetRoles(w http.ResponseWriter, r *http.Request) {
	return
}

func (d *desktopAPI) GetRole(w http.ResponseWriter, r *http.Request) {
	return
}

func (d *desktopAPI) CreateRole(w http.ResponseWriter, r *http.Request) {
	return
}

func (d *desktopAPI) DeleteRole(w http.ResponseWriter, r *http.Request) {
	return
}

func (d *desktopAPI) UpdateRole(w http.ResponseWriter, r *http.Request) {
	return
}
