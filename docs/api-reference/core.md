<p>Packages:</p>
<ul>
<li>
<a href="#core.ironcore.dev%2fv1alpha1">core.ironcore.dev/v1alpha1</a>
</li>
</ul>
<h2 id="core.ironcore.dev/v1alpha1">core.ironcore.dev/v1alpha1</h2>
<div>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</div>
Resource Types:
<ul><li>
<a href="#core.ironcore.dev/v1alpha1.ResourceQuota">ResourceQuota</a>
</li></ul>
<h3 id="core.ironcore.dev/v1alpha1.ResourceQuota">ResourceQuota
</h3>
<div>
<p>ResourceQuota is the Schema for the resourcequotas API</p>
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
<code>apiVersion</code><br/>
string</td>
<td>
<code>
core.ironcore.dev/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>ResourceQuota</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#objectmeta-v1-meta">
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
<a href="#core.ironcore.dev/v1alpha1.ResourceQuotaSpec">
ResourceQuotaSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>hard</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceList">
ResourceList
</a>
</em>
</td>
<td>
<p>Hard is a ResourceList of the strictly enforced amount of resources.</p>
</td>
</tr>
<tr>
<td>
<code>scopeSelector</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceScopeSelector">
ResourceScopeSelector
</a>
</em>
</td>
<td>
<p>ScopeSelector selects the resources that are subject to this quota.
Note: By using certain ScopeSelectors, only certain resources may be tracked.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceQuotaStatus">
ResourceQuotaStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ClassType">ClassType
(<code>string</code> alias)</h3>
<div>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;machine&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;volume&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ObjectSelector">ObjectSelector
</h3>
<div>
<p>ObjectSelector specifies how to select objects of a certain kind.</p>
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
<code>kind</code><br/>
<em>
string
</em>
</td>
<td>
<p>Kind is the kind of object to select.</p>
</td>
</tr>
<tr>
<td>
<code>LabelSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.34/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>
(Members of <code>LabelSelector</code> are embedded into this type.)
</p>
<p>LabelSelector is the label selector to select objects of the specified Kind by.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ResourceName">ResourceName
(<code>string</code> alias)</h3>
<div>
<p>ResourceName is the name of a resource, most often used alongside a resource.Quantity.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;cpu&#34;</p></td>
<td><p>ResourceCPU is the amount of cpu in cores.</p>
</td>
</tr><tr><td><p>&#34;iops&#34;</p></td>
<td><p>ResourceIOPS defines max IOPS in input/output operations per second.</p>
</td>
</tr><tr><td><p>&#34;memory&#34;</p></td>
<td><p>ResourceMemory is the amount of memory in bytes.</p>
</td>
</tr><tr><td><p>&#34;requests.cpu&#34;</p></td>
<td><p>ResourceRequestsCPU is the amount of requested cpu in cores.</p>
</td>
</tr><tr><td><p>&#34;requests.iops&#34;</p></td>
<td><p>ResourceRequestsIOPS is the amount of requested IOPS in input/output operations per second.</p>
</td>
</tr><tr><td><p>&#34;requests.memory&#34;</p></td>
<td><p>ResourceRequestsMemory is the amount of requested memory in bytes.</p>
</td>
</tr><tr><td><p>&#34;requests.storage&#34;</p></td>
<td><p>ResourceRequestsStorage is the amount of requested storage in bytes.</p>
</td>
</tr><tr><td><p>&#34;requests.tps&#34;</p></td>
<td><p>ResourceRequestsTPS is the amount of requested throughput per second.</p>
</td>
</tr><tr><td><p>&#34;storage&#34;</p></td>
<td><p>ResourceStorage is the amount of storage, in bytes.</p>
</td>
</tr><tr><td><p>&#34;tps&#34;</p></td>
<td><p>ResourceTPS defines max throughput per second. (e.g. 1Gi)</p>
</td>
</tr></tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ResourceQuotaSpec">ResourceQuotaSpec
</h3>
<p>
(<em>Appears on:</em><a href="#core.ironcore.dev/v1alpha1.ResourceQuota">ResourceQuota</a>)
</p>
<div>
<p>ResourceQuotaSpec defines the desired state of ResourceQuotaSpec</p>
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
<code>hard</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceList">
ResourceList
</a>
</em>
</td>
<td>
<p>Hard is a ResourceList of the strictly enforced amount of resources.</p>
</td>
</tr>
<tr>
<td>
<code>scopeSelector</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceScopeSelector">
ResourceScopeSelector
</a>
</em>
</td>
<td>
<p>ScopeSelector selects the resources that are subject to this quota.
Note: By using certain ScopeSelectors, only certain resources may be tracked.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ResourceQuotaStatus">ResourceQuotaStatus
</h3>
<p>
(<em>Appears on:</em><a href="#core.ironcore.dev/v1alpha1.ResourceQuota">ResourceQuota</a>)
</p>
<div>
<p>ResourceQuotaStatus is the status of a ResourceQuota.</p>
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
<code>hard</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceList">
ResourceList
</a>
</em>
</td>
<td>
<p>Hard are the currently enforced hard resource limits. Hard may be less than used in
case the limits were introduced / updated after more than allowed resources were already present.</p>
</td>
</tr>
<tr>
<td>
<code>used</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceList">
ResourceList
</a>
</em>
</td>
<td>
<p>Used is the amount of currently used resources.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ResourceScope">ResourceScope
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#core.ironcore.dev/v1alpha1.ResourceScopeSelectorRequirement">ResourceScopeSelectorRequirement</a>)
</p>
<div>
<p>ResourceScope is a scope of a resource.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;BucketClass&#34;</p></td>
<td><p>ResourceScopeBucketClass refers to the bucket class of a resource.</p>
</td>
</tr><tr><td><p>&#34;MachineClass&#34;</p></td>
<td><p>ResourceScopeMachineClass refers to the machine class of a resource.</p>
</td>
</tr><tr><td><p>&#34;VolumeClass&#34;</p></td>
<td><p>ResourceScopeVolumeClass refers to the volume class of a resource.</p>
</td>
</tr></tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ResourceScopeSelector">ResourceScopeSelector
</h3>
<p>
(<em>Appears on:</em><a href="#core.ironcore.dev/v1alpha1.ResourceQuotaSpec">ResourceQuotaSpec</a>)
</p>
<div>
<p>ResourceScopeSelector selects</p>
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
<code>matchExpressions</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceScopeSelectorRequirement">
[]ResourceScopeSelectorRequirement
</a>
</em>
</td>
<td>
<p>MatchExpressions is a list of ResourceScopeSelectorRequirement to match resources by.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ResourceScopeSelectorOperator">ResourceScopeSelectorOperator
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#core.ironcore.dev/v1alpha1.ResourceScopeSelectorRequirement">ResourceScopeSelectorRequirement</a>)
</p>
<div>
<p>ResourceScopeSelectorOperator is an operator to compare a ResourceScope with values.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;DoesNotExist&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Exists&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;In&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;NotIn&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="core.ironcore.dev/v1alpha1.ResourceScopeSelectorRequirement">ResourceScopeSelectorRequirement
</h3>
<p>
(<em>Appears on:</em><a href="#core.ironcore.dev/v1alpha1.ResourceScopeSelector">ResourceScopeSelector</a>)
</p>
<div>
<p>ResourceScopeSelectorRequirement is a requirement for a resource using a ResourceScope alongside
a ResourceScopeSelectorOperator with Values (depending on the ResourceScopeSelectorOperator).</p>
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
<code>scopeName</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceScope">
ResourceScope
</a>
</em>
</td>
<td>
<p>ScopeName is the ResourceScope to make a requirement for.</p>
</td>
</tr>
<tr>
<td>
<code>operator</code><br/>
<em>
<a href="#core.ironcore.dev/v1alpha1.ResourceScopeSelectorOperator">
ResourceScopeSelectorOperator
</a>
</em>
</td>
<td>
<p>Operator is the ResourceScopeSelectorOperator to check the ScopeName with in a resource.</p>
</td>
</tr>
<tr>
<td>
<code>values</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Values are the values to compare the Operator with the ScopeName. May be optional.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
</em></p>
