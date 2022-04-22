<p>Packages:</p>
<ul>
<li>
<a href="#networking.api.onmetal.de%2fv1alpha1">networking.api.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="networking.api.onmetal.de/v1alpha1">networking.api.onmetal.de/v1alpha1</h2>
<div>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</div>
Resource Types:
<ul><li>
<a href="#networking.api.onmetal.de/v1alpha1.NetworkInterface">NetworkInterface</a>
</li></ul>
<h3 id="networking.api.onmetal.de/v1alpha1.NetworkInterface">NetworkInterface
</h3>
<div>
<p>NetworkInterface is the Schema for the networkinterfaces API</p>
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
networking.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>NetworkInterface</code></td>
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
<a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfaceSpec">
NetworkInterfaceSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the Network this NetworkInterface is connected to</p>
</td>
</tr>
<tr>
<td>
<code>machineRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachineRef is the Machine this NetworkInterface is used by</p>
</td>
</tr>
<tr>
<td>
<code>ipFamilies</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#ipfamily-v1-core">
[]Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamilies defines which IPFamilies this NetworkInterface is supporting</p>
</td>
</tr>
<tr>
<td>
<code>ips</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.IPSource">
[]IPSource
</a>
</em>
</td>
<td>
<p>IPs is the list of provided IPs or EphemeralIPs which should be assigned to
this NetworkInterface</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfaceStatus">
NetworkInterfaceStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.EphemeralPrefixSource">EphemeralPrefixSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.IPSource">IPSource</a>)
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
<code>prefixTemplate</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.PrefixTemplate">
PrefixTemplate
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.IPSource">IPSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfaceSpec">NetworkInterfaceSpec</a>)
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
<code>value</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>ephemeralPrefix</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.EphemeralPrefixSource">
EphemeralPrefixSource
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.NetworkInterfaceSpec">NetworkInterfaceSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.NetworkInterface">NetworkInterface</a>)
</p>
<div>
<p>NetworkInterfaceSpec defines the desired state of NetworkInterface</p>
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
<code>networkRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the Network this NetworkInterface is connected to</p>
</td>
</tr>
<tr>
<td>
<code>machineRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachineRef is the Machine this NetworkInterface is used by</p>
</td>
</tr>
<tr>
<td>
<code>ipFamilies</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#ipfamily-v1-core">
[]Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamilies defines which IPFamilies this NetworkInterface is supporting</p>
</td>
</tr>
<tr>
<td>
<code>ips</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.IPSource">
[]IPSource
</a>
</em>
</td>
<td>
<p>IPs is the list of provided IPs or EphemeralIPs which should be assigned to
this NetworkInterface</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.NetworkInterfaceStatus">NetworkInterfaceStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.NetworkInterface">NetworkInterface</a>)
</p>
<div>
<p>NetworkInterfaceStatus defines the observed state of NetworkInterface</p>
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
<code>ips</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>TODO: Add State, Conditions
IPs represent the effective IP addresses of the NetworkInterface</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.PrefixTemplate">PrefixTemplate
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.EphemeralPrefixSource">EphemeralPrefixSource</a>)
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
github.com/onmetal/onmetal-api/apis/ipam/v1alpha1.PrefixSpec
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
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>965c4f3</code>.
</em></p>
