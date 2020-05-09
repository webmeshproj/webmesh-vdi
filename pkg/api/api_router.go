package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// buildRouter builds the API router
func (d *desktopAPI) buildRouter() error {
	r := mux.NewRouter()

	// Setup the decoder
	r.Use(DecodeRequest)

	// Login route is not protected since it generates the tokens for which a user
	// can use the protected routes.
	r.PathPrefix("/api/login").HandlerFunc(d.PostLogin).Methods("POST")

	// Main HTTP routes

	protected := r.PathPrefix("/api").Subrouter()

	// SUBROUTER ASSUMES /api PREFIX ON ALL ROUTES

	// Authorizing tokens for when MFA is required
	protected.HandleFunc("/authorize", d.PostAuthorize).Methods("POST")

	// Misc routes
	protected.HandleFunc("/logout", d.PostLogout).Methods("POST")
	protected.HandleFunc("/whoami", d.GetWhoAmI).Methods("GET")
	protected.HandleFunc("/config", d.GetConfig).Methods("GET")
	protected.HandleFunc("/config/reload", d.PostReloadConfig).Methods("POST")
	protected.HandleFunc("/namespaces", d.GetNamespaces).Methods("GET")

	// User operations
	protected.HandleFunc("/users", d.GetUsers).Methods("GET")
	protected.HandleFunc("/users", d.PostUsers).Methods("POST")
	protected.HandleFunc("/users/{user}", d.GetUser).Methods("GET")
	protected.HandleFunc("/users/{user}", d.PutUser).Methods("PUT")
	protected.HandleFunc("/users/{user}/mfa", d.GetUserMFA).Methods("GET")
	protected.HandleFunc("/users/{user}/mfa", d.PutUserMFA).Methods("PUT")
	protected.HandleFunc("/users/{user}", d.DeleteUser).Methods("DELETE")

	// Role operations
	protected.HandleFunc("/roles", d.GetRoles).Methods("GET")
	protected.HandleFunc("/roles", d.CreateRole).Methods("POST")
	protected.HandleFunc("/roles/{role}", d.GetRole).Methods("GET")
	protected.HandleFunc("/roles/{role}", d.UpdateRole).Methods("PUT")
	protected.HandleFunc("/roles/{role}", d.DeleteRole).Methods("DELETE")

	// Template operations
	protected.HandleFunc("/templates", d.GetDesktopTemplates).Methods("GET")
	protected.HandleFunc("/templates", d.PostDesktopTemplates).Methods("POST")
	protected.HandleFunc("/templates/{template}", d.GetDesktopTemplate).Methods("GET")
	protected.HandleFunc("/templates/{template}", d.PutDesktopTemplate).Methods("PUT")
	protected.HandleFunc("/templates/{template}", d.DeleteDesktopTemplate).Methods("DELETE")

	// Desktop session operations
	protected.HandleFunc("/sessions", d.StartDesktopSession).Methods("POST")
	protected.HandleFunc("/sessions/{namespace}/{name}", d.GetDesktopSessionStatus).Methods("GET")
	protected.HandleFunc("/sessions/{namespace}/{name}", d.DeleteDesktopSession).Methods("DELETE")

	// Websockify proxy
	protected.HandleFunc("/desktops/{namespace}/{name}/websockify", d.GetWebsockify)
	protected.HandleFunc("/desktops/{namespace}/{name}/wsaudio", d.GetWebsockifyAudio)

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
