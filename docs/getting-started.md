# Getting started

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


**ATTENTION:** The `helm` chart repository has been moved to a [separate repo](https://github.com/kvdi/helm-charts) to tidy things up here more. To update your repository you can do the following:

```sh
helm repo remove kvdi
helm repo add kvdi https://kvdi.github.io/helm-charts/charts
helm repo update

helm install kvdi kvdi/kvdi  # yes, that's a lot of kvdi
```
