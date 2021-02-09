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

// Package secrets contains an engine for reading and writing secrets from
// configurable backends. Currently only a K8s secret backend provider is
// available, but eventually other interfaces can be added such as for vault.
//
// The purpose of this package is to provide "filesystem" like access to
// sensitive values, e.g. JWT signing secrets, user credential hashes, OTP secrets,
// etc.
//
// The main methods provided are `ReadSecret`, `WriteSecret`, and `AppendSecret`
// with the added ability to grab locks and use an optional cache.
package secrets
