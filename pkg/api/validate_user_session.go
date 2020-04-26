package api

import (
	"net/http"
	"time"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

func (d *desktopAPI) ValidateUserSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get(TokenHeader)
		if token == "" {
			if keys, ok := r.URL.Query()["token"]; ok {
				token = keys[0]
			}
		}
		if token == "" {
			apiutil.ReturnAPIForbidden(nil, "No token provided in request", w)
			return
		}
		sess, err := d.getDB()
		if err != nil {
			apiutil.ReturnAPIForbidden(err, "Could not connect to database backend", w)
			return
		}
		defer sess.Close()
		userSess, err := sess.GetUserSession(token)
		if err != nil {
			apiutil.ReturnAPIForbidden(err, "Could not retrieve user session", w)
			return
		} else if userSess.ExpiresAt.Before(time.Now()) {
			if err := sess.DeleteUserSession(userSess); err != nil {
				apiLogger.Error(err, "Failed to remove user session from database", "Session.Token", userSess.Token)
			}
			if err := d.CleanupUserDesktops(userSess.User.Name); err != nil {
				apiLogger.Error(err, "Failed to cleanup user desktops", "User.Name", userSess.User.Name)
			}
			// TODO - need a separate reaper process
			apiutil.ReturnAPIForbidden(nil, "User session has expired", w)
			return
		}
		SetRequestUserSession(r, userSess)
		next.ServeHTTP(w, r)
	})
}
