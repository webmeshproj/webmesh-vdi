## kVDI CRD Reference

### Packages:

-   [app.kvdi.io/v1](#app.kvdi.io%2fv1)

Types

-   [AppConfig](#AppConfig)
-   [AuthConfig](#AuthConfig)
-   [DesktopsConfig](#DesktopsConfig)
-   [GrafanaConfig](#GrafanaConfig)
-   [K8SSecretConfig](#K8SSecretConfig)
-   [LDAPConfig](#LDAPConfig)
-   [LocalAuthConfig](#LocalAuthConfig)
-   [MetricsConfig](#MetricsConfig)
-   [OIDCConfig](#OIDCConfig)
-   [PrometheusConfig](#PrometheusConfig)
-   [SecretsConfig](#SecretsConfig)
-   [ServiceMonitorConfig](#ServiceMonitorConfig)
-   [TLSConfig](#TLSConfig)
-   [VDICluster](#VDICluster)
-   [VDIClusterSpec](#VDIClusterSpec)
-   [VaultConfig](#VaultConfig)

## app.kvdi.io/v1

Package v1 contains API Schema definitions for the App v1 API group

Resource Types:

### AppConfig

(*Appears on:* [VDIClusterSpec](#VDIClusterSpec))

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
<td><code>image</code> <em>string</em></td>
<td><p>The image to use for the app instances. Defaults to the public image matching the version of the currently running manager.</p></td>
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
<td><code>serviceType</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#servicetype-v1-core">Kubernetes core/v1.ServiceType</a></em></td>
<td><p>The type of service to create in front of the app instance. Defaults to <code>LoadBalancer</code>.</p></td>
</tr>
<tr class="even">
<td><code>serviceAnnotations</code> <em>map[string]string</em></td>
<td><p>Extra annotations to apply to the app service.</p></td>
</tr>
<tr class="odd">
<td><code>tls</code> <em><a href="#TLSConfig">TLSConfig</a></em></td>
<td><p>TLS configurations for the app instance</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to place on the app pods</p></td>
</tr>
</tbody>
</table>

### AuthConfig

(*Appears on:* [VDIClusterSpec](#VDIClusterSpec))

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
<td><code>tokenDuration</code> <em>string</em></td>
<td><p>How long issued access tokens should be valid for. When using OIDC auth you may want to set this to a higher value (e.g. 8-10h) since the refresh token flow will not be able to lookup a user’s grants from the provider. Defaults to <code>15m</code>.</p></td>
</tr>
<tr class="even">
<td><code>defaultRoleRules</code> <em><a href="rbacv1.md#Rule">[]rbacv1.Rule</a></em></td>
<td><p>The rules to apply to the default role created for this cluster. These are the rules applied to anonymous users (if allowed) and non-grouped OIDC users. They can also be used for convenience when getting started. The defaults only allow for launching templates in the <code>appNamespace</code>.</p></td>
</tr>
<tr class="odd">
<td><code>localAuth</code> <em><a href="#LocalAuthConfig">LocalAuthConfig</a></em></td>
<td><p>Use local auth (secret-backed) authentication</p></td>
</tr>
<tr class="even">
<td><code>ldapAuth</code> <em><a href="#LDAPConfig">LDAPConfig</a></em></td>
<td><p>Use LDAP for authentication.</p></td>
</tr>
<tr class="odd">
<td><code>oidcAuth</code> <em><a href="#OIDCConfig">OIDCConfig</a></em></td>
<td><p>Use OIDC for authentication</p></td>
</tr>
</tbody>
</table>

### DesktopsConfig

(*Appears on:* [VDIClusterSpec](#VDIClusterSpec))

DesktopsConfig represents global configurations for desktop sessions.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>maxSessionLength</code> <em>string</em></td>
<td><p>When configured, desktop sessions will be forcefully terminated when the time limit is reached.</p></td>
</tr>
<tr class="even">
<td><code>sessionsPerUser</code> <em>int</em></td>
<td><p>The maximum number of sessions a user can run at a time. A zero value (or undefined) means no limit. When using a <code>userdataSpec</code>, you might want to set this value to 1 if you aren’t using ReadWriteMany volumes. The storage controller would inevitably enforce this behavior anyway, but you would save the <code>kvdi-manager</code> some extra work.</p></td>
</tr>
</tbody>
</table>

### GrafanaConfig

(*Appears on:* [MetricsConfig](#MetricsConfig))

GrafanaConfig contains configuration options for the grafana sidecar.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>enabled</code> <em>bool</em></td>
<td><p>Set to true to run a grafana sidecar with the app pods. This can be used to visualize data in the prometheus deployment.</p></td>
</tr>
</tbody>
</table>

### K8SSecretConfig

(*Appears on:* [SecretsConfig](#SecretsConfig))

K8SSecretConfig uses a Kubernetes secret to store and retrieve sensitive
values.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>secretName</code> <em>string</em></td>
<td><p>The name of the secret backing the values. Default is <code>&lt;cluster-name&gt;-app-secrets</code>.</p></td>
</tr>
</tbody>
</table>

### LDAPConfig

(*Appears on:* [AuthConfig](#AuthConfig))

LDAPConfig represents the configurations for using LDAP as the
authentication backend.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>url</code> <em>string</em></td>
<td><p>The URL to the LDAP server.</p></td>
</tr>
<tr class="even">
<td><code>tlsInsecureSkipVerify</code> <em>bool</em></td>
<td><p>Set to true to skip TLS verification of an <code>ldaps</code> connection.</p></td>
</tr>
<tr class="odd">
<td><code>tlsCACert</code> <em>string</em></td>
<td><p>The base64 encoded CA certificate to use when verifying the TLS certificate of the LDAP server.</p></td>
</tr>
<tr class="even">
<td><code>bindUserDNSecretKey</code> <em>string</em></td>
<td><p>If you want to use the built-in secrets backend (vault or k8s currently), set this to either the name of the secret in the vault path (the key must be “data” for now), or the key of the secret used in <code>secrets.k8sSecret.secretName</code>. In default configurations this is <code>kvdi-app-secrets</code>. Defaults to <code>ldap-userdn</code>.</p></td>
</tr>
<tr class="odd">
<td><code>bindPasswordSecretKey</code> <em>string</em></td>
<td><p>Similar to the <code>bindUserDNSecretKey</code>, but for the location of the password secret. Defaults to <code>ldap-password</code>.</p></td>
</tr>
<tr class="even">
<td><code>bindCredentialsSecret</code> <em>string</em></td>
<td><p>If you’d rather create a separate k8s secret (instead of the configured backend) for the LDAP credentials, set its name here. The keys in the secret need to be defined in the other fields still. Default is to use the secret backend.</p></td>
</tr>
<tr class="odd">
<td><code>adminGroups</code> <em>[]string</em></td>
<td><p>Group DNs that are allowed administrator access to the cluster. Kubernetes admins will still have the ability to change configurations via the CRDs.</p></td>
</tr>
<tr class="even">
<td><code>userSearchBase</code> <em>string</em></td>
<td><p>The base scope to search for users in. Default is to search the entire directory.</p></td>
</tr>
<tr class="odd">
<td><code>userIDAttribute</code> <em>string</em></td>
<td><p>The user ID attribute to use when looking up a provided username. Defaults to <code>uid</code>. This value may be different depending on the LDAP provider. For example, in an Active Directory environment you may want to set this value to <code>sAMAccountName</code>.</p></td>
</tr>
<tr class="even">
<td><code>userGroupsAttribute</code> <em>string</em></td>
<td><p>The user attribute use to lookup group membership in LDAP. Defaults to <code>memberOf</code>.</p></td>
</tr>
<tr class="odd">
<td><code>userStatusAttribute</code> <em>string</em></td>
<td><p>The user attribute to use when querying if an account is active. Defaults to <code>accountStatus</code>. Only takes effect if <code>doStatusCheck</code> is <code>true</code>. A user is considered disabled when the attribute is both present and matches the value in <code>userStatusDisabledValue</code>.</p></td>
</tr>
<tr class="even">
<td><code>userStatusDisabledValue</code> <em>string</em></td>
<td><p>The value for the <code>userStatusAttribute</code> that signifies that the user is disabled. Defaults to <code>inactive</code>.</p></td>
</tr>
<tr class="odd">
<td><code>doStatusCheck</code> <em>bool</em></td>
<td><p>When set to true, the authentication provider will query the user’s attributes for the <code>userStatusAttribute</code> and make sure it matches the value in <code>userStatusEnabledValue</code> before attemtping to bind.</p></td>
</tr>
</tbody>
</table>

### LocalAuthConfig

(*Appears on:* [AuthConfig](#AuthConfig))

LocalAuthConfig represents a local, ‘passwd’-like authentication driver.

### MetricsConfig

(*Appears on:* [VDIClusterSpec](#VDIClusterSpec))

MetricsConfig contains configuration options for gathering metrics.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>serviceMonitor</code> <em><a href="#ServiceMonitorConfig">ServiceMonitorConfig</a></em></td>
<td><p>Configurations for creating a ServiceMonitor CR for a pre-existing prometheus-operator installation.</p></td>
</tr>
<tr class="even">
<td><code>prometheus</code> <em><a href="#PrometheusConfig">PrometheusConfig</a></em></td>
<td><p>Prometheus deployment configurations.g.</p></td>
</tr>
<tr class="odd">
<td><code>grafana</code> <em><a href="#GrafanaConfig">GrafanaConfig</a></em></td>
<td><p>Grafana sidecar configurations.</p></td>
</tr>
</tbody>
</table>

### OIDCConfig

(*Appears on:* [AuthConfig](#AuthConfig))

OIDCConfig represents configurations for using an OIDC/OAuth provider
for authentication.

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
<td><code>issuerURL</code> <em>string</em></td>
<td><p>The OIDC issuer URL used for discovery</p></td>
</tr>
<tr class="even">
<td><code>clientIDKey</code> <em>string</em></td>
<td><p>When using the built-in secrets backend, the key to where the client-id is stored. Set this to either the name of the secret in the vault path (the key must be “data” for now), or the key of the secret used in <code>secrets.k8sSecret.secretName</code>. When configuring <code>clientCredentialsSecret</code>, set this to the key in that secret. Defaults to <code>oidc-clientid</code>.</p></td>
</tr>
<tr class="odd">
<td><code>clientSecretKey</code> <em>string</em></td>
<td><p>Similar to <code>clientIDKey</code>, but for the location of the client secret. Defaults to <code>oidc-clientsecret</code>.</p></td>
</tr>
<tr class="even">
<td><code>clientCredentialsSecret</code> <em>string</em></td>
<td><p>When creating your own kubernets secret with the <code>clientIDKey</code> and <code>clientSecretKey</code>, set this to the name of the created secret. It must be in the same namespace as the manager and app instances. Defaults to <code>oidc-clientsecret</code>.</p></td>
</tr>
<tr class="odd">
<td><code>redirectURL</code> <em>string</em></td>
<td><p>The redirect URL path configured in the OIDC provider. This should be the full path where kvdi is hosted followed by <code>/api/login</code>. For example, if <code>kvdi</code> is hosted at <a href="https://kvdi.local">https://kvdi.local</a>, then this value should be set <code>https://kvdi.local/api/login</code>.</p></td>
</tr>
<tr class="even">
<td><code>scopes</code> <em>[]string</em></td>
<td><p>The scopes to request with the authentication request. Defaults to <code>["openid", "email", "profile", "groups"]</code>.</p></td>
</tr>
<tr class="odd">
<td><code>groupScope</code> <em>string</em></td>
<td><p>If your OIDC provider does not return a <code>groups</code> object, set this to the user attribute to use for binding authenticated users to VDIRoles. Defaults to <code>groups</code>.</p></td>
</tr>
<tr class="even">
<td><code>adminGroups</code> <em>[]string</em></td>
<td><p>Groups that are allowed administrator access to the cluster. Kubernetes admins will still have the ability to change rbac configurations via the CRDs.</p></td>
</tr>
<tr class="odd">
<td><code>tlsInsecureSkipVerify</code> <em>bool</em></td>
<td><p>Set to true to skip TLS verification of an OIDC provider.</p></td>
</tr>
<tr class="even">
<td><code>tlsCACert</code> <em>string</em></td>
<td><p>The base64 encoded CA certificate to use when verifying the TLS certificate of the OIDC provider.</p></td>
</tr>
<tr class="odd">
<td><code>allowNonGroupedReadOnly</code> <em>bool</em></td>
<td><p>Set to true if the OIDC provider does not support the “groups” claim (or any valid alternative) and/or you would like to allow any authenticated user read-only access.</p></td>
</tr>
<tr class="even">
<td><code>preserveTokens</code> <em>bool</em></td>
<td><p>The access tokens returned by the OIDC provider are usually discarded after identify information is retrieved from them. If you set this to true, these fields will be available for mapping in desktops at the following paths:</p>
<ul>
<li><code>{{ .Session.Data.access_token }}</code></li>
<li><code>{{ .Session.Data.token_type }}</code></li>
<li><code>{{ .Session.Data.refresh_token }}</code></li>
<li><code>{{ .Session.Data.expiry }}</code></li>
</ul>
<p><strong>NOTE:</strong> This should be considered an insecure option and only turned on taking into account the inherent risks. If the access token used for authorizing actions against the kvdi API gets compromised, it would be relatively easy for the attacker to extract this information from the token and use it for authenticating against third-party resources. Additionally, when mapping these values to desktops, they will be stored temporarily in Kubernetes Secrets. The security of those secrets depends highly on your Kubernetes RBAC setup and who has access to secrets in the namespace where the Desktop is. So in short, it would be wise to only use this setting in trusted environments where access to the necessary kubernetes APIs is only available to a select group of administrators, and the risk of the user using a compromised browser is minimal.</p></td>
</tr>
</tbody>
</table>

### PrometheusConfig

(*Appears on:* [MetricsConfig](#MetricsConfig))

PrometheusConfig contains configuration options for a prometheus
deployment.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>create</code> <em>bool</em></td>
<td><p>Set to true to create a prometheus instance.</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to place on the Prometheus deployment</p></td>
</tr>
</tbody>
</table>

### SecretsConfig

(*Appears on:* [VDIClusterSpec](#VDIClusterSpec))

SecretsConfig configurese the backend for secrets management.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>k8sSecret</code> <em><a href="#K8SSecretConfig">K8SSecretConfig</a></em></td>
<td><p>Use a kubernetes secret for storing sensitive values. If no other coniguration is provided then this is the fallback.</p></td>
</tr>
<tr class="even">
<td><code>vault</code> <em><a href="#VaultConfig">VaultConfig</a></em></td>
<td><p>Use vault for storing sensitive values. Requires kubernetes service account authentication.</p></td>
</tr>
</tbody>
</table>

### ServiceMonitorConfig

(*Appears on:* [MetricsConfig](#MetricsConfig))

ServiceMonitorConfig contains configuration options for creating a
ServiceMonitor.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>create</code> <em>bool</em></td>
<td><p>Set to true to create a ServiceMonitor object for the kvdi metrics.</p></td>
</tr>
<tr class="even">
<td><code>labels</code> <em>map[string]string</em></td>
<td><p>Extra labels to apply to the ServiceMonitor object. Set these to the selector in your prometheus-operator configuration (usually <code>{"release": "&lt;helm_release_name&gt;"}</code>). Defaults to <code>{"release": "prometheus"}</code>.</p></td>
</tr>
</tbody>
</table>

### TLSConfig

(*Appears on:* [AppConfig](#AppConfig))

TLSConfig contains TLS configurations for kVDI.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>serverSecret</code> <em>string</em></td>
<td><p>A pre-existing TLS secret to use for the HTTPS listener. If not defined, a certificate is generated.</p></td>
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
<td><code>metadata</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">Kubernetes meta/v1.ObjectMeta</a></em></td>
<td>Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.</td>
</tr>
<tr class="even">
<td><code>spec</code> <em><a href="#VDIClusterSpec">VDIClusterSpec</a></em></td>
<td><br />
<br />

<table>
<tbody>
<tr class="odd">
<td><code>appNamespace</code> <em>string</em></td>
<td><p>The namespace to provision application resurces in. Defaults to the <code>default</code> namespace</p></td>
</tr>
<tr class="even">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Pull secrets to use when pulling container images</p></td>
</tr>
<tr class="odd">
<td><code>userdataSpec</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#persistentvolumeclaimspec-v1-core">Kubernetes core/v1.PersistentVolumeClaimSpec</a></em></td>
<td><p>The configuration for user volumes. <strong>NOTE:</strong> Even though the controller will try to force the reclaim policy on created volumes to <code>Retain</code>, you may want to set it explicitly on your storage-class controller as an extra safeguard.</p></td>
</tr>
<tr class="even">
<td><code>app</code> <em><a href="#AppConfig">AppConfig</a></em></td>
<td><p>App configurations.</p></td>
</tr>
<tr class="odd">
<td><code>auth</code> <em><a href="#AuthConfig">AuthConfig</a></em></td>
<td><p>Authentication configurations</p></td>
</tr>
<tr class="even">
<td><code>desktops</code> <em><a href="#DesktopsConfig">DesktopsConfig</a></em></td>
<td><p>Global desktop configurations</p></td>
</tr>
<tr class="odd">
<td><code>secrets</code> <em><a href="#SecretsConfig">SecretsConfig</a></em></td>
<td><p>Secrets backend configurations</p></td>
</tr>
<tr class="even">
<td><code>metrics</code> <em><a href="#MetricsConfig">MetricsConfig</a></em></td>
<td><p>Metrics configurations.</p></td>
</tr>
</tbody>
</table></td>
</tr>
<tr class="odd">
<td><code>status</code> <em><a href="#VDIClusterStatus">VDIClusterStatus</a></em></td>
<td></td>
</tr>
</tbody>
</table>

### VDIClusterSpec

(*Appears on:* [VDICluster](#VDICluster))

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
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Pull secrets to use when pulling container images</p></td>
</tr>
<tr class="odd">
<td><code>userdataSpec</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#persistentvolumeclaimspec-v1-core">Kubernetes core/v1.PersistentVolumeClaimSpec</a></em></td>
<td><p>The configuration for user volumes. <strong>NOTE:</strong> Even though the controller will try to force the reclaim policy on created volumes to <code>Retain</code>, you may want to set it explicitly on your storage-class controller as an extra safeguard.</p></td>
</tr>
<tr class="even">
<td><code>app</code> <em><a href="#AppConfig">AppConfig</a></em></td>
<td><p>App configurations.</p></td>
</tr>
<tr class="odd">
<td><code>auth</code> <em><a href="#AuthConfig">AuthConfig</a></em></td>
<td><p>Authentication configurations</p></td>
</tr>
<tr class="even">
<td><code>desktops</code> <em><a href="#DesktopsConfig">DesktopsConfig</a></em></td>
<td><p>Global desktop configurations</p></td>
</tr>
<tr class="odd">
<td><code>secrets</code> <em><a href="#SecretsConfig">SecretsConfig</a></em></td>
<td><p>Secrets backend configurations</p></td>
</tr>
<tr class="even">
<td><code>metrics</code> <em><a href="#MetricsConfig">MetricsConfig</a></em></td>
<td><p>Metrics configurations.</p></td>
</tr>
</tbody>
</table>

### VaultConfig

(*Appears on:* [SecretsConfig](#SecretsConfig))

VaultConfig represents the configurations for connecting to a vault
server.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>address</code> <em>string</em></td>
<td><p>The full URL to the vault server. Same as the <code>VAULT_ADDR</code> variable.</p></td>
</tr>
<tr class="even">
<td><code>caCertBase64</code> <em>string</em></td>
<td><p>The base64 encoded CA certificate for verifying the vault server certificate.</p></td>
</tr>
<tr class="odd">
<td><code>insecure</code> <em>bool</em></td>
<td><p>Set to true to disable TLS verification.</p></td>
</tr>
<tr class="even">
<td><code>tlsServerName</code> <em>string</em></td>
<td><p>Optionally set the SNI when connecting using HTTPS.</p></td>
</tr>
<tr class="odd">
<td><code>authRole</code> <em>string</em></td>
<td><p>The auth role to assume when authenticating against vault. Defaults to <code>kvdi</code>.</p></td>
</tr>
<tr class="even">
<td><code>secretsPath</code> <em>string</em></td>
<td><p>The base path to store secrets in vault. “Keys” for other configurations in the context of the vault backend can be put at <code>&lt;secretsPath&gt;/&lt;secretKey&gt;.data</code>. This will change in the future to support keys inside the secret itself, instead of assuming <code>data</code>.</p></td>
</tr>
</tbody>
</table>

------------------------------------------------------------------------

*Generated with `gen-crd-api-reference-docs` on git commit `b6bee99`.*
