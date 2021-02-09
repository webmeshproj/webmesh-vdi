package api

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/websocket"
)

type proxyLogger struct{}

func (p *proxyLogger) Printf(format string, args ...interface{}) {
	apiLogger.Info(fmt.Sprintf(format, args...))
}

// buildRouter builds the API router
func (d *desktopAPI) buildRouter() error {
	r := mux.NewRouter()

	// Run the metrics middleware first
	r.Use(prometheusMiddleware)

	// Setup the decoder
	r.Use(DecodeRequest)

	// metrics
	r.PathPrefix("/api/metrics").Handler(promhttp.Handler())

	// Readiness/liveness probes
	r.PathPrefix("/api/healthz").HandlerFunc(d.Healthz).Methods("GET")
	r.PathPrefix("/api/readyz").HandlerFunc(d.Readyz).Methods("GET")

	// Grafana proxy - This is unprotected for now, but should figure out what
	// permission model would work well for it. It doesn't fit well into the existing
	// paradigms.
	grafanaProxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:3000",
	})
	grafanaProxy.ModifyResponse = func(res *http.Response) error {
		res.Header.Del("X-Frame-Options")
		return nil
	}
	r.PathPrefix("/api/grafana").Handler(grafanaProxy)

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
	protected.HandleFunc("/logout", d.PostLogout).Methods("POST")                             // Cleans up user's desktops
	protected.HandleFunc("/whoami", d.GetWhoAmI).Methods("GET")                               // Convenience route for decoding JWTs
	protected.HandleFunc("/config", d.GetConfig).Methods("GET")                               // Retrieve server configuration
	protected.HandleFunc("/namespaces", d.GetNamespaces).Methods("GET")                       // Retrieve a list of available namespaces for the requesting user
	protected.HandleFunc("/serviceaccounts/{namespace}", d.GetServiceAccounts).Methods("GET") // Retrieve a list of available service accounts for the requesting user

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
	protected.HandleFunc("/sessions", d.GetDesktopSessions).Methods("GET")                         // Retrieve status information for all desktop sessions
	protected.HandleFunc("/sessions", d.StartDesktopSession).Methods("POST")                       // Start a new desktop session
	protected.HandleFunc("/sessions/{namespace}/{name}", d.GetDesktopSessionStatus).Methods("GET") // Get the status of a desktop session
	protected.HandleFunc("/sessions/{namespace}/{name}", d.DeleteDesktopSession).Methods("DELETE") // Stop a desktop session

	// Methods for interacting with the kvdi-proxy
	// // Plain HTTP routes
	protected.HandleFunc("/desktops/{namespace}/{name}/logs/{container}", d.GetDesktopLogs).Methods("GET") // Retrieve the logs a container in the desktop
	// // Websocket routes
	protected.Path("/desktops/ws/{namespace}/{name}/status").Handler(&websocket.Server{ // Do a follow the session status for a desktop. Used to query connect readiness.
		Handshake: func(*websocket.Config, *http.Request) error { return nil },
		Handler:   d.GetDesktopSessionStatusWebsocket,
	})
	protected.Path("/desktops/ws/{namespace}/{name}/logs/{container}").Handler(&websocket.Server{ // Do a follow of the logs for a container in the desktop
		Handshake: func(*websocket.Config, *http.Request) error { return nil },
		Handler:   d.GetDesktopLogsWebsocket,
	})
	protected.HandleFunc("/desktops/ws/{namespace}/{name}/display", d.GetWebsockify)    // Connect to the VNC socket on a desktop over websockets
	protected.HandleFunc("/desktops/ws/{namespace}/{name}/audio", d.GetWebsockifyAudio) // Connect to the audio stream of a desktop over websockets

	// // Filesystem access
	protected.PathPrefix("/desktops/fs/{namespace}/{name}/stat/").HandlerFunc(d.GetStatDesktopFile).Methods("GET")    // Retrieve file info or a directory listing from a desktop
	protected.PathPrefix("/desktops/fs/{namespace}/{name}/get/").HandlerFunc(d.GetDownloadDesktopFile).Methods("GET") // Retrieve the contents of a file from a desktop
	protected.HandleFunc("/desktops/fs/{namespace}/{name}/put", d.PutDesktopFile).Methods("PUT")                      // Uploads a file to a desktop

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
