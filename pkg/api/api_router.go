package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth"

	"github.com/gorilla/mux"
)

// buildRouter builds the API router
func (d *desktopAPI) buildRouter() error {
	r := mux.NewRouter()

	// Setup the decoder
	r.Use(DecodeRequest)

	// login handler is provided by the authentication provider, does not go
	// through middlewares
	loginHandler := auth.GetAuthProvider(d.vdiCluster)
	if err := loginHandler.Setup(d.vdiCluster); err != nil {
		return err
	}
	r.PathPrefix("/api/login").HandlerFunc(loginHandler.Authenticate).Methods("POST")

	// Main HTTP routes

	protected := r.PathPrefix("/api").Subrouter()

	// Subrouter assumes /api prefix

	// Misc routes
	protected.HandleFunc("/logout", d.Logout).Methods("POST")
	protected.HandleFunc("/whoami", d.WhoAmI).Methods("GET")
	protected.HandleFunc("/config", d.GetConfig).Methods("GET")
	protected.HandleFunc("/grants", d.GetGrants).Methods("GET")
	protected.HandleFunc("/namespaces", d.GetNamespaces).Methods("GET")

	// User operations
	protected.HandleFunc("/users", d.GetUsers).Methods("GET")
	protected.HandleFunc("/users", d.CreateUser).Methods("POST")
	protected.HandleFunc("/users/{user}", d.GetUser).Methods("GET")
	protected.HandleFunc("/users/{user}", d.UpdateUser).Methods("PUT")
	protected.HandleFunc("/users/{user}", d.DeleteUser).Methods("DELETE")

	// Role operations
	protected.HandleFunc("/roles", d.GetRoles).Methods("GET")
	protected.HandleFunc("/roles", d.CreateRole).Methods("POST")
	protected.HandleFunc("/roles/{role}", d.GetRole).Methods("GET")
	protected.HandleFunc("/roles/{role}", d.UpdateRole).Methods("PUT")
	protected.HandleFunc("/roles/{role}", d.DeleteRole).Methods("DELETE")

	// Desktop session operations
	protected.HandleFunc("/templates", d.GetDesktopTemplates).Methods("GET")
	protected.HandleFunc("/sessions", d.StartDesktopSession).Methods("POST")
	protected.HandleFunc("/sessions/{namespace}/{name}", d.GetDesktopSessionStatus).Methods("GET")
	protected.HandleFunc("/sessions/{namespace}/{name}", d.DeleteDesktopSession).Methods("DELETE")

	// Websockify proxy
	protected.HandleFunc("/websockify/{namespace}/{name}", d.mtlsWebsockify)

	// Validate the user session on all requests
	protected.Use(d.ValidateUserSession)
	// check the grants for the request user
	protected.Use(d.ValidateUserGrants)

	d.router = r
	return nil
}

// ServeHTTP implements an http.Handler for the API
func (d *desktopAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.router.ServeHTTP(w, r)
}
