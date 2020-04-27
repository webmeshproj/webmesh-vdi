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
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"
)

// LocalAuthProvider implements an AuthProvider that uses the rethinkdb database
// as the store for user credentials. This is currently tightly coupled with
// the rethinkdb util pkg which expects password information in user objects.
// However, it's possible those fields can just be ignored when implementing new
// providers.
type LocalAuthProvider struct {
	types.AuthProvider

	// utility functions for mocking
	getDB     func() (rethinkdb.RethinkDBSession, error)
	getKey    func() ([]byte, error)
	signToken func([]byte, *types.User) (types.JWTClaims, string, error)
	compHash  func(string, string) bool
}

// New returns a new LocalAuthProvider.
func New() types.AuthProvider {
	return &LocalAuthProvider{
		compHash:  common.PasswordMatchesHash,
		signToken: apiutil.GenerateJWT,
		getKey: func() ([]byte, error) {
			_, key := tlsutil.ServerKeypair()
			return ioutil.ReadFile(key)
		},
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

	secret, err := a.getKey()
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	claims, newToken, err := a.signToken(secret, user)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	response := &types.SessionResponse{
		Token:     newToken,
		ExpiresAt: claims.ExpiresAt,
		User:      user,
	}

	apiutil.WriteJSON(response, w)
}
