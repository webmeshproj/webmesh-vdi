package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

type DecoderFunc func(r *http.Request) (interface{}, error)

var Decoders = map[string]map[string]DecoderFunc{
	"/api/sessions": {
		"POST": func(r *http.Request) (interface{}, error) {
			req := &PostSessionsRequest{}
			return req, apiutil.UnmarshalRequest(r, req)
		},
	},
}

func DecodeRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := getGorillaPath(r)
		if decoder, ok := Decoders[path]; ok {
			if decoderFunc, ok := decoder[r.Method]; ok {
				req, err := decoderFunc(r)
				if err != nil {
					apiutil.ReturnAPIError(err, w)
					return
				}
				SetRequestObject(r, req)
			}
		}
		next.ServeHTTP(w, r)
	})
}
