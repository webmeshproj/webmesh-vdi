package local

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1"
	"github.com/tinyzimmer/kvdi/pkg/util"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"github.com/tinyzimmer/kvdi/pkg/util/rethinkdb"
	"github.com/tinyzimmer/kvdi/pkg/util/tlsutil"
)

type LocalAuthProvider struct {
	apiutil.AuthProvider

	tlsConfig *tls.Config
	rdbAddr   string
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func New() apiutil.AuthProvider {
	return &LocalAuthProvider{}
}

func (a *LocalAuthProvider) Setup(cluster *v1alpha1.VDICluster) error {
	var err error
	a.rdbAddr = rethinkdb.RDBAddrForCR(cluster)
	a.tlsConfig, err = tlsutil.NewClientTLSConfig()
	return err
}

func (a *LocalAuthProvider) Authenticate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	req := &LoginRequest{}
	if err := json.Unmarshal(body, req); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	sess, err := rethinkdb.New(a.rdbAddr)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	defer sess.Close()

	user, err := sess.GetUser(req.Username)
	if err != nil {
		apiutil.ReturnAPIError(errors.New("Invalid credentials"), w)
		return
	}
	if !util.PasswordMatchesHash(req.Password, user.PasswordSalt) {
		apiutil.ReturnAPIError(errors.New("Invalid credentials"), w)
		return
	}

	session, err := sess.CreateUserSession(user)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}

	apiutil.WriteJSON(session, w)
}
