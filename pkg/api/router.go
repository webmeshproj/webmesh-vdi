package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/tinyzimmer/kvdi/pkg/auth"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

const TokenHeader = "X-Session-Token"

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
		token := r.Header.Get(TokenHeader)
		if token == "" && !d.vdiCluster.AnonymousAllowed() {
			apiutil.ReturnAPIForbidden(errors.New("No token and allowAnonymous is false"), w)
			return
		}
		sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
		if err != nil {
			apiutil.ReturnAPIForbidden(err, w)
			return
		}
		defer sess.Close()
		if token == "" && d.vdiCluster.AnonymousAllowed() {
			newSession, err := sess.CreateUserSession(&rethinkdb.User{Name: "anonymous"})
			if err != nil {
				apiutil.ReturnAPIForbidden(err, w)
				return
			}
			r.Header.Set(TokenHeader, newSession.ID)
			w.Header().Set(TokenHeader, newSession.ID)
			next.ServeHTTP(w, r)
			return
		}
		if sess, err := sess.GetUserSession(token); err != nil {
			apiutil.ReturnAPIForbidden(err, w)
			return
		} else if sess.ExpiresAt.Before(time.Now()) {
			// TODO cleanup the session (maybe a seperate reaper process)
			apiutil.ReturnAPIForbidden(errors.New("Your session has expired"), w)
			return
		}
		next.ServeHTTP(w, r)
	})
}
