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
