package api

import (
	"net/http"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

func (d *desktopAPI) ValidateUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(TokenHeader)
		if token == "" && !d.vdiCluster.AnonymousAllowed() {
			apiutil.ReturnAPIForbidden(nil, "No token and allowAnonymous is false", w)
			return
		}
		sess, err := rethinkdb.New(rethinkdb.RDBAddrForCR(d.vdiCluster))
		if err != nil {
			apiutil.ReturnAPIForbidden(err, "Could not connect to database backend", w)
			return
		}
		defer sess.Close()
		if token == "" && d.vdiCluster.AnonymousAllowed() {
			newSession, err := sess.CreateUserSession(&rethinkdb.User{Name: "anonymous"})
			if err != nil {
				apiutil.ReturnAPIForbidden(err, "Could not create new user session", w)
				return
			}
			SetRequestUserSession(r, newSession)
			w.Header().Set(TokenHeader, newSession.ID)
			next.ServeHTTP(w, r)
			return
		}
		userSess, err := sess.GetUserSession(token)
		if err != nil {
			apiutil.ReturnAPIForbidden(err, "Could not retrieve user session", w)
			return
		} else if userSess.ExpiresAt.Before(time.Now()) {
			// TODO cleanup the session (maybe a seperate reaper process)
			apiutil.ReturnAPIForbidden(nil, "User session has expired", w)
			return
		}
		SetRequestUserSession(r, userSess)
		next.ServeHTTP(w, r)
	})
}
