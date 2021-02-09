package api

import (
	"context"
	"net/http"

	"github.com/tinyzimmer/kvdi/pkg/util/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
)

// swagger:operation GET /api/serviceaccounts/{namespace} Miscellaneous getServiceAccounts
// ---
// summary: Retrieve the service accounts in the given namespace that the user is allowed to use
// parameters:
// - name: namespace
//   in: path
//   description: The namespace to list service accounts in
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/serviceAccountsResponse"
//   "400":
//     "$ref": "#/responses/error"
//   "403":
//     "$ref": "#/responses/error"
func (d *desktopAPI) GetServiceAccounts(w http.ResponseWriter, r *http.Request) {
	namespace := apiutil.GetNamespaceFromRequest(r)
	sess := apiutil.GetRequestUserSession(r)
	serviceAccounts, err := d.ListServiceAccounts(namespace)
	if err != nil {
		apiutil.ReturnAPIError(err, w)
		return
	}
	apiutil.WriteJSON(sess.User.FilterServiceAccounts(serviceAccounts, namespace), w)
}

// ListServiceAccounts returns a string slice of all the service accounts
// in a given namespace.
func (d *desktopAPI) ListServiceAccounts(ns string) ([]string, error) {
	saList := &corev1.ServiceAccountList{}
	if err := d.client.List(context.TODO(), saList, client.InNamespace(ns)); err != nil {
		return nil, err
	}
	saNames := make([]string, 0)
	for _, sa := range saList.Items {
		saNames = append(saNames, sa.GetName())
	}
	return saNames, nil
}

// Service Accounts Response
// swagger:response serviceAccountsResponse
type swaggerServiceAccountsResponse struct {
	// in:body
	Body []string
}
