kVDI CRD Reference
------------------

### Packages:

-   [kvdi.io/v1alpha1](#kvdi.io%2fv1alpha1)

Types

-   [AppConfig](#AppConfig)
-   [AuthConfig](#AuthConfig)
-   [AuthProvider](#AuthProvider)
-   [AuthResult](#AuthResult)
-   [AuthorizeRequest](#AuthorizeRequest)
-   [CreateRoleRequest](#CreateRoleRequest)
-   [CreateSessionRequest](#CreateSessionRequest)
-   [CreateUserRequest](#CreateUserRequest)
-   [Desktop](#Desktop)
-   [DesktopConfig](#DesktopConfig)
-   [DesktopInit](#DesktopInit)
-   [DesktopSpec](#DesktopSpec)
-   [DesktopTemplate](#DesktopTemplate)
-   [DesktopTemplateSpec](#DesktopTemplateSpec)
-   [JWTClaims](#JWTClaims)
-   [K8SSecretConfig](#K8SSecretConfig)
-   [LDAPConfig](#LDAPConfig)
-   [LocalAuthConfig](#LocalAuthConfig)
-   [LoginRequest](#LoginRequest)
-   [OIDCConfig](#OIDCConfig)
-   [Resource](#Resource)
-   [ResourceGetter](#ResourceGetter)
-   [RolesGetter](#RolesGetter)
-   [Rule](#Rule)
-   [SecretsConfig](#SecretsConfig)
-   [SecretsProvider](#SecretsProvider)
-   [SessionResponse](#SessionResponse)
-   [TemplatesGetter](#TemplatesGetter)
-   [UpdateMFARequest](#UpdateMFARequest)
-   [UpdateMFAResponse](#UpdateMFAResponse)
-   [UpdateRoleRequest](#UpdateRoleRequest)
-   [UpdateUserRequest](#UpdateUserRequest)
-   [UsersGetter](#UsersGetter)
-   [VDICluster](#VDICluster)
-   [VDIClusterSpec](#VDIClusterSpec)
-   [VDIRole](#VDIRole)
-   [VDIUser](#VDIUser)
-   [VDIUserRole](#VDIUserRole)
-   [VaultConfig](#VaultConfig)
-   [Verb](#Verb)

kvdi.io/v1alpha1
----------------

Package v1alpha1 contains API Schema definitions for the kvdi v1alpha1
API group

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

### AuthProvider

AuthProvider defines an interface for handling login attempts. Currently
only local auth (using the secrets backend) is supported, however other
integrations such as LDAP or OAuth can implement this interface.

### AuthResult

AuthResult represents a response from an authentication attempt to a
provider. It contains user information, roles, and any other auth
requirements.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>User</code> <em><a href="#VDIUser">VDIUser</a></em></td>
<td><p>The authenticated user and their roles</p></td>
</tr>
<tr class="even">
<td><code>RedirectURL</code> <em>string</em></td>
<td><p>The provider can populate this field to signify a redirect is required, e.g. for OIDC.</p></td>
</tr>
</tbody>
</table>

### AuthorizeRequest

AuthorizeRequest is a request with an OTP for receiving an authorized
token.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>otp</code> <em>string</em></td>
<td><p>The one-time password</p></td>
</tr>
</tbody>
</table>

### CreateRoleRequest

CreateRoleRequest represents a request for a new role.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>name</code> <em>string</em></td>
<td><p>The name of the new role</p></td>
</tr>
<tr class="even">
<td><code>annotations</code> <em>map[string]string</em></td>
<td><p>Annotations to apply to the role</p></td>
</tr>
<tr class="odd">
<td><code>rules</code> <em><a href="#Rule">[]Rule</a></em></td>
<td><p>Rules to apply to the new role.</p></td>
</tr>
</tbody>
</table>

### CreateSessionRequest

CreateSessionRequest requests a new desktop session with the givin
parameters.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>template</code> <em>string</em></td>
<td><p>The template to create the session from.</p></td>
</tr>
<tr class="even">
<td><code>namespace</code> <em>string</em></td>
<td><p>The namespace to launch the template in. Defaults to default.</p></td>
</tr>
</tbody>
</table>

### CreateUserRequest

CreateUserRequest represents a request to create a new user. Not all
auth providers will be able to implement this route and can instead
return an error describing why.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>username</code> <em>string</em></td>
<td><p>The user name for the new user.</p></td>
</tr>
<tr class="even">
<td><code>password</code> <em>string</em></td>
<td><p>The password for the new user.</p></td>
</tr>
<tr class="odd">
<td><code>roles</code> <em>[]string</em></td>
<td><p>Roles to assign the new user. These are the names of VDIRoles in the cluster.</p></td>
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
<td><code>metadata</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">Kubernetes meta/v1.ObjectMeta</a></em></td>
<td>Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.</td>
</tr>
<tr class="even">
<td><code>spec</code> <em><a href="#DesktopSpec">DesktopSpec</a></em></td>
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
<td><code>status</code> <em><a href="#DesktopStatus">DesktopStatus</a></em></td>
<td></td>
</tr>
</tbody>
</table>

### DesktopConfig

(*Appears on:*
[DesktopTemplateSpec](#DesktopTemplateSpec))

DesktopConfig represents configurations for the template and desktops
booted from it.

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
<td><code>capabilities</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#capability-v1-core">[]Kubernetes core/v1.Capability</a></em></td>
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
<tr class="even">
<td><code>proxyImage</code> <em>string</em></td>
<td><p>The image to use for the sidecar that proxies mTLS connections to the local VNC server inside the Desktop. Defaults to the public novnc-proxy image matching the version of the currrently running manager.</p></td>
</tr>
<tr class="odd">
<td><code>init</code> <em><a href="#DesktopInit">DesktopInit</a></em></td>
<td><p>The type of init system inside the image, currently only supervisord and systemd are supported. Defaults to <code>supervisord</code> (but depending on how much I like systemd in this use case, that could change).</p></td>
</tr>
</tbody>
</table>

DesktopInit (`string` alias)

(*Appears on:* [DesktopConfig](#DesktopConfig))

DesktopInit represents the init system that the desktop container uses.

### DesktopSpec

(*Appears on:* [Desktop](#Desktop))

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
<td><code>metadata</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">Kubernetes meta/v1.ObjectMeta</a></em></td>
<td>Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.</td>
</tr>
<tr class="even">
<td><code>spec</code> <em><a href="#DesktopTemplateSpec">DesktopTemplateSpec</a></em></td>
<td><br />
<br />

<table>
<tbody>
<tr class="odd">
<td><code>image</code> <em>string</em></td>
<td><p>The docker repository and tag to use for desktops booted from this template.</p></td>
</tr>
<tr class="even">
<td><code>imagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the container image.</p></td>
</tr>
<tr class="odd">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Any pull secrets required for pulling the container image.</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to apply to desktops booted from this template.</p></td>
</tr>
<tr class="odd">
<td><code>config</code> <em><a href="#DesktopConfig">DesktopConfig</a></em></td>
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
<td><code>status</code> <em><a href="#DesktopTemplateStatus">DesktopTemplateStatus</a></em></td>
<td></td>
</tr>
</tbody>
</table>

### DesktopTemplateSpec

(*Appears on:* [DesktopTemplate](#DesktopTemplate))

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
<td><code>imagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the container image.</p></td>
</tr>
<tr class="odd">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Any pull secrets required for pulling the container image.</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to apply to desktops booted from this template.</p></td>
</tr>
<tr class="odd">
<td><code>config</code> <em><a href="#DesktopConfig">DesktopConfig</a></em></td>
<td><p>Configuration options for the instances. This is highly dependant on using the Dockerfiles (or close derivitives) provided in this repository.</p></td>
</tr>
<tr class="even">
<td><code>tags</code> <em>map[string]string</em></td>
<td><p>Arbitrary tags for displaying in the app UI.</p></td>
</tr>
</tbody>
</table>

### JWTClaims

JWTClaims represents the claims used when issuing JWT tokens.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>user</code> <em><a href="#VDIUser">VDIUser</a></em></td>
<td><p>The user with their permissions when the token was generated</p></td>
</tr>
<tr class="even">
<td><code>authorized</code> <em>bool</em></td>
<td><p>Whether the user is fully authorized</p></td>
</tr>
<tr class="odd">
<td><code>StandardClaims</code> <em>github.com/dgrijalva/jwt-go.StandardClaims</em></td>
<td><p>The standard JWT claims</p></td>
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
<td><p>If you want to use the built-in secrets backend (vault or k8s currently), set this to either the name of the secret in the vault path, or the key of the secret used in <code>k8sSecret.secretName. In default configurations this is</code> <code>kvdi-app-secrets</code>. Defaults to <code>ldap-userdn</code>.</p></td>
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
</tbody>
</table>

### LocalAuthConfig

(*Appears on:* [AuthConfig](#AuthConfig))

LocalAuthConfig represents a local, ‘passwd’-like authentication driver.

### LoginRequest

LoginRequest represents a request for a session token. Different auth
providers may not always need this request, and can instead redirect
/api/login as needed. All the auth provider needs to do in the end is
return a JWT token that contains a fulfilled VDIUser.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>username</code> <em>string</em></td>
<td><p>Username</p></td>
</tr>
<tr class="even">
<td><code>password</code> <em>string</em></td>
<td><p>Password</p></td>
</tr>
<tr class="odd">
<td><code>state</code> <em>string</em></td>
<td><p>State generated by requesting client to prevent CSRF and retrieve tokens from an oidc flow</p></td>
</tr>
<tr class="even">
<td><code>request</code> <em>net/http.Request</em></td>
<td><p>the underlying request object for usage by auth providers</p></td>
</tr>
</tbody>
</table>

### OIDCConfig

(*Appears on:* [AuthConfig](#AuthConfig))

OIDCConfig represents configurations for using an OIDC/OAuth provider
for authentication.

<table>
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
<td><p>When using the built-in secrets backend, the key to where the client-id is stored. When configuring <code>clientCredentialsSecret</code>, set this to the key in that secret. Defaults to <code>oidc-clientid</code>.</p></td>
</tr>
<tr class="odd">
<td><code>clientSecretKey</code> <em>string</em></td>
<td><p>Similar to <code>clientIDKey</code>, but for the location of the client secret. Defaults to <code>oidc-clientsecret</code>.</p></td>
</tr>
<tr class="even">
<td><code>clientCredentialsSecret</code> <em>string</em></td>
<td><p>When creating your own kubernets secret with the <code>clientIDKey</code> and <code>clientSecretKey</code>, set this to the name of the created secret. It must be in the same namespace as the manager and app instances.</p></td>
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
</tbody>
</table>

Resource (`string` alias)

(*Appears on:* [Rule](#Rule))

Resource represents the target of an API action

### ResourceGetter

ResourceGetter is an interface for retrieving lists of kVDI related
resources. Its primary purpose is to pass an interface to rbac
evaluations so they can check permissions against present resources.

### RolesGetter

RolesGetter is an interface that can be used to retrieve available roles
while checking user permissions.

### Rule

(*Appears on:* [CreateRoleRequest](#CreateRoleRequest),
[UpdateRoleRequest](#UpdateRoleRequest),
[VDIRole](#VDIRole),
[VDIUserRole](#VDIUserRole))

Rule represents a set of permissions applied to a VDIRole. It mostly
resembles an rbacv1.PolicyRule, with resources being a regex and the
addition of a namespace selector.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>verbs</code> <em><a href="#Verb">[]Verb</a></em></td>
<td><p>The actions this rule applies for. VerbAll matches all actions.</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="#Resource">[]Resource</a></em></td>
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

### SecretsProvider

SecretsProvider provides an interface for an app instance to get and
store any secrets it needs. Currenetly there is only a k8s secret
provider, but this intreface could be implemented for things like vault.

### SessionResponse

SessionResponse represents a response with a new session token

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>token</code> <em>string</em></td>
<td><p>The X-Session-Token to use for future requests.</p></td>
</tr>
<tr class="even">
<td><code>expiresAt</code> <em>int64</em></td>
<td><p>The time the token expires.</p></td>
</tr>
<tr class="odd">
<td><code>user</code> <em><a href="#VDIUser">VDIUser</a></em></td>
<td><p>Information about the authenticated user and their permissions.</p></td>
</tr>
<tr class="even">
<td><code>authorized</code> <em>bool</em></td>
<td><p>Whether the user is fully authorized (e.g. false if MFA is required but not provided yet)</p></td>
</tr>
</tbody>
</table>

### TemplatesGetter

TemplatesGetter is an interface that can be used to retrieve available
templates while checking user permissions.

### UpdateMFARequest

UpdateMFARequest sets the MFA configuration for the user. If enabling, a
provisioning URI will be returned.

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
<td><p>When set, will enable MFA for the given user. If false, will disable MFA.</p></td>
</tr>
</tbody>
</table>

### UpdateMFAResponse

UpdateMFAResponse contains the response to an UpdateMFARequest.

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
<td><p>Whether MFA is enabled for the user</p></td>
</tr>
<tr class="even">
<td><code>provisioningURI</code> <em>string</em></td>
<td><p>If enabled is set, a provisioning URI is also returned.</p></td>
</tr>
</tbody>
</table>

### UpdateRoleRequest

UpdateRoleRequest requests updates to an existing role. The existing
attributes will be entirely replaced with those supplied in the payload.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>annotations</code> <em>map[string]string</em></td>
<td><p>The new annotations for the role</p></td>
</tr>
<tr class="even">
<td><code>rules</code> <em><a href="#Rule">[]Rule</a></em></td>
<td><p>The new rules for the role.</p></td>
</tr>
</tbody>
</table>

### UpdateUserRequest

UpdateUserRequest requests updates to an existing user. Not all auth
providers will be able to implement this route and can instead return an
error describing why.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>password</code> <em>string</em></td>
<td><p>When populated, will change the password for the user.</p></td>
</tr>
<tr class="even">
<td><code>roles</code> <em>[]string</em></td>
<td><p>When populated will change the roles for the user.</p></td>
</tr>
</tbody>
</table>

### UsersGetter

UsersGetter is an interface that can be used to retrieve available users
while checking user permissions.

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
<td><code>certManagerNamespace</code> <em>string</em></td>
<td><p>The namespace cert-manager is running in. Defaults to <code>cert-manager</code>.</p></td>
</tr>
<tr class="even">
<td><code>userdataSpec</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#persistentvolumeclaimspec-v1-core">Kubernetes core/v1.PersistentVolumeClaimSpec</a></em></td>
<td><p>The configuration for user volumes. <em>NOTE:</em> Even though the controller will try to force the reclaim policy on created volumes to <code>Retain</code>, you may want to set it explicitly on your storage-class controller as an extra safeguard.</p></td>
</tr>
<tr class="odd">
<td><code>app</code> <em><a href="#AppConfig">AppConfig</a></em></td>
<td><p>App configurations.</p></td>
</tr>
<tr class="even">
<td><code>auth</code> <em><a href="#AuthConfig">AuthConfig</a></em></td>
<td><p>Authentication configurations</p></td>
</tr>
<tr class="odd">
<td><code>secrets</code> <em><a href="#SecretsConfig">SecretsConfig</a></em></td>
<td><p>Secrets backend configurations</p></td>
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
<td><code>certManagerNamespace</code> <em>string</em></td>
<td><p>The namespace cert-manager is running in. Defaults to <code>cert-manager</code>.</p></td>
</tr>
<tr class="even">
<td><code>userdataSpec</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#persistentvolumeclaimspec-v1-core">Kubernetes core/v1.PersistentVolumeClaimSpec</a></em></td>
<td><p>The configuration for user volumes. <em>NOTE:</em> Even though the controller will try to force the reclaim policy on created volumes to <code>Retain</code>, you may want to set it explicitly on your storage-class controller as an extra safeguard.</p></td>
</tr>
<tr class="odd">
<td><code>app</code> <em><a href="#AppConfig">AppConfig</a></em></td>
<td><p>App configurations.</p></td>
</tr>
<tr class="even">
<td><code>auth</code> <em><a href="#AuthConfig">AuthConfig</a></em></td>
<td><p>Authentication configurations</p></td>
</tr>
<tr class="odd">
<td><code>secrets</code> <em><a href="#SecretsConfig">SecretsConfig</a></em></td>
<td><p>Secrets backend configurations</p></td>
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
<td><code>metadata</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#objectmeta-v1-meta">Kubernetes meta/v1.ObjectMeta</a></em></td>
<td>Refer to the Kubernetes API documentation for the fields of the <code>metadata</code> field.</td>
</tr>
<tr class="even">
<td><code>rules</code> <em><a href="#Rule">[]Rule</a></em></td>
<td><p>A list of rules granting access to resources in the VDICluster.</p></td>
</tr>
</tbody>
</table>

### VDIUser

(*Appears on:* [AuthResult](#AuthResult),
[JWTClaims](#JWTClaims),
[SessionResponse](#SessionResponse))

VDIUser represents a user in kVDI. It is the auth providers
responsibility to take an authentication request and generate a JWT with
claims defining this object.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>name</code> <em>string</em></td>
<td><p>A unique name for the user</p></td>
</tr>
<tr class="even">
<td><code>roles</code> <em><a href="#*github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1.VDIUserRole">[]*github.com/tinyzimmer/kvdi/pkg/apis/kvdi/v1alpha1.VDIUserRole</a></em></td>
<td><p>A list of roles applide to the user. The grants associated with each user are embedded in the JWT signed when authenticating.</p></td>
</tr>
<tr class="odd">
<td><code>mfaEnabled</code> <em>bool</em></td>
<td><p>Whether or not MFA is enabled for this user</p></td>
</tr>
</tbody>
</table>

### VDIUserRole

VDIUserRole represents a VDIRole, but only with the data that is to be
embedded in the JWT. Primarily, leaving out useless metadata that will
inflate the token.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>name</code> <em>string</em></td>
<td><p>The name of the role, this must match the VDIRole from which this object derives.</p></td>
</tr>
<tr class="even">
<td><code>rules</code> <em><a href="#Rule">[]Rule</a></em></td>
<td><p>The rules for this role.</p></td>
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
<td><p>The base path to store secrets in vault.</p></td>
</tr>
</tbody>
</table>

Verb (`string` alias)

(*Appears on:* [Rule](#Rule))

Verb represents an API action

------------------------------------------------------------------------

*Generated with `gen-crd-api-reference-docs` on git commit `a251039`.*
