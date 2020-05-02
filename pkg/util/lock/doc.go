// Package lock implements a ConfigMap lock similar to the one found in
// github.com/operator-framework/operator-sdk/pkg/leader.
//
// The main difference is that it provides a Release mechanism as opposed to
// the Leader-for-life strategy the operator-sdk uses. This is useful for getting
// temporary locks on K8s resources. Also, since the user of the lock will not
// always dissapear, an expiration key is placed in the configMap to signal to
// another process when it's okay to release a stale lock.
package lock
