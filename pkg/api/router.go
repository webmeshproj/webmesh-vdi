package api

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/tinyzimmer/kvdi/pkg/auth"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

func (d *desktopAPI) buildRouter() error {
	r := mux.NewRouter()

	loginHandler := auth.GetAuthProvider(d.vdiCluster)
	if err := loginHandler.Setup(d.vdiCluster); err != nil {
		return err
	}
	r.PathPrefix("/api/login").HandlerFunc(loginHandler.Authenticate)

	protected := r.PathPrefix("/api").Subrouter()
	protected.HandleFunc("/users", d.GetUsers).Methods("GET")
	protected.HandleFunc("/users/{user}", d.GetUser).Methods("GET")
	protected.HandleFunc("/templates", d.GetDesktopTemplates).Methods("GET")
	protected.HandleFunc("/sessions", d.StartDesktopSession).Methods("POST")
	protected.HandleFunc("/sessions/{namespace}/{name}", d.GetSessionStatus).Methods("GET")

	protected.Use(d.ValidateUserSession)

	d.router = r
	return nil
}

func (d *desktopAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.router.ServeHTTP(w, r)
}

func (d *desktopAPI) ValidateUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Session-Token")
		if token == "" && !d.vdiCluster.AnonymousAllowed() {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
		if err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		defer sess.Close()
		if token == "" && d.vdiCluster.AnonymousAllowed() {
			newSession, err := sess.CreateUserSession(&rethinkdb.User{Name: "anonymous"})
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			r.Header.Set("X-Session-Token", newSession.ID)
			w.Header().Set("X-Session-Token", newSession.ID)
			next.ServeHTTP(w, r)
			return
		}
		if _, err := sess.GetUserSession(token); err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
