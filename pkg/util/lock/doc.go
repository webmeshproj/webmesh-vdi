/*
Copyright 2020,2021 Avi Zimmerman

This file is part of kvdi.

kvdi is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kvdi is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with kvdi.  If not, see <https://www.gnu.org/licenses/>.
*/

// Package lock implements a ConfigMap lock similar to the one found in
// github.com/operator-framework/operator-sdk/pkg/leader.
//
// The main difference is that it provides a Release mechanism as opposed to
// the Leader-for-life strategy the operator-sdk uses. This is useful for getting
// temporary locks on K8s resources. Also, since the user of the lock will not
// always dissapear, an expiration key is placed in the configMap to signal to
// another process when it's okay to release a stale lock.
package lock
