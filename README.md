# Webmesh Desktop

A Virtual Desktop Infrastructure running on Kubernetes. With soon to come webmesh integration.

[![Go Report Card](https://goreportcard.com/badge/github.com/webmeshproj/webmesh-vdi)](https://goreportcard.com/report/github.com/webmeshproj/webmesh-vdi)
![Tests](https://github.com/webmeshproj/webmesh-vdi/actions/workflows/tests.yml/badge.svg)
![Build](https://github.com/webmeshproj/webmesh-vdi/actions/workflows/build.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/webmeshproj/webmesh-vdi.svg)](https://pkg.go.dev/github.com/webmeshproj/webmesh-vdi)

**ATTENTION:** The `helm` chart repository has been moved to a [separate repo](https://github.com/kvdi/helm-charts) to tidy things up here more. To update your repository you can do the following:

```sh
helm repo remove kvdi
helm repo add kvdi https://kvdi.github.io/helm-charts/charts
helm repo update

helm install kvdi kvdi/kvdi  # yes, that's a lot of kvdi
```

This project has reached a point where I am not going to be making enormous changes all the time anymore. As such I am tagging a "stable" release and incrementing from there.
That still doesn't mean I highly recommend it's usage, but rather I am relatively confident in its overall stability.

- [API Reference](doc/appv1.md)
  - [RBAC Reference](doc/rbacv1.md)
  - [Templates Reference](doc/desktopsv1.md)
- [Installing](#Installing)
  - [Standalone](#Install-standalone)
  - [Kubernetes](#Install-to-a-pre-existing-cluster)
    - [Helm](#helm)
    - [Bundle](#bundle-manifest)
    - [Kustomize](#kustomize)
- [CLI](doc/kvdictl/kvdictl.md)
- [Upgrading](#Upgrading)
- [Building Desktop Images](build/desktops/README.md)
- [Security](#security)
- [Screenshots/Video](doc/screenshots.md)
- [Donating](#donating)

If you are interested in helping out or just simply launching a design discussion, feel free to send PRs and/or issues.
I wrote up a [`CONTRIBUTING`](CONTRIBUTING.md) doc just outlining some of the stuff I have in mind that would need to be acomplished for this to be considered "stable".

## Features

- Containerized user desktops running on Kubernetes with no virtualization required (`libvirt` options may come in the future). All traffic between the end user and the "desktop" is encrypted.

- Persistent user data

- Audio playback and microphone support

- File transfer to/from "desktop" sessions. Directories get archived into a gzipped tarball prior to download.

- RBAC system for managing user access to templates, roles, users, namespaces, serviceaccounts, etc.

- MFA Support

- Configurable backend for internal secrets. Currently `vault` or Kubernetes Secrets

- Use built-in local authentication, LDAP, or OpenID.

- App metrics to either scrape externally or view in the UI. More details in the `helm` doc.

### TODO

- "App Profiles" - I have a POC implementation on `main` but it is still pretty buggy
- DOSBox/Game profiles could be cool...same as "App Profiles"
- ARM64 support. Should be easy, but the build files will need some shuffling.
- UI could use a serious makeover from someone who actually knows what they are doing

## Requirements

For building and running locally you will need:

- `go >= 1.14`
- `docker`

## Installing

### Install standalone

If you don't have access to a Kubernetes cluster, or you just want to try `kVDI` out on a VM real quick, there is a script in this repository for setting up kVDI using `k3s`.
It requires the instance running the script to have `docker` and the `dialog` package installed.

_If you have an existing `k3s` installation, the ingress may not work since this script assumes `kVDI` will be the only LoadBalancer installed._

```bash
# Download the script from this repository.
curl -JLO https://raw.githubusercontent.com/kvdi/kvdi/main/deploy/architect/kvdi-architect.sh
# Run the script. You will be prompted via dialogs to make configuration changes.
bash kvdi-architect.sh   # Use --help to see all available options.
```

_NOTE: This script is fairly new and still has some bugs_

### Install to a pre-existing cluster

#### Helm

For more complete installation instructions see the `helm` chart docs [here](https://github.com/kvdi/helm-charts/blob/main/charts/kvdi/README.md) for available configuration options.

The [API Reference](doc/appv1.md) can also be used for details on `kVDI` app-level configurations.

```bash
helm repo add kvdi https://kvdi.github.io/helm-charts/charts  # Add the kvdi charts repo
helm repo update                                              # Sync your repositories

# Install kVDI
helm install kvdi kvdi/kvdi
```

It will take a minute or two for all the parts to start running after the install command.
Once the app is launched, you can retrieve the admin password from `kvdi-admin-secret` in your cluster (if you are using `ldap` auth, log in with a user in one of the `adminGroups`).

To access the app interface either do a `port-forward` (`make forward-app` is another helper for that when developing locally with `kind`), or go to the "LoadBalancer" IP of the service.

By default there are no desktop templates configured. If you'd like, you can apply the ones in `deploy/examples/example-desktop-templates.yaml` to get started quickly.

#### Bundle Manifest

There is a manifest in this repository that will **just** lay down the manager instance, its dependencies, and all of the CRDs.
You can then create a [VDICluster](doc/appv1.md#VDIClusterSpec) object manually to spin up an actual application instance.

To install the manifest:

```bash
export KVDI_VERSION=v0.3.6

kubectl apply -f https://raw.githubusercontent.com/kvdi/kvdi/${KVDI_VERSION}/deploy/bundle.yaml --validate=false
```

#### Kustomize

The `kustomize` manifests in this repository are generated by `kubebuilder` and are usable as well similar to the [Bundle Manifest](#bundle-manifest).
They can be found in the [config](config/) directory in this repository.

## Upgrading

Most of the time you can just run a regular helm upgrade to update your deployment manifests to the latest images.

```bash
helm upgrade kvdi kvdi/kvdi --version v0.3.6
```

However, sometimes there may be changes to the CRDs, though I will always do my best to make sure they are backwards compatible.
Due to the way helm manages CRDs, it will ignore changes to those on an existing installation.
You can get around this by applying the CRDs for the version you are upgrading to directly from this repo.

For example:

```bash
export KVDI_VERSION=v0.3.6

kubectl apply \
  -f https://github.com/kvdi/kvdi/raw/${KVDI_VERSION}/config/crd/bases/app.kvdi.io_vdiclusters.yaml \
  -f https://github.com/kvdi/kvdi/raw/${KVDI_VERSION}/config/crd/bases/desktops.kvdi.io_sessions.yaml \
  -f https://github.com/kvdi/kvdi/raw/${KVDI_VERSION}/config/crd/bases/desktops.kvdi.io_templates.yaml \
  -f https://github.com/kvdi/kvdi/raw/${KVDI_VERSION}/config/crd/bases/rbac.kvdi.io_vdiroles.yaml
```

When there is a change to one or more CRDs, it will be mentioned in the notes for that release.

## Building and Running Locally

The `Makefile` contains helpers for testing the full solution locally using `k3d`. Run `make help` to see all the available options.

_If you choose to pull the images from the registry instead of building and loading first - you probably want to set `VERSION=latest` (or a previous version) in your environment also.
The `Makefile` is usually pointed at the next version to be released and published images may not exist yet_.

```bash
# Builds all the docker images (optional, they are also available in the github registry)
$> make build-all
# Spin up a kind cluster for local testing
$> make test-cluster
# Load all the docker images into the kind cluster (optional for same reason as build)
$> make load-all
# Deploy the manager, kvdi, and setup the example templates
$> make deploy example-vdi-templates
# To test using custom helm values
$> HELM_ARGS="-f my_values.yaml" make deploy
```

After the manager has started the `app` instance, get the IP of its service with `kubectl get svc` to access the frontend, or you can run `make-forward-app` to start a local port-forward.

If not using anonymous auth, look for `kvdi-admin-secret` to retrieve the `admin` password.

## Security

All traffic between processes is encrypted with mTLS.
The UI for the "desktop" containers is placed behind a VNC server listening on a UNIX socket and a sidecar to the container will proxy validated websocket connections to it.

![img](doc/kvdi_arch.png)

User authentication is provided by "providers". There are currently three implementations:

- `local-auth` : A `passwd` like file is kept in the Secrets backend (k8s or vault) mapping users to roles and password hashes. This is primarily meant for development, but you could secure your environment in a way to make it viable for a small number of users.

- `ldap-auth` : An LDAP/AD server is used for autenticating users. VDIRoles can be tied to
  security groups in LDAP via annotations. When a user is authenticated, their groups are queried to see if they are bound to any VDIRoles.

- `oidc-auth` : An OpenID or OAuth provider is used for authenticating users. If using an Oauth provider, it must support the `openid` scope. When a user is authenticated, a configurable `groups` claim is requested from the provider that can be mapped to VDIRoles similarly to `ldap-auth`. If the provider does not support a `groups` claim, you can configure `kVDI` to allow all authenticated users.

All three authentication methods also support MFA.

# Donating

kVDI started as just a fun project, but as more people have started to use it, I've really wanted to find more time to continue making improvements.
Unfortunately, it just doesn't pay the bills so I can only really settle into it when I have nothing else going on.
I've set up a Patreon and an ETH wallet if people would like to support further development:

- Patreon: [![Support me on Patreon](https://img.shields.io/endpoint.svg?url=https%3A%2F%2Fshieldsio-patreon.vercel.app%2Fapi%3Fusername%3Dkvdi%26type%3Dpatrons&style=flat)](https://www.patreon.com/kvdi?fan_landing=true)

- ETH: `0xdFC61298BdFe4a6F7fb1eFae5Da27d905c1bD172`

Feel free to email me at the address on my Github profile if you have any other questions.
