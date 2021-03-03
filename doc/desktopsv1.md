## kVDI CRD Reference

### Packages:

-   [desktops.kvdi.io/v1](#desktops.kvdi.io%2fv1)

Types

-   [DesktopConfig](#%23desktops.kvdi.io%2fv1.DesktopConfig)
-   [DesktopInit](#%23desktops.kvdi.io%2fv1.DesktopInit)
-   [DockerInDockerConfig](#%23desktops.kvdi.io%2fv1.DockerInDockerConfig)
-   [ProxyConfig](#%23desktops.kvdi.io%2fv1.ProxyConfig)
-   [QEMUConfig](#%23desktops.kvdi.io%2fv1.QEMUConfig)
-   [Session](#%23desktops.kvdi.io%2fv1.Session)
-   [SessionSpec](#%23desktops.kvdi.io%2fv1.SessionSpec)
-   [Template](#%23desktops.kvdi.io%2fv1.Template)
-   [TemplateSpec](#%23desktops.kvdi.io%2fv1.TemplateSpec)

## desktops.kvdi.io/v1

Package v1 contains API Schema definitions for the Desktops v1 API group

Resource Types:

### DesktopConfig

(*Appears on:* [TemplateSpec](#TemplateSpec))

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
<td><code>image</code> <em>string</em></td>
<td><p>The docker repository and tag to use for desktops booted from this template.</p></td>
</tr>
<tr class="even">
<td><code>imagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the container image.</p></td>
</tr>
<tr class="odd">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to apply to desktops booted from this template.</p></td>
</tr>
<tr class="even">
<td><code>env</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#envvar-v1-core">[]Kubernetes core/v1.EnvVar</a></em></td>
<td><p>Additional environment variables to pass to containers booted from this template.</p></td>
</tr>
<tr class="odd">
<td><code>envTemplates</code> <em>map[string]string</em></td>
<td><p>Optionally map additional information about the user (and potentially extended further in the future) into the environment of desktops booted from this template. The keys in the map are the environment variable to set inside the desktop, and the values are go templates or strings to set to the value. Currently the go templates are only passed a <code>Session</code> object containing the information in the claims for the user that created the desktop. For more information see the <a href="https://github.com/tinyzimmer/kvdi/blob/main/pkg/types/auth_types.go#L79">JWTCaims object</a> and corresponding go types.</p></td>
</tr>
<tr class="even">
<td><code>volumeMounts</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">[]Kubernetes core/v1.VolumeMount</a></em></td>
<td><p>Volume mounts for the desktop container.</p></td>
</tr>
<tr class="odd">
<td><code>volumeDevices</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumedevice-v1-core">[]Kubernetes core/v1.VolumeDevice</a></em></td>
<td><p>Volume devices for the desktop container.</p></td>
</tr>
<tr class="even">
<td><code>capabilities</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#capability-v1-core">[]Kubernetes core/v1.Capability</a></em></td>
<td><p>Extra system capabilities to add to desktops booted from this template.</p></td>
</tr>
<tr class="odd">
<td><code>dnsPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#dnspolicy-v1-core">Kubernetes core/v1.DNSPolicy</a></em></td>
<td><p>Set the DNS policy for desktops booted from this template. Defaults to the Kubernetes default (ClusterFirst).</p></td>
</tr>
<tr class="even">
<td><code>dnsConfig</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#poddnsconfig-v1-core">Kubernetes core/v1.PodDNSConfig</a></em></td>
<td><p>Specify the DNS parameters for desktops booted from this template. Parameters will be merged into the configuration based off the <code>dnsPolicy</code>.</p></td>
</tr>
<tr class="odd">
<td><code>allowRoot</code> <em>bool</em></td>
<td><p>AllowRoot will pass the ENABLE_ROOT envvar to the container. In the Dockerfiles in this repository, this will add the user to the sudo group and ability to sudo with no password.</p></td>
</tr>
<tr class="even">
<td><code>init</code> <em><a href="#DesktopInit">DesktopInit</a></em></td>
<td><p>The type of init system inside the image, currently only <code>supervisord</code> and <code>systemd</code> are supported. Defaults to <code>systemd</code>. <code>systemd</code> containers are run privileged and downgrading to the desktop user must be done within the image’s init process. <code>supervisord</code> containers are run with minimal capabilities and directly as the desktop user.</p></td>
</tr>
</tbody>
</table>

DesktopInit (`string` alias)

(*Appears on:* [DesktopConfig](#DesktopConfig))

DesktopInit represents the init system that the desktop container uses.

### DockerInDockerConfig

(*Appears on:* [TemplateSpec](#TemplateSpec))

DockerInDockerConfig is a configuration for mounting a DinD sidecar with
desktops booted from the template. This will provide ephemeral docker
daemons and storage to sessions.

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
<td><p>The image to use for the dind sidecar. Defaults to <code>docker:dind</code>.</p></td>
</tr>
<tr class="even">
<td><code>imagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the container image.</p></td>
</tr>
<tr class="odd">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource restraints to place on the dind sidecar.</p></td>
</tr>
<tr class="even">
<td><code>volumeMounts</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumemount-v1-core">[]Kubernetes core/v1.VolumeMount</a></em></td>
<td><p>Volume mounts for the dind container.</p></td>
</tr>
<tr class="odd">
<td><code>volumeDevices</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volumedevice-v1-core">[]Kubernetes core/v1.VolumeDevice</a></em></td>
<td><p>Volume devices for the dind container.</p></td>
</tr>
</tbody>
</table>

### ProxyConfig

(*Appears on:* [TemplateSpec](#TemplateSpec))

ProxyConfig represents configurations for the display/audio proxy.

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
<td><p>The image to use for the sidecar that proxies mTLS connections to the local VNC server inside the Desktop. Defaults to the public kvdi-proxy image matching the version of the currrently running manager.</p></td>
</tr>
<tr class="even">
<td><code>imagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the container image.</p></td>
</tr>
<tr class="odd">
<td><code>allowFileTransfer</code> <em>bool</em></td>
<td><p>AllowFileTransfer will mount the user’s home directory inside the kvdi-proxy image. This enables the API endpoint for exploring, downloading, and uploading files to desktop sessions booted from this template. When using a <code>qemu</code> configuration with SPICE, file upload is enabled by default.</p></td>
</tr>
<tr class="even">
<td><code>socketAddr</code> <em>string</em></td>
<td><p>The address the display server listens on inside the image. This defaults to the UNIX socket <code>/var/run/kvdi/display.sock</code>. The kvdi-proxy sidecar will forward websockify requests validated by mTLS to this socket. Must be in the format of <code>tcp://{host}:{port}</code> or <code>unix://{path}</code>. This will usually be a VNC server unless using a <code>qemu</code> configuration with SPICE. If using custom init scripts inside your containers, this value is set to the <code>DISPLAY_SOCK_ADDR</code> environment variable.</p></td>
</tr>
<tr class="odd">
<td><code>pulseServer</code> <em>string</em></td>
<td><p>Override the address of the PulseAudio server that the proxy will try to connect to when serving audio. This defaults to what the ubuntu/arch desktop images are configured to do during init, which is to place a socket in the user’s run directory. The value is assumed to be a unix socket.</p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource restraints to place on the proxy sidecar.</p></td>
</tr>
</tbody>
</table>

### QEMUConfig

(*Appears on:* [TemplateSpec](#TemplateSpec))

QEMUConfig represents configurations for running a qemu virtual machine
for instances booted from this template.

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>diskImage</code> <em>string</em></td>
<td><p>The container image bundling the disks for this template.</p></td>
</tr>
<tr class="even">
<td><code>diskImagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the disk image.</p></td>
</tr>
<tr class="odd">
<td><code>useCSI</code> <em>bool</em></td>
<td><p>Set to true to use the image-populator CSI to mount the disk images to a qemu container. You must have the <a href="https://github.com/kubernetes-csi/csi-driver-image-populator">image-populator</a> driver installed. Defaults to copying the contents out of the disk image via an init container. This is experimental and not really tested.</p></td>
</tr>
<tr class="even">
<td><code>qemuImage</code> <em>string</em></td>
<td><p>The container image containing the QEMU utilities to use to launch the VM. Defaults to <code>ghcr.io/tinyzimmer/kvdi:qemu-latest</code>.</p></td>
</tr>
<tr class="odd">
<td><code>qemuImagePullPolicy</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#pullpolicy-v1-core">Kubernetes core/v1.PullPolicy</a></em></td>
<td><p>The pull policy to use when pulling the QEMU image.</p></td>
</tr>
<tr class="even">
<td><code>qemuResources</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#resourcerequirements-v1-core">Kubernetes core/v1.ResourceRequirements</a></em></td>
<td><p>Resource requirements to place on the qemu runner instance.</p></td>
</tr>
<tr class="odd">
<td><code>diskPath</code> <em>string</em></td>
<td><p>The path to the boot volume inside the disk image. Defaults to <code>/disk/boot.img</code>.</p></td>
</tr>
<tr class="even">
<td><code>cloudInitPath</code> <em>string</em></td>
<td><p>The path to a pre-built cloud init image to use when booting the VM inside the disk image. Defaults to an auto-generated one at runtime.</p></td>
</tr>
<tr class="odd">
<td><code>cpus</code> <em>int</em></td>
<td><p>The number of vCPUs to assign the virtual machine. Defaults to 1.</p></td>
</tr>
<tr class="even">
<td><code>memory</code> <em>int</em></td>
<td><p>The amount of memory to assign the virtual machine (in MB). Defaults to 1024.</p></td>
</tr>
<tr class="odd">
<td><code>spice</code> <em>bool</em></td>
<td><p>Set to true to use the SPICE protocol when proxying the display. If using custom qemu runners, this sets the <code>SPICE_DISPLAY</code> environment variable to <code>true</code>. The runners provided by this repository will tell qemu to set up a SPICE server at <code>proxy.socketAddr</code>. The default is to use VNC. This value is also used by the UI to determine which protocol to expect from a display connection.</p></td>
</tr>
</tbody>
</table>

### Session

Session is the Schema for the sessions API

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
<td><code>spec</code> <em><a href="#SessionSpec">SessionSpec</a></em></td>
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
<tr class="even">
<td><code>serviceAccount</code> <em>string</em></td>
<td><p>A service account to tie to the pod for this instance.</p></td>
</tr>
</tbody>
</table></td>
</tr>
<tr class="odd">
<td><code>status</code> <em><a href="#SessionStatus">SessionStatus</a></em></td>
<td></td>
</tr>
</tbody>
</table>

### SessionSpec

(*Appears on:* [Session](#Session))

SessionSpec defines the desired state of Session

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
<tr class="even">
<td><code>serviceAccount</code> <em>string</em></td>
<td><p>A service account to tie to the pod for this instance.</p></td>
</tr>
</tbody>
</table>

### Template

Template is the Schema for the templates API

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
<td><code>spec</code> <em><a href="#TemplateSpec">TemplateSpec</a></em></td>
<td><br />
<br />

<table>
<tbody>
<tr class="odd">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Any pull secrets required for pulling the container image.</p></td>
</tr>
<tr class="even">
<td><code>volumes</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">[]Kubernetes core/v1.Volume</a></em></td>
<td><p>Additional volumes to attach to pods booted from this template. To mount them there must be corresponding <code>volumeMounts</code> or <code>volumeDevices</code> specified.</p></td>
</tr>
<tr class="odd">
<td><code>desktop</code> <em><a href="#DesktopConfig">DesktopConfig</a></em></td>
<td><p>Configuration options for the instances. These are highly dependant on using the Dockerfiles (or close derivitives) provided in this repository.</p></td>
</tr>
<tr class="even">
<td><code>proxy</code> <em><a href="#ProxyConfig">ProxyConfig</a></em></td>
<td><p>Configurations for the display proxy.</p></td>
</tr>
<tr class="odd">
<td><code>dind</code> <em><a href="#DockerInDockerConfig">DockerInDockerConfig</a></em></td>
<td><p>Docker-in-docker configurations for running a dind sidecar along with desktop instances.</p></td>
</tr>
<tr class="even">
<td><code>qemu</code> <em><a href="#QEMUConfig">QEMUConfig</a></em></td>
<td><p>QEMU configurations for this template. When defined, VMs are used instead of containers for desktop sessions. This object is mututally exclusive with <code>desktop</code> and will take precedence when defined.</p></td>
</tr>
<tr class="odd">
<td><code>tags</code> <em>map[string]string</em></td>
<td><p>Arbitrary tags for displaying in the app UI.</p></td>
</tr>
</tbody>
</table></td>
</tr>
</tbody>
</table>

### TemplateSpec

(*Appears on:* [Template](#Template))

TemplateSpec defines the desired state of Template

<table>
<thead>
<tr class="header">
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr class="odd">
<td><code>imagePullSecrets</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#localobjectreference-v1-core">[]Kubernetes core/v1.LocalObjectReference</a></em></td>
<td><p>Any pull secrets required for pulling the container image.</p></td>
</tr>
<tr class="even">
<td><code>volumes</code> <em><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#volume-v1-core">[]Kubernetes core/v1.Volume</a></em></td>
<td><p>Additional volumes to attach to pods booted from this template. To mount them there must be corresponding <code>volumeMounts</code> or <code>volumeDevices</code> specified.</p></td>
</tr>
<tr class="odd">
<td><code>desktop</code> <em><a href="#DesktopConfig">DesktopConfig</a></em></td>
<td><p>Configuration options for the instances. These are highly dependant on using the Dockerfiles (or close derivitives) provided in this repository.</p></td>
</tr>
<tr class="even">
<td><code>proxy</code> <em><a href="#ProxyConfig">ProxyConfig</a></em></td>
<td><p>Configurations for the display proxy.</p></td>
</tr>
<tr class="odd">
<td><code>dind</code> <em><a href="#DockerInDockerConfig">DockerInDockerConfig</a></em></td>
<td><p>Docker-in-docker configurations for running a dind sidecar along with desktop instances.</p></td>
</tr>
<tr class="even">
<td><code>qemu</code> <em><a href="#QEMUConfig">QEMUConfig</a></em></td>
<td><p>QEMU configurations for this template. When defined, VMs are used instead of containers for desktop sessions. This object is mututally exclusive with <code>desktop</code> and will take precedence when defined.</p></td>
</tr>
<tr class="odd">
<td><code>tags</code> <em>map[string]string</em></td>
<td><p>Arbitrary tags for displaying in the app UI.</p></td>
</tr>
</tbody>
</table>

------------------------------------------------------------------------

*Generated with `gen-crd-api-reference-docs` on git commit `5275727`.*
