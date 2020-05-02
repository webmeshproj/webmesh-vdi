// Package reconcile contains functions for reconciling Kubernetes resources.
//
// The functions in this package are intended to be idempotent. They should create
// resources that don't exist, and when they do exist, they should be checked for
// equality with the desired state and updated and/or requeued appropriately.
package reconcile
