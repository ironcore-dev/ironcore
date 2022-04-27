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
<a href="#networking.api.onmetal.de/v1alpha1.Network">Network</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.NetworkInterface">NetworkInterface</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfaceBinding">NetworkInterfaceBinding</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIP">VirtualIP</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPRouting">VirtualIPRouting</a>
</li></ul>
<h3 id="networking.api.onmetal.de/v1alpha1.Network">Network
</h3>
<div>
<p>Network is the Schema for the network API</p>
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
<td><code>Network</code></td>
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
</tbody>
</table>
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
<h3 id="networking.api.onmetal.de/v1alpha1.NetworkInterfaceBinding">NetworkInterfaceBinding
</h3>
<div>
<p>NetworkInterfaceBinding is the Schema for the networkinterfacebindings API</p>
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
<td><code>NetworkInterfaceBinding</code></td>
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
<code>ips</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIP">VirtualIP
</h3>
<div>
<p>VirtualIP is the Schema for the virtualips API</p>
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
<td><code>VirtualIP</code></td>
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
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPSpec">
VirtualIPSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>type</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPType">
VirtualIPType
</a>
</em>
</td>
<td>
<p>Type is the type of VirtualIP.</p>
</td>
</tr>
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
<p>IPFamily is the ip family of the VirtualIP.</p>
</td>
</tr>
<tr>
<td>
<code>networkInterfaceSelector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>NetworkInterfaceSelector selects any NetworkInterface that should get the VirtualIP routed.
If empty, it is assumed that an external process manages the VirtualIPRouting for this VirtualIP.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPStatus">
VirtualIPStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPRouting">VirtualIPRouting
</h3>
<div>
<p>VirtualIPRouting is the Schema for the virtualiproutings API</p>
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
<td><code>VirtualIPRouting</code></td>
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
<code>subsets</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPRoutingSubset">
[]VirtualIPRoutingSubset
</a>
</em>
</td>
<td>
<p>Subsets are the subsets that make up a VirtualIPRouting.</p>
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
github.com/onmetal/onmetal-api/apis/ipam/v1alpha1.PrefixTemplateSpec
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
<h3 id="networking.api.onmetal.de/v1alpha1.LocalUIDReference">LocalUIDReference
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPRoutingSubset">VirtualIPRoutingSubset</a>, <a href="#networking.api.onmetal.de/v1alpha1.VirtualIPRoutingSubsetTarget">VirtualIPRoutingSubsetTarget</a>)
</p>
<div>
<p>LocalUIDReference is a reference to another entity including its UID.</p>
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
k8s.io/apimachinery/pkg/types.UID
</em>
</td>
<td>
<p>UID is the UID of the referenced entity.</p>
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
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPRoutingSubset">VirtualIPRoutingSubset
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPRouting">VirtualIPRouting</a>)
</p>
<div>
<p>VirtualIPRoutingSubset is one of the targets of a VirtualIPRouting.</p>
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
<a href="#networking.api.onmetal.de/v1alpha1.LocalUIDReference">
LocalUIDReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the network all targets are in.</p>
</td>
</tr>
<tr>
<td>
<code>targets</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPRoutingSubsetTarget">
[]VirtualIPRoutingSubsetTarget
</a>
</em>
</td>
<td>
<p>Targets are the targets of the virtual IP.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPRoutingSubsetTarget">VirtualIPRoutingSubsetTarget
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPRoutingSubset">VirtualIPRoutingSubset</a>)
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
<code>LocalUIDReference</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.LocalUIDReference">
LocalUIDReference
</a>
</em>
</td>
<td>
<p>
(Members of <code>LocalUIDReference</code> are embedded into this type.)
</p>
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
<p>IP is the target ip to route to.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPSpec">VirtualIPSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIP">VirtualIP</a>)
</p>
<div>
<p>VirtualIPSpec defines the desired state of VirtualIP</p>
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
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPType">
VirtualIPType
</a>
</em>
</td>
<td>
<p>Type is the type of VirtualIP.</p>
</td>
</tr>
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
<p>IPFamily is the ip family of the VirtualIP.</p>
</td>
</tr>
<tr>
<td>
<code>networkInterfaceSelector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>NetworkInterfaceSelector selects any NetworkInterface that should get the VirtualIP routed.
If empty, it is assumed that an external process manages the VirtualIPRouting for this VirtualIP.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPStatus">VirtualIPStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIP">VirtualIP</a>)
</p>
<div>
<p>VirtualIPStatus defines the observed state of VirtualIP</p>
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
<code>ip</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IP is the allocated IP, if any.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPType">VirtualIPType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPSpec">VirtualIPSpec</a>)
</p>
<div>
<p>VirtualIPType is a type of VirtualIP.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Public&#34;</p></td>
<td><p>VirtualIPTypePublic is a VirtualIP that allocates and routes a stable public IP.</p>
</td>
</tr></tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
</em></p>
