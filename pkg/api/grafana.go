package api

import (
	"io"
	"net/http"
	"path"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

func (d *desktopAPI) ProxyGrafana(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	u := r.URL
	u.Scheme = "http"
	u.Host = "127.0.0.1:3000"
	u.Path = path.Clean(u.Path)
	req, err := http.NewRequest(r.Method, u.String(), r.Body)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	w.WriteHeader(res.StatusCode)
	if _, err := io.Copy(w, res.Body); err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
}
