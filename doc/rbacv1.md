## kVDI CRD Reference

### Packages:

-   [rbac.kvdi.io/v1](#rbac.kvdi.io%2fv1)

Types

-   [Resource](#%23rbac.kvdi.io%2fv1.Resource)
-   [Rule](#%23rbac.kvdi.io%2fv1.Rule)
-   [VDIRole](#%23rbac.kvdi.io%2fv1.VDIRole)
-   [Verb](#%23rbac.kvdi.io%2fv1.Verb)

## rbac.kvdi.io/v1

Package v1 contains API Schema definitions for the RBAC v1 API group

Resource Types:

Resource (`string` alias)

(*Appears on:* [Rule](#Rule))

Resource represents the target of an API action

### Rule

(*Appears on:* [VDIRole](#VDIRole))

Rule represents a set of permissions applied to a VDIRole. It mostly
resembles an rbacv1.PolicyRule, with resources being a regex and the
addition of a namespace selector.

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
<td><code>verbs</code> <em><a href="#Verb">[]Verb</a></em></td>
<td><p>The actions this rule applies for. VerbAll matches all actions. Recognized options are: <code>["create", "read", "update", "delete", "use", "launch", "*"]</code></p></td>
</tr>
<tr class="even">
<td><code>resources</code> <em><a href="#Resource">[]Resource</a></em></td>
<td><p>Resources this rule applies to. ResourceAll matches all resources. Recognized options are: <code>["users", "roles", "templates", "serviceaccounts", "*"]</code></p></td>
</tr>
<tr class="odd">
<td><code>resourcePatterns</code> <em>[]string</em></td>
<td><p>Resource regexes that match this rule. This can be template patterns, role names or user names. There is no All representation because * will have that effect on its own when the regex is evaluated. When referring to “serviceaccounts”, only the “use” verb is evaluated in the context of assuming those accounts in desktop sessions.</p>
<p><strong>NOTE</strong>: The <code>kvdi-manager</code> is responsible for launching pods with a service account requested for a given Desktop. If the service account itself contains more permissions than the manager itself, the Kubernetes API will deny the request. The way to remedy this would be to either mirror permissions to that ClusterRole, or make the <code>kvdi-manager</code> itself a cluster admin, both of which come with inherent risks. In the end, you can decide the best approach for your use case with regards to exposing access to the Kubernetes APIs via kvdi sessions.</p></td>
</tr>
<tr class="even">
<td><code>namespaces</code> <em>[]string</em></td>
<td><p>Namespaces this rule applies to. Only evaluated for template launching permissions. Including “*” as an option matches all namespaces.</p></td>
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

Verb (`string` alias)

(*Appears on:* [Rule](#Rule))

Verb represents an API action

------------------------------------------------------------------------

*Generated with `gen-crd-api-reference-docs` on git commit `5275727`.*
