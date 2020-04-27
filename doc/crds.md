kVDI CRD Reference
------------------

### Packages:

-   [kvdi.io/v1alpha1](#kvdi.io%2fv1alpha1)

Types

-   [AppConfig](#%23kvdi.io%2fv1alpha1.AppConfig)
-   [AuthConfig](#%23kvdi.io%2fv1alpha1.AuthConfig)
-   [Desktop](#%23kvdi.io%2fv1alpha1.Desktop)
-   [DesktopConfig](#%23kvdi.io%2fv1alpha1.DesktopConfig)
-   [DesktopSpec](#%23kvdi.io%2fv1alpha1.DesktopSpec)
-   [DesktopTemplate](#%23kvdi.io%2fv1alpha1.DesktopTemplate)
-   [DesktopTemplateSpec](#%23kvdi.io%2fv1alpha1.DesktopTemplateSpec)
-   [LocalAuthConfig](#%23kvdi.io%2fv1alpha1.LocalAuthConfig)
-   [Resource](#%23kvdi.io%2fv1alpha1.Resource)
-   [RethinkDBConfig](#%23kvdi.io%2fv1alpha1.RethinkDBConfig)
-   [Rule](#%23kvdi.io%2fv1alpha1.Rule)
-   [VDICluster](#%23kvdi.io%2fv1alpha1.VDICluster)
-   [VDIClusterSpec](#%23kvdi.io%2fv1alpha1.VDIClusterSpec)
-   [VDIRole](#%23kvdi.io%2fv1alpha1.VDIRole)
-   [Verb](#%23kvdi.io%2fv1alpha1.Verb)

kvdi.io/v1alpha1
----------------

Package v1alpha1 contains API Schema definitions for the kvdi v1alpha1
API group

Resource Types:

### AppConfig

(*Appears on:* [VDIClusterSpec](#kvdi.io/v1alpha1.VDIClusterSpec))

AppConfig represents app configurations for the VDI cluster

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>externalHostname</code> <em>string</em></td>
<td><p>An exterenal host name that will be used for any routes that need to be broadcasted to the end user.</p></td>
</tr>
<tr class="even">
<td><code>corsEnabled</code> <em>bool</em></td>
<td><p>Whether to add CORS headers to API requests</p></td>
</tr>
<tr class="odd">
<td><code>auditLog</code> <em>bool</em></td>
<td><p>Whether to log auditing events to stdout</p></td>
</tr>
<tr class="even">
<td><code>replicas</code> <em>int32</em></td>
<td><p>The number of app replicas to run</p></td>
</tr>
<tr class="odd">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to place on the app pods</p></td>
</tr>
</tbody>
</table>

### AuthConfig

(*Appears on:* [VDIClusterSpec](#kvdi.io/v1alpha1.VDIClusterSpec))

AuthConfig will be for authentication driver configurations. The goal is
to support multiple backends, e.g. local, oauth, ldap, etc.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>allowAnonymous</code> <em>bool</em></td>
<td><p>Allow anonymous users to create desktop instances</p></td>
</tr>
<tr class="even">
<td><code>adminSecret</code> <em>string</em></td>
<td><p>A secret where a generated admin password will be stored</p></td>
</tr>
<tr class="odd">
<td><code>localAuth</code> <em><a href="#kvdi.io/v1alpha1.LocalAuthConfig">LocalAuthConfig</a></em></td>
<td><p>Use local auth (db-backed) authentication</p></td>
</tr>
</tbody>
</table>

### Desktop

Desktop is the Schema for the desktops API

<table>
<colgroup>
<col style="width: 50%" />
<col style="width: 50%" />
</colgroup>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>metadata</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta">Kubernetes meta/v1.ObjectMeta</a></em></td>
<td>Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.</td>
</tr>
<tr class="even">
<td><code>spec</code> <em><a href="#kvdi.io/v1alpha1.DesktopSpec">DesktopSpec</a></em></td>
<td><br />
<br />

<table>
<tbody>
<tr class="odd">
<td><code>vdiCluster</code> <em>string</em></td>
<td><p>The VDICluster this Desktop belongs to. This helps to determine which app instance certificates need to be created for.</p></td>
</tr>
<tr class="even">
<td><code>template</code> <em>string</em></td>
<td><p>The DesktopTemplate for booting this instance.</p></td>
</tr>
<tr class="odd">
<td><code>user</code> <em>string</em></td>
<td><p>The username to use inside the instance, defaults to <code>anonymous</code>.</p></td>
</tr>
</tbody>
</table></td>
</tr>
<tr class="odd">
<td><code>status</code> <em><a href="#kvdi.io/v1alpha1.DesktopStatus">DesktopStatus</a></em></td>
<td></td>
</tr>
</tbody>
</table>

### DesktopConfig

(*Appears on:*
[DesktopTemplateSpec](#kvdi.io/v1alpha1.DesktopTemplateSpec))

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>serviceAccount</code> <em>string</em></td>
<td><p>A service account to tie to desktops booted from this template. TODO: This should really be per-desktop and by user-grants.</p></td>
</tr>
<tr class="even">
<td><code>capabilities</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#capability-v1-core">[]Kubernetes core/v1.Capability</a></em></td>
<td><p>Extra system capabilities to add to desktops booted from this template.</p></td>
</tr>
<tr class="odd">
<td><code>enableSound</code> <em>bool</em></td>
<td><p>Whether the sound device should be mounted inside the container. Note that this also requires the image do proper setup if /dev/snd is present.</p></td>
</tr>
<tr class="even">
<td><code>allowRoot</code> <em>bool</em></td>
<td><p>AllowRoot will pass the ENABLE_ROOT envvar to the container. In the Dockerfiles in this repository, this will add the user to the sudo group and ability to sudo with no password.</p></td>
</tr>
<tr class="odd">
<td><code>socketAddr</code> <em>string</em></td>
<td><p>The address the VNC server listens on inside the image. This defaults to the UNIX socket /var/run/kvdi/vnc.sock. The novnc-proxy sidecar will forward websockify requests validated by mTLS to this socket. Must be in the format of <code>tcp://{host}:{port}</code> or <code>unix://{path}</code>.</p></td>
</tr>
</tbody>
</table>

### DesktopSpec

(*Appears on:* [Desktop](#kvdi.io/v1alpha1.Desktop))

DesktopSpec defines the desired state of Desktop

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>vdiCluster</code> <em>string</em></td>
<td><p>The VDICluster this Desktop belongs to. This helps to determine which app instance certificates need to be created for.</p></td>
</tr>
<tr class="even">
<td><code>template</code> <em>string</em></td>
<td><p>The DesktopTemplate for booting this instance.</p></td>
</tr>
<tr class="odd">
<td><code>user</code> <em>string</em></td>
<td><p>The username to use inside the instance, defaults to <code>anonymous</code>.</p></td>
</tr>
</tbody>
</table>

### DesktopTemplate

DesktopTemplate is the Schema for the desktoptemplates API

<table>
<colgroup>
<col style="width: 50%" />
<col style="width: 50%" />
</colgroup>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>metadata</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta">Kubernetes meta/v1.ObjectMeta</a></em></td>
<td>Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.</td>
</tr>
<tr class="even">
<td><code>spec</code> <em><a href="#kvdi.io/v1alpha1.DesktopTemplateSpec">DesktopTemplateSpec</a></em></td>
<td><br />
<br />

<table>
<tbody>
<tr class="odd">
<td><code>image</code> <em>string</em></td>
<td><p>The docker repository and tag to use for desktops booted from this template.</p></td>
</tr>
<tr class="even">
<td><code>imagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the container image.</p></td>
</tr>
<tr class="odd">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Any pull secrets required for pulling the container image.</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to apply to desktops booted from this template.</p></td>
</tr>
<tr class="odd">
<td><code>config</code> <em><a href="#kvdi.io/v1alpha1.DesktopConfig">DesktopConfig</a></em></td>
<td><p>Configuration options for the instances. This is highly dependant on using the Dockerfiles (or close derivitives) provided in this repository.</p></td>
</tr>
<tr class="even">
<td><code>tags</code> <em>map[string]string</em></td>
<td><p>Arbitrary tags for displaying in the app UI.</p></td>
</tr>
</tbody>
</table></td>
</tr>
<tr class="odd">
<td><code>status</code> <em><a href="#kvdi.io/v1alpha1.DesktopTemplateStatus">DesktopTemplateStatus</a></em></td>
<td></td>
</tr>
</tbody>
</table>

### DesktopTemplateSpec

(*Appears on:* [DesktopTemplate](#kvdi.io/v1alpha1.DesktopTemplate))

DesktopTemplateSpec defines the desired state of DesktopTemplate

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>image</code> <em>string</em></td>
<td><p>The docker repository and tag to use for desktops booted from this template.</p></td>
</tr>
<tr class="even">
<td><code>imagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the container image.</p></td>
</tr>
<tr class="odd">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Any pull secrets required for pulling the container image.</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to apply to desktops booted from this template.</p></td>
</tr>
<tr class="odd">
<td><code>config</code> <em><a href="#kvdi.io/v1alpha1.DesktopConfig">DesktopConfig</a></em></td>
<td><p>Configuration options for the instances. This is highly dependant on using the Dockerfiles (or close derivitives) provided in this repository.</p></td>
</tr>
<tr class="even">
<td><code>tags</code> <em>map[string]string</em></td>
<td><p>Arbitrary tags for displaying in the app UI.</p></td>
</tr>
</tbody>
</table>

### LocalAuthConfig

(*Appears on:* [AuthConfig](#kvdi.io/v1alpha1.AuthConfig))

LocalAuthConfig represents a local, db-based authentication driver.

Resource (`string` alias)

(*Appears on:* [Rule](#kvdi.io/v1alpha1.Rule))

Resource represents the target of an API action

### RethinkDBConfig

(*Appears on:* [VDIClusterSpec](#kvdi.io/v1alpha1.VDIClusterSpec))

RethinkDBConfig represents rethinkdb configurations for the VDI cluster

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>image</code> <em>string</em></td>
<td><p>The image to use for the rethinkdb instances. Defaults to rethinkdb:2.4.</p></td>
</tr>
<tr class="even">
<td><code>pvcSpec</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#persistentvolumeclaimspec-v1-core">Kubernetes core/v1.PersistentVolumeClaimSpec</a></em></td>
<td><p>The spec for persistent volumes attached to the reethinkdb nodes</p></td>
</tr>
<tr class="odd">
<td><code>shards</code> <em>int32</em></td>
<td><p>The number of shards to create for each table in the database.</p></td>
</tr>
<tr class="even">
<td><code>replicas</code> <em>int32</em></td>
<td><p>The number of data rpelicas to run for each table.</p></td>
</tr>
<tr class="odd">
<td><code>proxyReplicas</code> <em>int32</em></td>
<td><p>The number of proxy instances to run.</p></td>
</tr>
<tr class="even">
<td><code>dbResources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements for the database pods.</p></td>
</tr>
<tr class="odd">
<td><code>proxyResources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements forr the proxy pods.</p></td>
</tr>
</tbody>
</table>

### Rule

(*Appears on:* [VDIRole](#kvdi.io/v1alpha1.VDIRole))

Rule represents a set of permissions applied to a VDIRole. It mostly
resembles an rbacv1.PolicyRule, with resources being a regex and the
addition of a namespace selector. An empty rule is effectively admin
privileges.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>verbs</code> <em><a href="#kvdi.io/v1alpha1.Verb">[]Verb</a></em></td>
<td><p>The actions this rule applies for. VerbAll matches all actions.</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="#kvdi.io/v1alpha1.Resource">[]Resource</a></em></td>
<td><p>Resources this rule applies to. ResourceAll matches all resources.</p></td>
</tr>
<tr class="odd">
<td><code>resourcePatterns</code> <em>[]string</em></td>
<td><p>Resource regexes that match this rule. This can be template patterns, role names or user names. There is no All representation because * will have that effect on its own when the regex is evaluated.</p></td>
</tr>
<tr class="even">
<td><code>namespaces</code> <em>[]string</em></td>
<td><p>Namespaces this rule applies to. Only evaluated for template launching permissions. NamespaceAll matches all namespaces.</p></td>
</tr>
</tbody>
</table>

### VDICluster

VDICluster is the Schema for the vdiclusters API

<table>
<colgroup>
<col style="width: 50%" />
<col style="width: 50%" />
</colgroup>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>metadata</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta">Kubernetes meta/v1.ObjectMeta</a></em></td>
<td>Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.</td>
</tr>
<tr class="even">
<td><code>spec</code> <em><a href="#kvdi.io/v1alpha1.VDIClusterSpec">VDIClusterSpec</a></em></td>
<td><br />
<br />

<table>
<tbody>
<tr class="odd">
<td><code>appNamespace</code> <em>string</em></td>
<td><p>The namespace to provision application resurces in. Defaults to the <code>default</code> namespace</p></td>
</tr>
<tr class="even">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Pull secrets to use when pulling container images</p></td>
</tr>
<tr class="odd">
<td><code>certManagerNamespace</code> <em>string</em></td>
<td><p>The namespace cert-manager is running in. Defaults to <code>cert-manager</code>.</p></td>
</tr>
<tr class="even">
<td><code>userDataSpec</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#persistentvolumeclaimspec-v1-core">Kubernetes core/v1.PersistentVolumeClaimSpec</a></em></td>
<td><p>The configuration for user volumes. <em>NOTE:</em> Even though the controller will try to force the reclaim policy on created volumes to <code>Retain</code>, you may want to set it explicitly on your storage-class controller as an extra safeguard.</p></td>
</tr>
<tr class="odd">
<td><code>app</code> <em><a href="#kvdi.io/v1alpha1.AppConfig">AppConfig</a></em></td>
<td><p>App configurations.</p></td>
</tr>
<tr class="even">
<td><code>auth</code> <em><a href="#kvdi.io/v1alpha1.AuthConfig">AuthConfig</a></em></td>
<td><p>Authentication configurations</p></td>
</tr>
<tr class="odd">
<td><code>rethinkdb</code> <em><a href="#kvdi.io/v1alpha1.RethinkDBConfig">RethinkDBConfig</a></em></td>
<td><p>RethinkDB configurations</p></td>
</tr>
</tbody>
</table></td>
</tr>
<tr class="odd">
<td><code>status</code> <em><a href="#kvdi.io/v1alpha1.VDIClusterStatus">VDIClusterStatus</a></em></td>
<td></td>
</tr>
</tbody>
</table>

### VDIClusterSpec

(*Appears on:* [VDICluster](#kvdi.io/v1alpha1.VDICluster))

VDIClusterSpec defines the desired state of VDICluster

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>appNamespace</code> <em>string</em></td>
<td><p>The namespace to provision application resurces in. Defaults to the <code>default</code> namespace</p></td>
</tr>
<tr class="even">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Pull secrets to use when pulling container images</p></td>
</tr>
<tr class="odd">
<td><code>certManagerNamespace</code> <em>string</em></td>
<td><p>The namespace cert-manager is running in. Defaults to <code>cert-manager</code>.</p></td>
</tr>
<tr class="even">
<td><code>userDataSpec</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#persistentvolumeclaimspec-v1-core">Kubernetes core/v1.PersistentVolumeClaimSpec</a></em></td>
<td><p>The configuration for user volumes. <em>NOTE:</em> Even though the controller will try to force the reclaim policy on created volumes to <code>Retain</code>, you may want to set it explicitly on your storage-class controller as an extra safeguard.</p></td>
</tr>
<tr class="odd">
<td><code>app</code> <em><a href="#kvdi.io/v1alpha1.AppConfig">AppConfig</a></em></td>
<td><p>App configurations.</p></td>
</tr>
<tr class="even">
<td><code>auth</code> <em><a href="#kvdi.io/v1alpha1.AuthConfig">AuthConfig</a></em></td>
<td><p>Authentication configurations</p></td>
</tr>
<tr class="odd">
<td><code>rethinkdb</code> <em><a href="#kvdi.io/v1alpha1.RethinkDBConfig">RethinkDBConfig</a></em></td>
<td><p>RethinkDB configurations</p></td>
</tr>
</tbody>
</table>

### VDIRole

VDIRole is the Schema for the vdiroles API

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>metadata</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.17/#objectmeta-v1-meta">Kubernetes meta/v1.ObjectMeta</a></em></td>
<td>Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.</td>
</tr>
<tr class="even">
<td><code>rules</code> <em><a href="#kvdi.io/v1alpha1.Rule">[]Rule</a></em></td>
<td><p>A list of rules granting access to resources in the VDICluster.</p></td>
</tr>
</tbody>
</table>

Verb (`string` alias)

(*Appears on:* [Rule](#kvdi.io/v1alpha1.Rule))

Verb represents an API action

------------------------------------------------------------------------

*Generated with `gen-crd-api-reference-docs` on git commit `3c34300`.*
