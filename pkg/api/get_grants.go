package api

import (
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/auth/grants"
	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
)

// swagger:route GET /api/grants Miscellaneous getGrants
// Retrieves a mapping of grants to their bit values.
// responses:
//   200: grantResponse
//   403: error
//   500: error
func (d *desktopAPI) GetGrants(w http.ResponseWriter, r *http.Request) {
	out := map[string]int{
		"Admin": int(grants.All),
	}
	for idx, grant := range grants.Grants {
		out[grants.GrantNames[idx]] = int(grant)
	}
	apiutil.WriteJSON(out, w)
}

// Grants response
// swagger:response grantResponse
type swaggerGrantResponse struct {
	// in:body
	Body map[string]int
}
