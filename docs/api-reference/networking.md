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
<a href="#networking.api.onmetal.de/v1alpha1.AliasPrefix">AliasPrefix</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.AliasPrefixRouting">AliasPrefixRouting</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.Network">Network</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.NetworkInterface">NetworkInterface</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfaceBinding">NetworkInterfaceBinding</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIP">VirtualIP</a>
</li><li>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaim">VirtualIPClaim</a>
</li></ul>
<h3 id="networking.api.onmetal.de/v1alpha1.AliasPrefix">AliasPrefix
</h3>
<div>
<p>AliasPrefix is the Schema for the AliasPrefix API</p>
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
<td><code>AliasPrefix</code></td>
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
<a href="#networking.api.onmetal.de/v1alpha1.AliasPrefixSpec">
AliasPrefixSpec
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
<p>NetworkRef is the Network this AliasPrefix should belong to</p>
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
<p>NetworkInterfaceSelector defines the NetworkInterfaces
for which this AliasPrefix should be applied</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.PrefixSource">
PrefixSource
</a>
</em>
</td>
<td>
<p>Prefix is the provided Prefix or EphemeralPrefix which
should be used by this AliasPrefix</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.AliasPrefixStatus">
AliasPrefixStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.AliasPrefixRouting">AliasPrefixRouting
</h3>
<div>
<p>AliasPrefixRouting is the Schema for the aliasprefixrouting API</p>
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
<td><code>AliasPrefixRouting</code></td>
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
<code>networkRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the Network this AliasPrefixRouting should belong to</p>
</td>
</tr>
<tr>
<td>
<code>subsets</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.AliasPrefixRoutingSubset">
[]AliasPrefixRoutingSubset
</a>
</em>
</td>
<td>
<p>Subsets are the subsets that make up an AliasPrefixRouting</p>
</td>
</tr>
</tbody>
</table>
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
<tr>
<td>
<code>virtualIp</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPSource">
VirtualIPSource
</a>
</em>
</td>
<td>
<p>VirtualIP specifies the virtual ip that should be assigned to this NetworkInterface.</p>
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
<tr>
<td>
<code>virtualIPRef</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.LocalUIDReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
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
<code>claimRef</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.LocalUIDReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>ClaimRef references the VirtualIPClaim that claimed this virtual ip.</p>
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
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPClaim">VirtualIPClaim
</h3>
<div>
<p>VirtualIPClaim is the Schema for the virtualipclaims API</p>
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
<td><code>VirtualIPClaim</code></td>
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
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaimSpec">
VirtualIPClaimSpec
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
<code>virtualIPRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VirtualIPRef references the virtual ip to claim.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaimStatus">
VirtualIPClaimStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.AliasPrefixRoutingSubset">AliasPrefixRoutingSubset
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.AliasPrefixRouting">AliasPrefixRouting</a>)
</p>
<div>
<p>AliasPrefixRoutingSubset is one of the targets of a AliasPrefixRouting</p>
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
<code>machinePoolRef</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.LocalUIDReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>MachinePoolRef is the machine pool hosting the targeted entities.</p>
</td>
</tr>
<tr>
<td>
<code>targets</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.LocalUIDReference">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>Targets are the entities targeted by the alias prefix routing.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.AliasPrefixSpec">AliasPrefixSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.AliasPrefix">AliasPrefix</a>)
</p>
<div>
<p>AliasPrefixSpec defines the desired state of AliasPrefix</p>
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
<p>NetworkRef is the Network this AliasPrefix should belong to</p>
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
<p>NetworkInterfaceSelector defines the NetworkInterfaces
for which this AliasPrefix should be applied</p>
</td>
</tr>
<tr>
<td>
<code>prefix</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.PrefixSource">
PrefixSource
</a>
</em>
</td>
<td>
<p>Prefix is the provided Prefix or EphemeralPrefix which
should be used by this AliasPrefix</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.AliasPrefixStatus">AliasPrefixStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.AliasPrefix">AliasPrefix</a>)
</p>
<div>
<p>AliasPrefixStatus defines the observed state of AliasPrefix</p>
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
<p>Prefix is the Prefix reserved by this AliasPrefix</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.EphemeralPrefixSource">EphemeralPrefixSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.IPSource">IPSource</a>, <a href="#networking.api.onmetal.de/v1alpha1.PrefixSource">PrefixSource</a>)
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
<a href="/api-reference/common/#ipam.onmetal.de/v1alpha1.PrefixTemplateSpec">
github.com/onmetal/onmetal-api/apis/ipam/v1alpha1.PrefixTemplateSpec
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.EphemeralVirtualIPSource">EphemeralVirtualIPSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPSource">VirtualIPSource</a>)
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
<code>virtualIPClaimTemplate</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaimTemplateSpec">
VirtualIPClaimTemplateSpec
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
<tr>
<td>
<code>virtualIp</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPSource">
VirtualIPSource
</a>
</em>
</td>
<td>
<p>VirtualIP specifies the virtual ip that should be assigned to this NetworkInterface.</p>
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
<p>IPs represent the effective IP addresses of the NetworkInterface</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.PrefixSource">PrefixSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.AliasPrefixSpec">AliasPrefixSpec</a>)
</p>
<div>
<p>PrefixSource is the source of the Prefix definition in an AliasPrefix</p>
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
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Value is a single IPPrefix value as defined in the AliasPrefix</p>
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
<p>EphemeralPrefix defines the Prefix which should be allocated by the AliasPrefix</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPClaimSpec">VirtualIPClaimSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaim">VirtualIPClaim</a>, <a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaimTemplateSpec">VirtualIPClaimTemplateSpec</a>)
</p>
<div>
<p>VirtualIPClaimSpec defines the desired state of VirtualIPClaim</p>
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
<code>virtualIPRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VirtualIPRef references the virtual ip to claim.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPClaimStatus">VirtualIPClaimStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaim">VirtualIPClaim</a>)
</p>
<div>
<p>VirtualIPClaimStatus defines the observed state of VirtualIPClaim</p>
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
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPClaimTemplateSpec">VirtualIPClaimTemplateSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.EphemeralVirtualIPSource">EphemeralVirtualIPSource</a>)
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
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaimSpec">
VirtualIPClaimSpec
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
<code>virtualIPRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VirtualIPRef references the virtual ip to claim.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPSource">VirtualIPSource
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
<code>virtualIPClaimRef</code><br/>
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
<code>ephemeral</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.EphemeralVirtualIPSource">
EphemeralVirtualIPSource
</a>
</em>
</td>
<td>
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
<code>claimRef</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.LocalUIDReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>ClaimRef references the VirtualIPClaim that claimed this virtual ip.</p>
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
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPClaimSpec">VirtualIPClaimSpec</a>, <a href="#networking.api.onmetal.de/v1alpha1.VirtualIPSpec">VirtualIPSpec</a>)
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
