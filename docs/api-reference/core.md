<p>Packages:</p>
<ul>
<li>
<a href="#core.onmetal.de%2fv1alpha1">core.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="core.onmetal.de/v1alpha1">core.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="core.onmetal.de/v1alpha1.AZState">AZState
</h3>
<p>
(<em>Appears on:</em><a href="#core.onmetal.de/v1alpha1.RegionStatus">RegionStatus</a>)
</p>
<div>
<p>ZoneState describes the state of an AvailabilityZone within a region</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>StateFields</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.StateFields">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.StateFields
</a>
</em>
</td>
<td>
<p>
(Members of <code>StateFields</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.Account">Account
</h3>
<div>
<p>Account is the Schema for the accounts API</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#core.onmetal.de/v1alpha1.AccountSpec">
AccountSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>createdBy</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#subject-v1-rbac">
Kubernetes rbac/v1.Subject
</a>
</em>
</td>
<td>
<p>CreatedBy is a subject representing a user name, an email address, or any other identifier of a user
who created the account.</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<p>Description is a human-readable description of what the account is used for.</p>
</td>
</tr>
<tr>
<td>
<code>owner</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#subject-v1-rbac">
Kubernetes rbac/v1.Subject
</a>
</em>
</td>
<td>
<p>Owner is a subject representing a user name, an email address, or any other identifier of a user owning
the account.</p>
</td>
</tr>
<tr>
<td>
<code>purpose</code><br/>
<em>
string
</em>
</td>
<td>
<p>Purpose is a human-readable explanation of the account&rsquo;s purpose.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#core.onmetal.de/v1alpha1.AccountStatus">
AccountStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.AccountSpec">AccountSpec
</h3>
<p>
(<em>Appears on:</em><a href="#core.onmetal.de/v1alpha1.Account">Account</a>)
</p>
<div>
<p>AccountSpec defines the desired state of Account</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>createdBy</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#subject-v1-rbac">
Kubernetes rbac/v1.Subject
</a>
</em>
</td>
<td>
<p>CreatedBy is a subject representing a user name, an email address, or any other identifier of a user
who created the account.</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<p>Description is a human-readable description of what the account is used for.</p>
</td>
</tr>
<tr>
<td>
<code>owner</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#subject-v1-rbac">
Kubernetes rbac/v1.Subject
</a>
</em>
</td>
<td>
<p>Owner is a subject representing a user name, an email address, or any other identifier of a user owning
the account.</p>
</td>
</tr>
<tr>
<td>
<code>purpose</code><br/>
<em>
string
</em>
</td>
<td>
<p>Purpose is a human-readable explanation of the account&rsquo;s purpose.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.AccountStatus">AccountStatus
</h3>
<p>
(<em>Appears on:</em><a href="#core.onmetal.de/v1alpha1.Account">Account</a>)
</p>
<div>
<p>AccountStatus defines the observed state of Account</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>StateFields</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.StateFields">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.StateFields
</a>
</em>
</td>
<td>
<p>
(Members of <code>StateFields</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code><br/>
<em>
string
</em>
</td>
<td>
<p>Namespace references the namespace of the account</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.Region">Region
</h3>
<div>
<p>Region is the Schema for the regions API</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#core.onmetal.de/v1alpha1.RegionSpec">
RegionSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>location</code><br/>
<em>
string
</em>
</td>
<td>
<p>Location describes the physical location of the region</p>
</td>
</tr>
<tr>
<td>
<code>availabiltyZone</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>AvailabilityZones represents the availability zones in a given region</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#core.onmetal.de/v1alpha1.RegionStatus">
RegionStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.RegionSpec">RegionSpec
</h3>
<p>
(<em>Appears on:</em><a href="#core.onmetal.de/v1alpha1.Region">Region</a>)
</p>
<div>
<p>RegionSpec defines the desired state of Region</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>location</code><br/>
<em>
string
</em>
</td>
<td>
<p>Location describes the physical location of the region</p>
</td>
</tr>
<tr>
<td>
<code>availabiltyZone</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>AvailabilityZones represents the availability zones in a given region</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.RegionStatus">RegionStatus
</h3>
<p>
(<em>Appears on:</em><a href="#core.onmetal.de/v1alpha1.Region">Region</a>)
</p>
<div>
<p>RegionStatus defines the observed state of Region</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>StateFields</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.StateFields">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.StateFields
</a>
</em>
</td>
<td>
<p>
(Members of <code>StateFields</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>availabilityZones</code><br/>
<em>
<a href="#core.onmetal.de/v1alpha1.AZState">
[]AZState
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.Scope">Scope
</h3>
<div>
<p>Scope is the Schema for the scopes API</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#objectmeta-v1-meta">
Kubernetes meta/v1.ObjectMeta
</a>
</em>
</td>
<td>
Refer to the Kubernetes API documentation for the fields of the
<code>metadata</code> field.
</td>
</tr>
<tr>
<td>
<code>spec</code><br/>
<em>
<a href="#core.onmetal.de/v1alpha1.ScopeSpec">
ScopeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<p>Description is a human-readable description of what the scope is used for.</p>
</td>
</tr>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region describes the region scope</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#core.onmetal.de/v1alpha1.ScopeStatus">
ScopeStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.ScopeSpec">ScopeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#core.onmetal.de/v1alpha1.Scope">Scope</a>)
</p>
<div>
<p>ScopeSpec defines the desired state of Scope</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<p>Description is a human-readable description of what the scope is used for.</p>
</td>
</tr>
<tr>
<td>
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region describes the region scope</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.onmetal.de/v1alpha1.ScopeStatus">ScopeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#core.onmetal.de/v1alpha1.Scope">Scope</a>)
</p>
<div>
<p>ScopeStatus defines the observed state of Scope</p>
</div>
<table>
<thead>
<tr>
<th>Field</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>
<code>StateFields</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.StateFields">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.StateFields
</a>
</em>
</td>
<td>
<p>
(Members of <code>StateFields</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>namespace</code><br/>
<em>
string
</em>
</td>
<td>
<p>Namespace references the namespace of the scope</p>
</td>
</tr>
<tr>
<td>
<code>parentScope</code><br/>
<em>
string
</em>
</td>
<td>
<p>ParentScope describes the parent scope, if empty the account
should be used as the top reference</p>
</td>
</tr>
<tr>
<td>
<code>parentNamespace</code><br/>
<em>
string
</em>
</td>
<td>
<p>ParentNamespace represents the namespace of the parent scope</p>
</td>
</tr>
<tr>
<td>
<code>account</code><br/>
<em>
string
</em>
</td>
<td>
<p>Account describes the account this scope belongs to</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>003d1e1</code>.
</em></p>
