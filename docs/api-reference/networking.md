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
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIP">VirtualIP</a>
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
<p>Prefix is the provided Prefix or Ephemeral which
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
<code>destinations</code><br/>
<em>
<a href="/api-reference/common.md#localuidreference">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>Destinations are the destinations for an AliasPrefix.</p>
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
<a href="/api-reference/common.md#localuidreference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
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
<code>virtualIP</code><br/>
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
<code>targetRef</code><br/>
<em>
<a href="/api-reference/common.md#localuidreference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>TargetRef references the target for this VirtualIP (currently only NetworkInterface).</p>
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
<p>Prefix is the provided Prefix or Ephemeral which
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
<a href="/api-reference/common.md#ip">
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
<p>EphemeralPrefixSource contains the definition to create an ephemeral (i.e. coupled to the lifetime of the
surrounding object) Prefix.</p>
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
<a href="/api-reference/ipam.md#ipam.api.onmetal.de/v1alpha1.PrefixTemplateSpec">
github.com/onmetal/onmetal-api/apis/ipam/v1alpha1.PrefixTemplateSpec
</a>
</em>
</td>
<td>
<p>PrefixTemplate is the template for the Prefix.</p>
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
<p>EphemeralVirtualIPSource contains the definition to create an ephemeral (i.e. coupled to the lifetime of the
surrounding object) VirtualIP.</p>
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
<code>virtualIPTemplate</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPTemplateSpec">
VirtualIPTemplateSpec
</a>
</em>
</td>
<td>
<p>VirtualIPTemplate is the template for the VirtualIP.</p>
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
<p>IPSource is the definition of how to obtain an IP.</p>
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
<a href="/api-reference/common.md#ip">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>Value specifies an IP by using an IP literal.</p>
</td>
</tr>
<tr>
<td>
<code>ephemeral</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.EphemeralPrefixSource">
EphemeralPrefixSource
</a>
</em>
</td>
<td>
<p>Ephemeral specifies an IP by creating an ephemeral Prefix to allocate the IP with.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.NetworkInterfacePhase">NetworkInterfacePhase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfaceStatus">NetworkInterfaceStatus</a>)
</p>
<div>
<p>NetworkInterfacePhase is the binding phase of a NetworkInterface.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Bound&#34;</p></td>
<td><p>NetworkInterfacePhaseBound is used for any NetworkInterface that is properly bound.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>NetworkInterfacePhasePending is used for any NetworkInterface that is currently awaiting binding.</p>
</td>
</tr><tr><td><p>&#34;Unbound&#34;</p></td>
<td><p>NetworkInterfacePhaseUnbound is used for any NetworkInterface that is not bound.</p>
</td>
</tr></tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.NetworkInterfaceSpec">NetworkInterfaceSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.NetworkInterface">NetworkInterface</a>, <a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfaceTemplateSpec">NetworkInterfaceTemplateSpec</a>)
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
<a href="/api-reference/common.md#localuidreference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
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
<code>virtualIP</code><br/>
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
<a href="/api-reference/common.md#ip">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IPs represent the effective IP addresses of the NetworkInterface</p>
</td>
</tr>
<tr>
<td>
<code>virtualIP</code><br/>
<em>
<a href="/api-reference/common.md#ip">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>VirtualIP is any virtual ip assigned to the NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfacePhase">
NetworkInterfacePhase
</a>
</em>
</td>
<td>
<p>Phase is the NetworkInterfacePhase of the NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>phaseLastTransitionTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastPhaseTransitionTime is the last time the Phase transitioned from one value to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.NetworkInterfaceTemplateSpec">NetworkInterfaceTemplateSpec
</h3>
<div>
<p>NetworkInterfaceTemplateSpec is the specification of a NetworkInterface template.</p>
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
<a href="/api-reference/common.md#localuidreference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
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
<code>virtualIP</code><br/>
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
<a href="/api-reference/common.md#ip">
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
<code>ephemeral</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.EphemeralPrefixSource">
EphemeralPrefixSource
</a>
</em>
</td>
<td>
<p>Ephemeral defines the Prefix which should be allocated by the AliasPrefix</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPPhase">VirtualIPPhase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIPStatus">VirtualIPStatus</a>)
</p>
<div>
<p>VirtualIPPhase is the binding phase of a VirtualIP.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Bound&#34;</p></td>
<td><p>VirtualIPPhaseBound is used for any VirtualIP that is properly bound.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>VirtualIPPhasePending is used for any VirtualIP that is currently awaiting binding.</p>
</td>
</tr><tr><td><p>&#34;Unbound&#34;</p></td>
<td><p>VirtualIPPhaseUnbound is used for any VirtualIP that is not bound.</p>
</td>
</tr></tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPSource">VirtualIPSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.NetworkInterfaceSpec">NetworkInterfaceSpec</a>)
</p>
<div>
<p>VirtualIPSource is the definition of how to obtain a VirtualIP.</p>
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
<code>virtualIPRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VirtualIPRef references a VirtualIP to use.</p>
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
<p>Ephemeral instructs to create an ephemeral (i.e. coupled to the lifetime of the surrounding object)
VirtualIP.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPSpec">VirtualIPSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.VirtualIP">VirtualIP</a>, <a href="#networking.api.onmetal.de/v1alpha1.VirtualIPTemplateSpec">VirtualIPTemplateSpec</a>)
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
<code>targetRef</code><br/>
<em>
<a href="/api-reference/common.md#localuidreference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>TargetRef references the target for this VirtualIP (currently only NetworkInterface).</p>
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
<a href="/api-reference/common.md#ip">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IP is the allocated IP, if any.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code><br/>
<em>
<a href="#networking.api.onmetal.de/v1alpha1.VirtualIPPhase">
VirtualIPPhase
</a>
</em>
</td>
<td>
<p>Phase is the VirtualIPPhase of the VirtualIP.</p>
</td>
</tr>
<tr>
<td>
<code>phaseLastTransitionTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastPhaseTransitionTime is the last time the Phase transitioned from one value to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.api.onmetal.de/v1alpha1.VirtualIPTemplateSpec">VirtualIPTemplateSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.api.onmetal.de/v1alpha1.EphemeralVirtualIPSource">EphemeralVirtualIPSource</a>)
</p>
<div>
<p>VirtualIPTemplateSpec is the specification of a VirtualIP template.</p>
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
<code>targetRef</code><br/>
<em>
<a href="/api-reference/common.md#localuidreference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>TargetRef references the target for this VirtualIP (currently only NetworkInterface).</p>
</td>
</tr>
</table>
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
