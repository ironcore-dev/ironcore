<p>Packages:</p>
<ul>
<li>
<a href="#ipam.api.onmetal.de%2fv1alpha1">ipam.api.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="ipam.api.onmetal.de/v1alpha1">ipam.api.onmetal.de/v1alpha1</h2>
<div>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</div>
Resource Types:
<ul><li>
<a href="#ipam.api.onmetal.de/v1alpha1.Prefix">Prefix</a>
</li><li>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocation">PrefixAllocation</a>
</li></ul>
<h3 id="ipam.api.onmetal.de/v1alpha1.Prefix">Prefix
</h3>
<div>
<p>Prefix is the Schema for the prefixes API</p>
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
ipam.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>Prefix</code></td>
</tr>
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
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSpec">
PrefixSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>ipFamily</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#ipfamily-v1-core">
Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamily is the IPFamily of the prefix.
If unset but Prefix is set, this can be inferred.</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Prefix is the prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>prefixLength</code><br/>
<em>
int32
</em>
</td>
<td>
<p>PrefixLength is the length of prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>parentRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>ParentRef references the parent to allocate the Prefix from.
If ParentRef and ParentSelector is empty, the Prefix is considered a root prefix and thus
allocated by itself.</p>
</td>
</tr>
<tr>
<td>
<code>parentSelector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ParentSelector is the LabelSelector to use for determining the parent for this Prefix.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixStatus">
PrefixStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixAllocation">PrefixAllocation
</h3>
<div>
<p>PrefixAllocation is the Schema for the prefixallocations API</p>
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
ipam.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>PrefixAllocation</code></td>
</tr>
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
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationSpec">
PrefixAllocationSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>ipFamily</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#ipfamily-v1-core">
Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamily is the IPFamily of the prefix.
If unset but Prefix is set, this can be inferred.</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Prefix is the prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>prefixLength</code><br/>
<em>
int32
</em>
</td>
<td>
<p>PrefixLength is the length of prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>prefixRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>PrefixRef references the prefix to allocate from.</p>
</td>
</tr>
<tr>
<td>
<code>prefixSelector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>PrefixSelector selects the prefix to allocate from.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationStatus">
PrefixAllocationStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixAllocationCondition">PrefixAllocationCondition
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationStatus">PrefixAllocationStatus</a>)
</p>
<div>
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
<code>type</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationConditionType">
PrefixAllocationConditionType
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>reason</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixAllocationConditionType">PrefixAllocationConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationCondition">PrefixAllocationCondition</a>)
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
<tbody><tr><td><p>&#34;Ready&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixAllocationSpec">PrefixAllocationSpec
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocation">PrefixAllocation</a>)
</p>
<div>
<p>PrefixAllocationSpec defines the desired state of PrefixAllocation</p>
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
<code>ipFamily</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#ipfamily-v1-core">
Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamily is the IPFamily of the prefix.
If unset but Prefix is set, this can be inferred.</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Prefix is the prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>prefixLength</code><br/>
<em>
int32
</em>
</td>
<td>
<p>PrefixLength is the length of prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>prefixRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>PrefixRef references the prefix to allocate from.</p>
</td>
</tr>
<tr>
<td>
<code>prefixSelector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>PrefixSelector selects the prefix to allocate from.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixAllocationStatus">PrefixAllocationStatus
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocation">PrefixAllocation</a>)
</p>
<div>
<p>PrefixAllocationStatus is the status of a PrefixAllocation.</p>
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
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Prefix is the allocated prefix, if any</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationCondition">
[]PrefixAllocationCondition
</a>
</em>
</td>
<td>
<p>Conditions represent various state aspects of a PrefixAllocation.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixCondition">PrefixCondition
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.PrefixStatus">PrefixStatus</a>)
</p>
<div>
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
<code>type</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixConditionType">
PrefixConditionType
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>reason</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>message</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixConditionType">PrefixConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.PrefixCondition">PrefixCondition</a>)
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
<tbody><tr><td><p>&#34;Ready&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixSpec">PrefixSpec
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.Prefix">Prefix</a>)
</p>
<div>
<p>PrefixSpec defines the desired state of Prefix</p>
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
<code>ipFamily</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#ipfamily-v1-core">
Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamily is the IPFamily of the prefix.
If unset but Prefix is set, this can be inferred.</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Prefix is the prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>prefixLength</code><br/>
<em>
int32
</em>
</td>
<td>
<p>PrefixLength is the length of prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>parentRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>ParentRef references the parent to allocate the Prefix from.
If ParentRef and ParentSelector is empty, the Prefix is considered a root prefix and thus
allocated by itself.</p>
</td>
</tr>
<tr>
<td>
<code>parentSelector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>ParentSelector is the LabelSelector to use for determining the parent for this Prefix.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixStatus">PrefixStatus
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.Prefix">Prefix</a>)
</p>
<div>
<p>PrefixStatus defines the observed state of Prefix</p>
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
<code>conditions</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixCondition">
[]PrefixCondition
</a>
</em>
</td>
<td>
<p>Conditions is a list of conditions of a Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>used</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Used is a list of used prefixes.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.Readiness">Readiness
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
<tbody><tr><td><p>&#34;Failed&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Succeeded&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unknown&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>33fd61e</code>.
</em></p>
