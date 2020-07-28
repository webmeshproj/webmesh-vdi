kvdi
====
A Kubernetes-Native Virtual Desktop Infrastructure

Current chart version is `0.0.8`



## Installation

```bash
$> helm repo add tinyzimmer https://tinyzimmer.github.io/kvdi/deploy/charts
$> helm install kvdi tinyzimmer/kvdi
```

Once the app pod is running (this may take a minute) you can retrieve the initial admin password with:

```bash
$> kubectl get secret kvdi-admin-secret -o go-template="{{ .data.password }}" | base64 -d && echo
```

The app service by default is called `kvdi-app` and you can retrieve the endpoint with `kubectl get svc kvdi-app`.
If you'd like to use `port-forward` you can run:

```bash
$> kubectl port-forward svc/kvdi-app 8443:443
```

Then visit https://localhost:8443 to use `kVDI`.

If you'd like to see an example of the `helm` values for using vault as the secrets backend,
you can find documentation in the [examples](../../examples/example-vault-helm-values.yaml) folder.

There is an example for LDAP authentication in the same folder.



## Chart Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| fullnameOverride | string | `""` | A full name override for resources created by the chart. |
| manager.affinity | object | `{}` | Node affinity for the manager pod. |
| manager.image.pullPolicy | string | `"IfNotPresent"` | The `ImagePullPolicy` to use for the manager pod. |
| manager.image.repository | string | `"quay.io/tinyzimmer/kvdi"` | The repository to pull the manager image from. The tag is assumed to be `manager-<chart_version>`, unless overwritten with `imageOverride`. |
| manager.image.tagOverride | string | `""` | Override the tag for the kVDI manager. Defaults to the chart version in the public repo. |
| manager.imagePullSecrets | list | `[]` | Image pull secrets for the manager pod. |
| manager.nodeSelector | object | `{}` | Node selectors for the manager pod. |
| manager.podSecurityContext | object | `{}` | The `PodSecurityContext` for the manager pod. |
| manager.replicaCount | int | `1` | The number of manager replicas to run. If more than one is set, they will run in active/standby mode. |
| manager.resources | object | `{}` | Resource limits for the manager pod. |
| manager.securityContext | object | `{}` | The container security context for the manager pod. |
| manager.tolerations | list | `[]` | Node tolerations for the manager pod. |
| nameOverride | string | `""` | A name override for resources created by the chart. |
| rbac.pspEnabled | bool | `false` | Specifies whether to create `PodSecurityPolicies` for the manager to use when booting desktops. |
| rbac.serviceAccount.create | bool | `true` | Specifies whether a `ServiceAccount` should be created. |
| rbac.serviceAccount.name | string | If not set and create is true, a name is generated using the fullname template. | The name of the `ServiceAccount` to use. |
| vdi.spec | object | The values described below are the same as the `VDICluster` CRD defaults. | The `VDICluster` spec. |
| vdi.spec.app | object | The values described below are the same as the `VDICluster` CRD defaults. | App level configurations for `kVDI`. |
| vdi.spec.app.auditLog | bool | `false` | Enables a detailed audit log of API events. At the moment, these just get logged to stdout on the app instance. |
| vdi.spec.app.corsEnabled | bool | `false` | Enables CORS headers in API responses. |
| vdi.spec.app.image | string | `quay.io/tinyzimmer/kvdi:app-${VERSION}` | The image to use for app pods. |
| vdi.spec.app.replicas | int | `1` | The number of app replicas to run. |
| vdi.spec.app.resources | object | `{}` | Resource limits for the app pods. |
| vdi.spec.appNamespace | string | `"default"` | The namespace where the `kvdi` app will run. This is different than the chart namespace. The chart lays down the manager and a VDI configuration, and the manager takes care of the rest. |
| vdi.spec.auth | object | The values described below are the same as the `VDICluster` CRD defaults. | Authentication configurations for `kVDI`. |
| vdi.spec.auth.adminSecret | string | `"kvdi-admin-secret"` | The secret to store the generated admin password in. |
| vdi.spec.auth.allowAnonymous | bool | `false` | Allow anonymous users to launch and use desktops. |
| vdi.spec.auth.localAuth | object | `{}` | Use local-auth for the authentication backend. This is currently the only supported auth provider, however more may come in the future. |
| vdi.spec.imagePullSecrets | list | `[]` | Image pull secrets to use for app containers. |
| vdi.spec.secrets.k8sSecret | object | `{"secretName":"kvdi-app-secrets"}` | Use the Kubernetes secret storage backend. This is the default if no other configuration is provided. For now, see the API reference for what to use in place of these values if using a different backend. |
| vdi.spec.secrets.k8sSecret.secretName | string | `"kvdi-app-secrets"` | The name of the Kubernetes `Secret`. backing the secret storage. |
| vdi.spec.userdataSpec | object | `{}` | If configured, enables userdata persistence with the given PVC spec. Every user will receive their own PV with the provided configuration. |
| vdi.templates | list | `[]` | Not implemented in the chart yet. This will be a place to preload desktop-templates into the cluster. |
