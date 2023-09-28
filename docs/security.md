## Security

All traffic between processes is encrypted with mTLS.
The UI for the "desktop" containers is placed behind a VNC server listening on a UNIX socket and a sidecar to the container will proxy validated websocket connections to it.

![img](doc/kvdi_arch.png)

User authentication is provided by "providers". There are currently three implementations:

 * `local-auth` : A `passwd` like file is kept in the Secrets backend (k8s or vault) mapping users to roles and password hashes. This is primarily meant for development, but you could secure your environment in a way to make it viable for a small number of users.

 * `ldap-auth` : An LDAP/AD server is used for autenticating users. VDIRoles can be tied to 
 security groups in LDAP via annotations. When a user is authenticated, their groups are queried to see if they are bound to any VDIRoles.

 * `oidc-auth` : An OpenID or OAuth provider is used for authenticating users. If using an Oauth provider, it must support the `openid` scope. When a user is authenticated, a configurable `groups` claim is requested from the provider that can be mapped to VDIRoles similarly to `ldap-auth`. If the provider does not support a `groups` claim, you can configure `kVDI` to allow all authenticated users.

 All three authentication methods also support MFA.