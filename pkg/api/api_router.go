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
	// TODO: Route accepts GET also to support the oidc flow. Method should probably
	// be renamed.  To be honest, the entire OIDC flow is a bit hacky and should be reworked.
	r.PathPrefix("/api/login").HandlerFunc(d.PostLogin).Methods("POST", "GET")

	r.PathPrefix("/api/refresh_token").HandlerFunc(d.GetRefreshToken).Methods("GET") // Refresh a user's access token

	// Main HTTP routes

	protected := r.PathPrefix("/api").Subrouter()

	// SUBROUTER ASSUMES /api PREFIX ON ALL ROUTES

	protected.HandleFunc("/authorize", d.PostAuthorize).Methods("POST") // Verify a user's MFA token

	// Misc routes
	protected.HandleFunc("/logout", d.PostLogout).Methods("POST")              // Cleans up user's desktops
	protected.HandleFunc("/whoami", d.GetWhoAmI).Methods("GET")                // Convenience route for decoding JWTs
	protected.HandleFunc("/config", d.GetConfig).Methods("GET")                // Retrieve server configuration
	protected.HandleFunc("/config/reload", d.PostReloadConfig).Methods("POST") // Reload the server configuration, this should become a watcher
	protected.HandleFunc("/namespaces", d.GetNamespaces).Methods("GET")        // Retrieve a list of available namespaces for the requesting user

	// User operations
	protected.HandleFunc("/users", d.GetUsers).Methods("GET")                           // Retrieve a list of all users
	protected.HandleFunc("/users", d.PostUsers).Methods("POST")                         // Create a new user
	protected.HandleFunc("/users/{user}", d.GetUser).Methods("GET")                     // Retrieve information for a single user
	protected.HandleFunc("/users/{user}", d.PutUser).Methods("PUT")                     // Update a user
	protected.HandleFunc("/users/{user}/mfa", d.GetUserMFA).Methods("GET")              // Retrieve MFA status for a user
	protected.HandleFunc("/users/{user}/mfa", d.PutUserMFA).Methods("PUT")              // Update MFA status for a user
	protected.HandleFunc("/users/{user}/mfa/verify", d.PutUserMFAVerify).Methods("PUT") // Verify that a user has succesfully configured MFA
	protected.HandleFunc("/users/{user}", d.DeleteUser).Methods("DELETE")               // Delete a user

	// Role operations
	protected.HandleFunc("/roles", d.GetRoles).Methods("GET")             // Retrieve a list of all VDIRoles
	protected.HandleFunc("/roles", d.CreateRole).Methods("POST")          // Create a new VDIRole
	protected.HandleFunc("/roles/{role}", d.GetRole).Methods("GET")       // Retrieve information for a single VDIRole
	protected.HandleFunc("/roles/{role}", d.UpdateRole).Methods("PUT")    // Update a VDIRole
	protected.HandleFunc("/roles/{role}", d.DeleteRole).Methods("DELETE") // Delete a VDIRole

	// Template operations
	protected.HandleFunc("/templates", d.GetDesktopTemplates).Methods("GET")                 // Retrieve a list of all available DesktopTemplates
	protected.HandleFunc("/templates", d.PostDesktopTemplates).Methods("POST")               // Create a new DesktopTemplate
	protected.HandleFunc("/templates/{template}", d.GetDesktopTemplate).Methods("GET")       // Retrieve information for a single DesktopTemplate
	protected.HandleFunc("/templates/{template}", d.PutDesktopTemplate).Methods("PUT")       // Update a DesktopTemplate
	protected.HandleFunc("/templates/{template}", d.DeleteDesktopTemplate).Methods("DELETE") // Delete a DesktopTemplate

	// Desktop session operations
	protected.HandleFunc("/sessions", d.StartDesktopSession).Methods("POST")                       // Start a new desktop session
	protected.HandleFunc("/sessions/{namespace}/{name}", d.GetDesktopSessionStatus).Methods("GET") // Get the status of a desktop session
	protected.HandleFunc("/sessions/{namespace}/{name}", d.DeleteDesktopSession).Methods("DELETE") // Stop a desktop session

	// Websockify proxy
	protected.HandleFunc("/desktops/{namespace}/{name}/websockify", d.GetWebsockify)   // Connect to the VNC socket on a desktop over websockets
	protected.HandleFunc("/desktops/{namespace}/{name}/wsaudio", d.GetWebsockifyAudio) // Connect to the audio stream of a desktop over websockets

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
