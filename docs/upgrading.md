# Upgrading

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