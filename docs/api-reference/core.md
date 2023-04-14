<p>Packages:</p>
<ul>
<li>
<a href="#core.api.onmetal.de%2fv1alpha1">core.api.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="core.api.onmetal.de/v1alpha1">core.api.onmetal.de/v1alpha1</h2>
<div>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</div>
Resource Types:
<ul><li>
<a href="#core.api.onmetal.de/v1alpha1.ResourceQuota">ResourceQuota</a>
</li></ul>
<h3 id="core.api.onmetal.de/v1alpha1.ResourceQuota">ResourceQuota
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
core.api.onmetal.de/v1alpha1
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta">
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceQuotaSpec">
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceList">
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceScopeSelector">
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceQuotaStatus">
ResourceQuotaStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.ConfigMapKeySelector">ConfigMapKeySelector
</h3>
<div>
<p>ConfigMapKeySelector is a reference to a specific &lsquo;key&rsquo; within a ConfigMap resource.
In some instances, <code>key</code> is a required field.</p>
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
<p>Name of the referent.
More info: <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names">https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names</a></p>
</td>
</tr>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The key of the entry in the ConfigMap resource&rsquo;s <code>data</code> field to be used.
Some instances of this field may be defaulted, in others it may be
required.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.IP">IP
</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.IPRange">IPRange</a>)
</p>
<div>
<p>IP is an IP address.</p>
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
<code>-</code><br/>
<em>
<a href="https://pkg.go.dev/net/netip#Addr">
net/netip.Addr
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.IPPrefix">IPPrefix
</h3>
<div>
<p>IPPrefix represents a network prefix.</p>
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
<code>-</code><br/>
<em>
<a href="https://pkg.go.dev/net/netip#Prefix">
net/netip.Prefix
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.IPRange">IPRange
</h3>
<div>
<p>IPRange is an IP range.</p>
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
<code>from</code><br/>
<em>
<a href="#core.api.onmetal.de/v1alpha1.IP">
IP
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>to</code><br/>
<em>
<a href="#core.api.onmetal.de/v1alpha1.IP">
IP
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.LocalUIDReference">LocalUIDReference
</h3>
<div>
<p>LocalUIDReference is a reference to another entity including its UID</p>
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
<p>Name is the name of the referenced entity.</p>
</td>
</tr>
<tr>
<td>
<code>uid</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/types#UID">
k8s.io/apimachinery/pkg/types.UID
</a>
</em>
</td>
<td>
<p>UID is the UID of the referenced entity.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.ObjectSelector">ObjectSelector
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#labelselector-v1-meta">
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
<h3 id="core.api.onmetal.de/v1alpha1.ResourceName">ResourceName
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
</tr><tr><td><p>&#34;requests.memory&#34;</p></td>
<td><p>ResourceRequestsMemory is the amount of requested memory in bytes.</p>
</td>
</tr><tr><td><p>&#34;requests.storage&#34;</p></td>
<td><p>ResourceRequestsStorage is the amount of requested storage in bytes.</p>
</td>
</tr><tr><td><p>&#34;storage&#34;</p></td>
<td><p>ResourceStorage is the amount of storage, in bytes.</p>
</td>
</tr><tr><td><p>&#34;tps&#34;</p></td>
<td><p>ResourceTPS defines max throughput per second. (e.g. 1Gi)</p>
</td>
</tr></tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.ResourceQuotaSpec">ResourceQuotaSpec
</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.ResourceQuota">ResourceQuota</a>)
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceList">
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceScopeSelector">
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
<h3 id="core.api.onmetal.de/v1alpha1.ResourceQuotaStatus">ResourceQuotaStatus
</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.ResourceQuota">ResourceQuota</a>)
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceList">
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceList">
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
<h3 id="core.api.onmetal.de/v1alpha1.ResourceScope">ResourceScope
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.ResourceScopeSelectorRequirement">ResourceScopeSelectorRequirement</a>)
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
<h3 id="core.api.onmetal.de/v1alpha1.ResourceScopeSelector">ResourceScopeSelector
</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.ResourceQuotaSpec">ResourceQuotaSpec</a>)
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceScopeSelectorRequirement">
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
<h3 id="core.api.onmetal.de/v1alpha1.ResourceScopeSelectorOperator">ResourceScopeSelectorOperator
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.ResourceScopeSelectorRequirement">ResourceScopeSelectorRequirement</a>)
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
<h3 id="core.api.onmetal.de/v1alpha1.ResourceScopeSelectorRequirement">ResourceScopeSelectorRequirement
</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.ResourceScopeSelector">ResourceScopeSelector</a>)
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceScope">
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
<a href="#core.api.onmetal.de/v1alpha1.ResourceScopeSelectorOperator">
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
<h3 id="core.api.onmetal.de/v1alpha1.SecretKeySelector">SecretKeySelector
</h3>
<div>
<p>SecretKeySelector is a reference to a specific &lsquo;key&rsquo; within a Secret resource.
In some instances, <code>key</code> is a required field.</p>
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
<p>Name of the referent.
More info: <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names">https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names</a></p>
</td>
</tr>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The key of the entry in the Secret resource&rsquo;s <code>data</code> field to be used.
Some instances of this field may be defaulted, in others it may be
required.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.Taint">Taint
</h3>
<div>
<p>Taint is attached to a resource pool. It has the specified Effect on
any resource that does not tolerate the Taint.</p>
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
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<p>The taint key to be applied to a resource pool.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<p>The taint value corresponding to the taint key.</p>
</td>
</tr>
<tr>
<td>
<code>effect</code><br/>
<em>
<a href="#core.api.onmetal.de/v1alpha1.TaintEffect">
TaintEffect
</a>
</em>
</td>
<td>
<p>The effect of the taint on resources
that do not tolerate the taint.
Valid effects are NoSchedule, PreferNoSchedule and NoExecute.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.TaintEffect">TaintEffect
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.Taint">Taint</a>, <a href="#core.api.onmetal.de/v1alpha1.Toleration">Toleration</a>)
</p>
<div>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;NoSchedule&#34;</p></td>
<td><p>TaintEffectNoSchedule causes to not allow new resources to schedule onto the resource pool unless they tolerate
the taint, but allow all already-running resources to continue running.
Enforced by the scheduler.</p>
</td>
</tr></tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.Toleration">Toleration
</h3>
<div>
<p>Toleration makes the resource the toleration is attached to tolerate any taint that matches
the triple <key,value,effect> using the matching operator <operator>.</p>
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
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<p>Key is the taint key that the toleration applies to. Empty means match all taint keys.
If the key is empty, operator must be Exists; this combination means to match all values and all keys.</p>
</td>
</tr>
<tr>
<td>
<code>operator</code><br/>
<em>
<a href="#core.api.onmetal.de/v1alpha1.TolerationOperator">
TolerationOperator
</a>
</em>
</td>
<td>
<p>Operator represents a key&rsquo;s relationship to the value.
Valid operators are Exists and Equal. Defaults to Equal.
Exists is equivalent to wildcard for value, so that a resource can
tolerate all taints of a particular category.</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
string
</em>
</td>
<td>
<p>Value is the taint value the toleration matches to.
If the operator is Exists, the value should be empty, otherwise just a regular string.</p>
</td>
</tr>
<tr>
<td>
<code>effect</code><br/>
<em>
<a href="#core.api.onmetal.de/v1alpha1.TaintEffect">
TaintEffect
</a>
</em>
</td>
<td>
<p>Effect indicates the taint effect to match. Empty means match all taint effects.
When specified, allowed values are NoSchedule.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.TolerationOperator">TolerationOperator
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#core.api.onmetal.de/v1alpha1.Toleration">Toleration</a>)
</p>
<div>
<p>TolerationOperator is the set of operators that can be used in a toleration.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Equal&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Exists&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="core.api.onmetal.de/v1alpha1.UIDReference">UIDReference
</h3>
<div>
<p>UIDReference is a reference to another entity in a potentially different namespace including its UID.</p>
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
<code>namespace</code><br/>
<em>
string
</em>
</td>
<td>
<p>Namespace is the namespace of the referenced entity. If empty,
the same namespace as the referring resource is implied.</p>
</td>
</tr>
<tr>
<td>
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
<p>Name is the name of the referenced entity.</p>
</td>
</tr>
<tr>
<td>
<code>uid</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/types#UID">
k8s.io/apimachinery/pkg/types.UID
</a>
</em>
</td>
<td>
<p>UID is the UID of the referenced entity.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
</em></p>
