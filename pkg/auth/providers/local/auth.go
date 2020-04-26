package local

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/auth/types"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/common"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
)

// LocalAuthProvider implements an AuthProvider that uses the rethinkdb database
// as the store for user credentials. This is currently tightly coupled with
// the rethinkdb util pkg which expects password information in user objects.
// However, it's possible those fields can just be ignored when implementing new
// providers.
type LocalAuthProvider struct {
	types.AuthProvider

	// getDB is a function configureed at initialization that
	// retrieves a db connection.
	getDB    func() (rethinkdb.RethinkDBSession, error)
	compHash func(string, string) bool
}

// New returns a new LocalAuthProvider.
func New() types.AuthProvider {
	return &LocalAuthProvider{
		compHash: common.PasswordMatchesHash,
	}
}

// Setup implements the AuthProvider interface and configures the provider's
// database connection options.
func (a *LocalAuthProvider) Setup(cluster *v1alpha1.VDICluster) error {
	rdbAddr := rethinkdb.RDBAddrForCR(cluster)
	a.getDB = func() (rethinkdb.RethinkDBSession, error) {
		sess, err := rethinkdb.New(rdbAddr)
		if err != nil {
			return nil, err
		}
		return sess, nil
	}
	return nil
}

// Authenticate implements AuthProvider and simply checks the provided password
// in the request against the hash in the database.
func (a *LocalAuthProvider) Authenticate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	req := &types.LoginRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	sess, err := a.getDB()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()

	user, err := sess.GetUser(req.Username)
	if err != nil {
		apiutil.ReturnAPIForbidden(nil, "Invalid credentials", w)
		return
	}
	if !a.compHash(req.Password, user.PasswordSalt) {
		apiutil.ReturnAPIForbidden(nil, "Invalid credentials", w)
		return
	}

	session, err := sess.CreateUserSession(user)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(session, w)
}
