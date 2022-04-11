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
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefix">ClusterPrefix</a>
</li><li>
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocation">ClusterPrefixAllocation</a>
</li><li>
<a href="#ipam.api.onmetal.de/v1alpha1.IP">IP</a>
</li><li>
<a href="#ipam.api.onmetal.de/v1alpha1.Prefix">Prefix</a>
</li><li>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocation">PrefixAllocation</a>
</li></ul>
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefix">ClusterPrefix
</h3>
<div>
<p>ClusterPrefix is the Schema for the clusterprefixes API</p>
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
<td><code>ClusterPrefix</code></td>
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
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixSpec">
ClusterPrefixSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
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
<tr>
<td>
<code>PrefixSpace</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSpace">
PrefixSpace
</a>
</em>
</td>
<td>
<p>
(Members of <code>PrefixSpace</code> are embedded into this type.)
</p>
<p>PrefixSpace is the space the ClusterPrefix manages.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixStatus">
ClusterPrefixStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocation">ClusterPrefixAllocation
</h3>
<div>
<p>ClusterPrefixAllocation is the Schema for the clusterprefixallocations API</p>
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
<td><code>ClusterPrefixAllocation</code></td>
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
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationSpec">
ClusterPrefixAllocationSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
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
</td>
</tr>
<tr>
<td>
<code>ClusterPrefixAllocationRequest</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationRequest">
ClusterPrefixAllocationRequest
</a>
</em>
</td>
<td>
<p>
(Members of <code>ClusterPrefixAllocationRequest</code> are embedded into this type.)
</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationStatus">
ClusterPrefixAllocationStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.IP">IP
</h3>
<div>
<p>IP is the Schema for the ips API</p>
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
<td><code>IP</code></td>
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
<a href="#ipam.api.onmetal.de/v1alpha1.IPSpec">
IPSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>prefixRef</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixReference">
PrefixReference
</a>
</em>
</td>
<td>
<p>PrefixRef references the parent to allocate the IP from.</p>
</td>
</tr>
<tr>
<td>
<code>prefixSelector</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSelector">
PrefixSelector
</a>
</em>
</td>
<td>
<p>PrefixSelector is the LabelSelector to use for determining the parent for this IP.</p>
</td>
</tr>
<tr>
<td>
<code>ip</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IP is the ip to allocate.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.IPStatus">
IPStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
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
<code>parentRef</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixReference">
PrefixReference
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
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSelector">
PrefixSelector
</a>
</em>
</td>
<td>
<p>ParentSelector is the LabelSelector to use for determining the parent for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>PrefixSpace</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSpace">
PrefixSpace
</a>
</em>
</td>
<td>
<p>
(Members of <code>PrefixSpace</code> are embedded into this type.)
</p>
<p>PrefixSpace is the definition of the space the prefix manages.</p>
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
<code>prefixRef</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixReference">
PrefixReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PrefixRef references the prefix to allocate from.</p>
</td>
</tr>
<tr>
<td>
<code>prefixSelector</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSelector">
PrefixSelector
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>PrefixAllocationRequest</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationRequest">
PrefixAllocationRequest
</a>
</em>
</td>
<td>
<p>
(Members of <code>PrefixAllocationRequest</code> are embedded into this type.)
</p>
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
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationCondition">ClusterPrefixAllocationCondition
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationStatus">ClusterPrefixAllocationStatus</a>)
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
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationConditionType">
ClusterPrefixAllocationConditionType
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
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationConditionType">ClusterPrefixAllocationConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationCondition">ClusterPrefixAllocationCondition</a>)
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
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationRequest">ClusterPrefixAllocationRequest
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationSpec">ClusterPrefixAllocationSpec</a>)
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
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
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
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationResult">ClusterPrefixAllocationResult
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationStatus">ClusterPrefixAllocationStatus</a>)
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
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationSpec">ClusterPrefixAllocationSpec
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocation">ClusterPrefixAllocation</a>)
</p>
<div>
<p>ClusterPrefixAllocationSpec defines the desired state of ClusterPrefixAllocation</p>
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
<code>prefixRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
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
</td>
</tr>
<tr>
<td>
<code>ClusterPrefixAllocationRequest</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationRequest">
ClusterPrefixAllocationRequest
</a>
</em>
</td>
<td>
<p>
(Members of <code>ClusterPrefixAllocationRequest</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationStatus">ClusterPrefixAllocationStatus
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocation">ClusterPrefixAllocation</a>)
</p>
<div>
<p>ClusterPrefixAllocationStatus defines the observed state of ClusterPrefixAllocation</p>
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
<code>ClusterPrefixAllocationResult</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationResult">
ClusterPrefixAllocationResult
</a>
</em>
</td>
<td>
<p>
(Members of <code>ClusterPrefixAllocationResult</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixAllocationCondition">
[]ClusterPrefixAllocationCondition
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixCondition">ClusterPrefixCondition
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixStatus">ClusterPrefixStatus</a>)
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
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixConditionType">
ClusterPrefixConditionType
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
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixConditionType">ClusterPrefixConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixCondition">ClusterPrefixCondition</a>)
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
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixSpec">ClusterPrefixSpec
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefix">ClusterPrefix</a>)
</p>
<div>
<p>ClusterPrefixSpec defines the desired state of ClusterPrefix</p>
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
<tr>
<td>
<code>PrefixSpace</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSpace">
PrefixSpace
</a>
</em>
</td>
<td>
<p>
(Members of <code>PrefixSpace</code> are embedded into this type.)
</p>
<p>PrefixSpace is the space the ClusterPrefix manages.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.ClusterPrefixStatus">ClusterPrefixStatus
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefix">ClusterPrefix</a>)
</p>
<div>
<p>ClusterPrefixStatus defines the observed state of ClusterPrefix</p>
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
<a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixCondition">
[]ClusterPrefixCondition
</a>
</em>
</td>
<td>
<p>Conditions is a list of conditions of a ClusterPrefix.</p>
</td>
</tr>
<tr>
<td>
<code>available</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Available is a list of available prefixes.</p>
</td>
</tr>
<tr>
<td>
<code>reserved</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Reserved is a list of reserved prefixes.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.IPCondition">IPCondition
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.IPStatus">IPStatus</a>)
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
<a href="#ipam.api.onmetal.de/v1alpha1.IPConditionType">
IPConditionType
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
<h3 id="ipam.api.onmetal.de/v1alpha1.IPConditionType">IPConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.IPCondition">IPCondition</a>)
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
<h3 id="ipam.api.onmetal.de/v1alpha1.IPSpec">IPSpec
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.IP">IP</a>)
</p>
<div>
<p>IPSpec defines the desired state of IP</p>
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
<code>prefixRef</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixReference">
PrefixReference
</a>
</em>
</td>
<td>
<p>PrefixRef references the parent to allocate the IP from.</p>
</td>
</tr>
<tr>
<td>
<code>prefixSelector</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSelector">
PrefixSelector
</a>
</em>
</td>
<td>
<p>PrefixSelector is the LabelSelector to use for determining the parent for this IP.</p>
</td>
</tr>
<tr>
<td>
<code>ip</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IP is the ip to allocate.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.IPStatus">IPStatus
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.IP">IP</a>)
</p>
<div>
<p>IPStatus defines the observed state of IP</p>
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
<a href="#ipam.api.onmetal.de/v1alpha1.IPCondition">
[]IPCondition
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
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixAllocationRequest">PrefixAllocationRequest
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationSpec">PrefixAllocationSpec</a>)
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
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
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
</td>
</tr>
<tr>
<td>
<code>range</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPRange
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>rangeLength</code><br/>
<em>
int64
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixAllocationResult">PrefixAllocationResult
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
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>range</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPRange
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
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
<code>prefixRef</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixReference">
PrefixReference
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PrefixRef references the prefix to allocate from.</p>
</td>
</tr>
<tr>
<td>
<code>prefixSelector</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSelector">
PrefixSelector
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>PrefixAllocationRequest</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationRequest">
PrefixAllocationRequest
</a>
</em>
</td>
<td>
<p>
(Members of <code>PrefixAllocationRequest</code> are embedded into this type.)
</p>
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
<code>PrefixAllocationResult</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationResult">
PrefixAllocationResult
</a>
</em>
</td>
<td>
<p>
(Members of <code>PrefixAllocationResult</code> are embedded into this type.)
</p>
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
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixReference">PrefixReference
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.IPSpec">IPSpec</a>, <a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationSpec">PrefixAllocationSpec</a>, <a href="#ipam.api.onmetal.de/v1alpha1.PrefixSpec">PrefixSpec</a>)
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
<code>kind</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Kind is the kind of prefix to select.</p>
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
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixSelector">PrefixSelector
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.IPSpec">IPSpec</a>, <a href="#ipam.api.onmetal.de/v1alpha1.PrefixAllocationSpec">PrefixAllocationSpec</a>, <a href="#ipam.api.onmetal.de/v1alpha1.PrefixSpec">PrefixSpec</a>)
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
<code>kind</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Kind is the kind of prefix to select.</p>
</td>
</tr>
<tr>
<td>
<code>LabelSelector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>
(Members of <code>LabelSelector</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="ipam.api.onmetal.de/v1alpha1.PrefixSpace">PrefixSpace
</h3>
<p>
(<em>Appears on:</em><a href="#ipam.api.onmetal.de/v1alpha1.ClusterPrefixSpec">ClusterPrefixSpec</a>, <a href="#ipam.api.onmetal.de/v1alpha1.PrefixSpec">PrefixSpec</a>)
</p>
<div>
<p>PrefixSpace is the space a prefix manages.</p>
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
<code>prefix</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Prefix is the prefix to allocate for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>reservations</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Reservations is a list of IPPrefixes to reserve for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>reservationLengths</code><br/>
<em>
[]int32
</em>
</td>
<td>
<p>ReservationLengths is a list of IPPrefixes to reserve for this Prefix.</p>
</td>
</tr>
</tbody>
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
<code>parentRef</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixReference">
PrefixReference
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
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSelector">
PrefixSelector
</a>
</em>
</td>
<td>
<p>ParentSelector is the LabelSelector to use for determining the parent for this Prefix.</p>
</td>
</tr>
<tr>
<td>
<code>PrefixSpace</code><br/>
<em>
<a href="#ipam.api.onmetal.de/v1alpha1.PrefixSpace">
PrefixSpace
</a>
</em>
</td>
<td>
<p>
(Members of <code>PrefixSpace</code> are embedded into this type.)
</p>
<p>PrefixSpace is the definition of the space the prefix manages.</p>
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
<code>available</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Available is a list of available prefixes.</p>
</td>
</tr>
<tr>
<td>
<code>reserved</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Reserved is a list of reserved prefixes.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>23732c2</code>.
</em></p>
