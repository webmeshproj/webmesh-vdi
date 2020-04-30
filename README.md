# kVDI

A Kubernetes-native Virtual Desktop Infrastructure.

**This is a very heavy work-in-progress and not even remotely close to ready for production usage**

 - [CRD Reference](doc/crds.md)
 - [Security](#security)
 - [Screenshots](#screenshots)

I'll write up CONTRIBUTING docs soon, but I am getting to the point where it'd be cool to have some collaboration.
If you are wanting this to become a real thing (or are just interested in trying it out), and you run into any issues, feel free to open an issue here and I can try to help out.

## Requirments

Cluster requirements

  - `cert-manager >= 0.14.1`
    - The manager uses the `v1alpha3` API for certificate provisioning
  - `snd-aloop` host-kernel support (optional, for sound emulation)

For building and running locally you will need:

  - `go >= 1.13`
  - `docker`

## Installing

Assuming you have `cert-manager` installed and running in the cluster:

```bash
$> helm install deploy/charts/kvdi
```

It will take a minute or two for all the parts to start running after the install command.
Once the app is launched, you can run `make get-admin-password` to retrieve the generated admin password.
To access the app interface either do a `port-forward` (`make forward-app` is another helper for that), or go to the "LoadBalancer" IP of the service.

## Building and Running Locally

The `Makefiles` contain helpers for testing the full solution locally using `kind`.

```bash
# Builds all the docker images (optional, they are also available in the quay repo)
$> make build-all
# Spin up a kind cluster and load cert-manager and metallb into it
$> make full-test-cluster
# Load all the docker images into the kind cluster (optional for same reason as build)
$> make load-all
# Deploy the manager, kvdi, and setup the example templates
$> make deploy example-vdi-templates
```

After the manager has finished spinning up the `app` instance, get the IP of its service with `kubectl get svc` to access the app interface.

If not using anonymous auth, look for `example-vdicluster-admin-secret` to retrieve the `admin` password.

## Security

All traffic between processes is encrypted with mTLS.
The UI for the "desktop" containers is placed behind a VNC server listening on a UNIX socket and a sidecar to the container will proxy validated websocket connections to it.

![img](doc/kvdi_arch.png)

User authentication is provided by "providers". Currently there is only one `local-auth` implementation meant for development and testing.
It keeps a `passwd` like file in a Kubernetes secret where it stores users, password hashes, and role mappings.

RBAC is provided by a `VDIRole` CRD that behaves similar to a Kubernetes `ClusterRole`.
These roles can restrict users to namespaces, desktop templates, and user/role management.
It should not be possible for a user to make an API request that grants them more priviliges than they already have.

## Screenshots

The UI is super basic but here are some shots of what I have so far.

There is actually more since I took these photos. Primarily user/role management.

![img](doc/templates.png)

![img](doc/term.png)

![img](doc/libre.png)
