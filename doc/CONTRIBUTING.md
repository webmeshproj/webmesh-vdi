# Contributing to kVDI

_TODO: Elaborate further everywhere. For now please refer to the [`godoc`](https://pkg.go.dev/github.com/tinyzimmer/kvdi) for finding your way around the code base_

If you are familiar with other `operator` projects, then this code base should be relatively easy to navigate.

Majority of types and K8s API definitions can be found in `pkg/apis`. These are still open to many change unless the project reaches a "stable release" phase.

## Controllers/Managers

### Relevant packages

  - `pkg/controller`
  - `pkg/resources`

## App/API

### Relevant Packages

  - `pkg/api`
  - `pkg/auth`
  - `pkg/secrets`

## UI

The UI is written in `Vue.js` using the [`quasar`](https://quasar.dev/) framework (which is awesome).

### Relevant Packages

  - `ui/app`
