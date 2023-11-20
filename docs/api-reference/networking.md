<p>Packages:</p>
<ul>
<li>
<a href="#networking.ironcore.dev%2fv1alpha1">networking.ironcore.dev/v1alpha1</a>
</li>
</ul>
<h2 id="networking.ironcore.dev/v1alpha1">networking.ironcore.dev/v1alpha1</h2>
<div>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</div>
Resource Types:
<ul><li>
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancer">LoadBalancer</a>
</li><li>
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerRouting">LoadBalancerRouting</a>
</li><li>
<a href="#networking.ironcore.dev/v1alpha1.NATGateway">NATGateway</a>
</li><li>
<a href="#networking.ironcore.dev/v1alpha1.Network">Network</a>
</li><li>
<a href="#networking.ironcore.dev/v1alpha1.NetworkInterface">NetworkInterface</a>
</li><li>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicy">NetworkPolicy</a>
</li><li>
<a href="#networking.ironcore.dev/v1alpha1.VirtualIP">VirtualIP</a>
</li></ul>
<h3 id="networking.ironcore.dev/v1alpha1.LoadBalancer">LoadBalancer
</h3>
<div>
<p>LoadBalancer is the Schema for the LoadBalancer API</p>
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
networking.ironcore.dev/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>LoadBalancer</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerSpec">
LoadBalancerSpec
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
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerType">
LoadBalancerType
</a>
</em>
</td>
<td>
<p>Type is the type of LoadBalancer.</p>
</td>
</tr>
<tr>
<td>
<code>ipFamilies</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
[]Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamilies are the ip families the load balancer should have.</p>
</td>
</tr>
<tr>
<td>
<code>ips</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.IPSource">
[]IPSource
</a>
</em>
</td>
<td>
<p>IPs are the ips to use. Can only be used when Type is LoadBalancerTypeInternal.</p>
</td>
</tr>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the Network this LoadBalancer should belong to.</p>
</td>
</tr>
<tr>
<td>
<code>networkInterfaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>NetworkInterfaceSelector defines the NetworkInterfaces
for which this LoadBalancer should be applied</p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerPort">
[]LoadBalancerPort
</a>
</em>
</td>
<td>
<p>Ports are the ports the load balancer should allow.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerStatus">
LoadBalancerStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.LoadBalancerRouting">LoadBalancerRouting
</h3>
<div>
<p>LoadBalancerRouting is the Schema for the loadbalancerroutings API</p>
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
networking.ironcore.dev/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>LoadBalancerRouting</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="../common/#common.ironcore.dev/v1alpha1.LocalUIDReference">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the network the load balancer is assigned to.</p>
</td>
</tr>
<tr>
<td>
<code>destinations</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerDestination">
[]LoadBalancerDestination
</a>
</em>
</td>
<td>
<p>Destinations are the destinations for an LoadBalancer.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NATGateway">NATGateway
</h3>
<div>
<p>NATGateway is the Schema for the NATGateway API</p>
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
networking.ironcore.dev/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>NATGateway</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="#networking.ironcore.dev/v1alpha1.NATGatewaySpec">
NATGatewaySpec
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
<a href="#networking.ironcore.dev/v1alpha1.NATGatewayType">
NATGatewayType
</a>
</em>
</td>
<td>
<p>Type is the type of NATGateway.</p>
</td>
</tr>
<tr>
<td>
<code>ipFamily</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamily is the ip family the NAT gateway should have.</p>
</td>
</tr>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the Network this NATGateway should belong to.</p>
</td>
</tr>
<tr>
<td>
<code>portsPerNetworkInterface</code><br/>
<em>
int32
</em>
</td>
<td>
<p>PortsPerNetworkInterface defines the number of concurrent connections per target network interface.
Has to be a power of 2. If empty, 2048 (DefaultPortsPerNetworkInterface) is the default.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NATGatewayStatus">
NATGatewayStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.Network">Network
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
networking.ironcore.dev/v1alpha1
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="#networking.ironcore.dev/v1alpha1.NetworkSpec">
NetworkSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<p>ProviderID is the provider-internal ID of the network.</p>
</td>
</tr>
<tr>
<td>
<code>peerings</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPeering">
[]NetworkPeering
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Peerings are the network peerings with this network.</p>
</td>
</tr>
<tr>
<td>
<code>incomingPeerings</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPeeringClaimRef">
[]NetworkPeeringClaimRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PeeringClaimRefs are the peering claim references of other networks.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkStatus">
NetworkStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkInterface">NetworkInterface
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
networking.ironcore.dev/v1alpha1
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceSpec">
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
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<p>ProviderID is the provider-internal ID of the network interface.</p>
</td>
</tr>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
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
<a href="../common/#common.ironcore.dev/v1alpha1.LocalUIDReference">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.LocalUIDReference
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
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
<a href="#networking.ironcore.dev/v1alpha1.IPSource">
[]IPSource
</a>
</em>
</td>
<td>
<p>IPs is the list of provided IPs or ephemeral IPs which should be assigned to
this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>prefixes</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.PrefixSource">
[]PrefixSource
</a>
</em>
</td>
<td>
<p>Prefixes is the list of provided prefixes or ephemeral prefixes which should be assigned to
this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>virtualIP</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPSource">
VirtualIPSource
</a>
</em>
</td>
<td>
<p>VirtualIP specifies the virtual ip that should be assigned to this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>attributes</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Attributes are provider-specific attributes for the network interface.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceStatus">
NetworkInterfaceStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicy">NetworkPolicy
</h3>
<div>
<p>NetworkPolicy is the Schema for the networkpolicies API</p>
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
networking.ironcore.dev/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>NetworkPolicy</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicySpec">
NetworkPolicySpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the network to regulate using this policy.</p>
</td>
</tr>
<tr>
<td>
<code>networkInterfaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>NetworkInterfaceSelector selects the network interfaces that are subject to this policy.</p>
</td>
</tr>
<tr>
<td>
<code>ingress</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyIngressRule">
[]NetworkPolicyIngressRule
</a>
</em>
</td>
<td>
<p>Ingress specifies rules for ingress traffic.</p>
</td>
</tr>
<tr>
<td>
<code>egress</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyEgressRule">
[]NetworkPolicyEgressRule
</a>
</em>
</td>
<td>
<p>Egress specifies rules for egress traffic.</p>
</td>
</tr>
<tr>
<td>
<code>policyTypes</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.PolicyType">
[]PolicyType
</a>
</em>
</td>
<td>
<p>PolicyTypes specifies the types of policies this network policy contains.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyStatus">
NetworkPolicyStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.VirtualIP">VirtualIP
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
networking.ironcore.dev/v1alpha1
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPSpec">
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
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPType">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
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
<a href="../common/#common.ironcore.dev/v1alpha1.LocalUIDReference">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.LocalUIDReference
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
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPStatus">
VirtualIPStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.EphemeralPrefixSource">EphemeralPrefixSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.IPSource">IPSource</a>, <a href="#networking.ironcore.dev/v1alpha1.PrefixSource">PrefixSource</a>)
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
<a href="../ipam/#ipam.ironcore.dev/v1alpha1.PrefixTemplateSpec">
github.com/ironcore-dev/ironcore/api/ipam/v1alpha1.PrefixTemplateSpec
</a>
</em>
</td>
<td>
<p>PrefixTemplate is the template for the Prefix.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.EphemeralVirtualIPSource">EphemeralVirtualIPSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.VirtualIPSource">VirtualIPSource</a>)
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
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPTemplateSpec">
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
<h3 id="networking.ironcore.dev/v1alpha1.IPBlock">IPBlock
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyPeer">NetworkPolicyPeer</a>)
</p>
<div>
<p>IPBlock specifies an ip block with optional exceptions.</p>
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
<code>cidr</code><br/>
<em>
<a href="../common/#common.ironcore.dev/v1alpha1.IPPrefix">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>CIDR is a string representing the ip block.</p>
</td>
</tr>
<tr>
<td>
<code>except</code><br/>
<em>
<a href="../common/#common.ironcore.dev/v1alpha1.IPPrefix">
[]github.com/ironcore-dev/ironcore/api/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Except is a slice of CIDRs that should not be included within the specified CIDR.
Values will be rejected if they are outside CIDR.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.IPSource">IPSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.LoadBalancerSpec">LoadBalancerSpec</a>, <a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceSpec">NetworkInterfaceSpec</a>)
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
<a href="../common/#common.ironcore.dev/v1alpha1.IP">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.IP
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
<a href="#networking.ironcore.dev/v1alpha1.EphemeralPrefixSource">
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
<h3 id="networking.ironcore.dev/v1alpha1.LoadBalancerDestination">LoadBalancerDestination
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.LoadBalancerRouting">LoadBalancerRouting</a>)
</p>
<div>
<p>LoadBalancerDestination is the destination of the load balancer.</p>
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
<a href="../common/#common.ironcore.dev/v1alpha1.IP">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IP is the target IP.</p>
</td>
</tr>
<tr>
<td>
<code>targetRef</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerTargetRef">
LoadBalancerTargetRef
</a>
</em>
</td>
<td>
<p>TargetRef is the target providing the destination.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.LoadBalancerPort">LoadBalancerPort
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.LoadBalancerSpec">LoadBalancerSpec</a>)
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
<code>protocol</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#protocol-v1-core">
Kubernetes core/v1.Protocol
</a>
</em>
</td>
<td>
<p>Protocol is the protocol the load balancer should allow.
If not specified, defaults to TCP.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Port is the port to allow.</p>
</td>
</tr>
<tr>
<td>
<code>endPort</code><br/>
<em>
int32
</em>
</td>
<td>
<p>EndPort marks the end of the port range to allow.
If unspecified, only a single port, Port, will be allowed.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.LoadBalancerSpec">LoadBalancerSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.LoadBalancer">LoadBalancer</a>)
</p>
<div>
<p>LoadBalancerSpec defines the desired state of LoadBalancer</p>
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
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerType">
LoadBalancerType
</a>
</em>
</td>
<td>
<p>Type is the type of LoadBalancer.</p>
</td>
</tr>
<tr>
<td>
<code>ipFamilies</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
[]Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamilies are the ip families the load balancer should have.</p>
</td>
</tr>
<tr>
<td>
<code>ips</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.IPSource">
[]IPSource
</a>
</em>
</td>
<td>
<p>IPs are the ips to use. Can only be used when Type is LoadBalancerTypeInternal.</p>
</td>
</tr>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the Network this LoadBalancer should belong to.</p>
</td>
</tr>
<tr>
<td>
<code>networkInterfaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>NetworkInterfaceSelector defines the NetworkInterfaces
for which this LoadBalancer should be applied</p>
</td>
</tr>
<tr>
<td>
<code>ports</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.LoadBalancerPort">
[]LoadBalancerPort
</a>
</em>
</td>
<td>
<p>Ports are the ports the load balancer should allow.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.LoadBalancerStatus">LoadBalancerStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.LoadBalancer">LoadBalancer</a>)
</p>
<div>
<p>LoadBalancerStatus defines the observed state of LoadBalancer</p>
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
<a href="../common/#common.ironcore.dev/v1alpha1.IP">
[]github.com/ironcore-dev/ironcore/api/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IPs are the IPs allocated for the load balancer.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.LoadBalancerTargetRef">LoadBalancerTargetRef
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.LoadBalancerDestination">LoadBalancerDestination</a>)
</p>
<div>
<p>LoadBalancerTargetRef is a load balancer target.</p>
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
<code>uid</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/types#UID">
k8s.io/apimachinery/pkg/types.UID
</a>
</em>
</td>
<td>
<p>UID is the UID of the target.</p>
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
<p>Name is the name of the target.</p>
</td>
</tr>
<tr>
<td>
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<p>ProviderID is the provider internal id of the target.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.LoadBalancerType">LoadBalancerType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.LoadBalancerSpec">LoadBalancerSpec</a>)
</p>
<div>
<p>LoadBalancerType is a type of LoadBalancer.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Internal&#34;</p></td>
<td><p>LoadBalancerTypeInternal is a LoadBalancer that allocates and routes network-internal, stable IPs.</p>
</td>
</tr><tr><td><p>&#34;Public&#34;</p></td>
<td><p>LoadBalancerTypePublic is a LoadBalancer that allocates and routes a stable public IP.</p>
</td>
</tr></tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NATGatewaySpec">NATGatewaySpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NATGateway">NATGateway</a>)
</p>
<div>
<p>NATGatewaySpec defines the desired state of NATGateway</p>
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
<a href="#networking.ironcore.dev/v1alpha1.NATGatewayType">
NATGatewayType
</a>
</em>
</td>
<td>
<p>Type is the type of NATGateway.</p>
</td>
</tr>
<tr>
<td>
<code>ipFamily</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
Kubernetes core/v1.IPFamily
</a>
</em>
</td>
<td>
<p>IPFamily is the ip family the NAT gateway should have.</p>
</td>
</tr>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the Network this NATGateway should belong to.</p>
</td>
</tr>
<tr>
<td>
<code>portsPerNetworkInterface</code><br/>
<em>
int32
</em>
</td>
<td>
<p>PortsPerNetworkInterface defines the number of concurrent connections per target network interface.
Has to be a power of 2. If empty, 2048 (DefaultPortsPerNetworkInterface) is the default.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NATGatewayStatus">NATGatewayStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NATGateway">NATGateway</a>)
</p>
<div>
<p>NATGatewayStatus defines the observed state of NATGateway</p>
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
<a href="../common/#common.ironcore.dev/v1alpha1.IP">
[]github.com/ironcore-dev/ironcore/api/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IPs are the IPs allocated for the NAT gateway.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NATGatewayType">NATGatewayType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NATGatewaySpec">NATGatewaySpec</a>)
</p>
<div>
<p>NATGatewayType is a type of NATGateway.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Public&#34;</p></td>
<td><p>NATGatewayTypePublic is a NATGateway that allocates and routes a stable public IP.</p>
</td>
</tr></tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkInterfaceSpec">NetworkInterfaceSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkInterface">NetworkInterface</a>, <a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceTemplateSpec">NetworkInterfaceTemplateSpec</a>)
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
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<p>ProviderID is the provider-internal ID of the network interface.</p>
</td>
</tr>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
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
<a href="../common/#common.ironcore.dev/v1alpha1.LocalUIDReference">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.LocalUIDReference
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
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
<a href="#networking.ironcore.dev/v1alpha1.IPSource">
[]IPSource
</a>
</em>
</td>
<td>
<p>IPs is the list of provided IPs or ephemeral IPs which should be assigned to
this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>prefixes</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.PrefixSource">
[]PrefixSource
</a>
</em>
</td>
<td>
<p>Prefixes is the list of provided prefixes or ephemeral prefixes which should be assigned to
this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>virtualIP</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPSource">
VirtualIPSource
</a>
</em>
</td>
<td>
<p>VirtualIP specifies the virtual ip that should be assigned to this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>attributes</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Attributes are provider-specific attributes for the network interface.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkInterfaceState">NetworkInterfaceState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceStatus">NetworkInterfaceStatus</a>)
</p>
<div>
<p>NetworkInterfaceState is the ironcore state of a NetworkInterface.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Available&#34;</p></td>
<td><p>NetworkInterfaceStateAvailable is used for any NetworkInterface where all properties are valid.</p>
</td>
</tr><tr><td><p>&#34;Error&#34;</p></td>
<td><p>NetworkInterfaceStateError is used for any NetworkInterface where any property has an error.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>NetworkInterfaceStatePending is used for any NetworkInterface that is pending.</p>
</td>
</tr></tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkInterfaceStatus">NetworkInterfaceStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkInterface">NetworkInterface</a>)
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
<code>state</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceState">
NetworkInterfaceState
</a>
</em>
</td>
<td>
<p>State is the NetworkInterfaceState of the NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>lastStateTransitionTime</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastStateTransitionTime is the last time the State transitioned from one value to another.</p>
</td>
</tr>
<tr>
<td>
<code>ips</code><br/>
<em>
<a href="../common/#common.ironcore.dev/v1alpha1.IP">
[]github.com/ironcore-dev/ironcore/api/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IPs represent the effective IP addresses of the NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>prefixes</code><br/>
<em>
<a href="../common/#common.ironcore.dev/v1alpha1.IPPrefix">
[]github.com/ironcore-dev/ironcore/api/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Prefixes represent the prefixes routed to the NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>virtualIP</code><br/>
<em>
<a href="../common/#common.ironcore.dev/v1alpha1.IP">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>VirtualIP is any virtual ip assigned to the NetworkInterface.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkInterfaceTemplateSpec">NetworkInterfaceTemplateSpec
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceSpec">
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
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<p>ProviderID is the provider-internal ID of the network interface.</p>
</td>
</tr>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
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
<a href="../common/#common.ironcore.dev/v1alpha1.LocalUIDReference">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.LocalUIDReference
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
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
<a href="#networking.ironcore.dev/v1alpha1.IPSource">
[]IPSource
</a>
</em>
</td>
<td>
<p>IPs is the list of provided IPs or ephemeral IPs which should be assigned to
this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>prefixes</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.PrefixSource">
[]PrefixSource
</a>
</em>
</td>
<td>
<p>Prefixes is the list of provided prefixes or ephemeral prefixes which should be assigned to
this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>virtualIP</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPSource">
VirtualIPSource
</a>
</em>
</td>
<td>
<p>VirtualIP specifies the virtual ip that should be assigned to this NetworkInterface.</p>
</td>
</tr>
<tr>
<td>
<code>attributes</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>Attributes are provider-specific attributes for the network interface.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPeering">NetworkPeering
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkSpec">NetworkSpec</a>)
</p>
<div>
<p>NetworkPeering defines a network peering with another network.</p>
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
<p>Name is the semantical name of the network peering.</p>
</td>
</tr>
<tr>
<td>
<code>networkRef</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPeeringNetworkRef">
NetworkPeeringNetworkRef
</a>
</em>
</td>
<td>
<p>NetworkRef is the reference to the network to peer with.
An empty namespace indicates that the target network resides in the same namespace as the source network.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPeeringClaimRef">NetworkPeeringClaimRef
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkSpec">NetworkSpec</a>)
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
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPeeringNetworkRef">NetworkPeeringNetworkRef
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPeering">NetworkPeering</a>)
</p>
<div>
<p>NetworkPeeringNetworkRef is a reference to a network to peer with.</p>
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
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPeeringStatus">NetworkPeeringStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkStatus">NetworkStatus</a>)
</p>
<div>
<p>NetworkPeeringStatus is the status of a network peering.</p>
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
<p>Name is the name of the network peering.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicyCondition">NetworkPolicyCondition
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyStatus">NetworkPolicyStatus</a>)
</p>
<div>
<p>NetworkPolicyCondition is one of the conditions of a network policy.</p>
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
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyConditionType">
NetworkPolicyConditionType
</a>
</em>
</td>
<td>
<p>Type is the type of the condition.</p>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#conditionstatus-v1-core">
Kubernetes core/v1.ConditionStatus
</a>
</em>
</td>
<td>
<p>Status is the status of the condition.</p>
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
<p>Reason is a machine-readable indication of why the condition is in a certain state.</p>
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
<p>Message is a human-readable explanation of why the condition has a certain reason / state.</p>
</td>
</tr>
<tr>
<td>
<code>observedGeneration</code><br/>
<em>
int64
</em>
</td>
<td>
<p>ObservedGeneration represents the .metadata.generation that the condition was set based upon.</p>
</td>
</tr>
<tr>
<td>
<code>lastTransitionTime</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastTransitionTime is the last time the status of a condition has transitioned from one state to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicyConditionType">NetworkPolicyConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyCondition">NetworkPolicyCondition</a>)
</p>
<div>
<p>NetworkPolicyConditionType is a type a NetworkPolicyCondition can have.</p>
</div>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicyEgressRule">NetworkPolicyEgressRule
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicySpec">NetworkPolicySpec</a>)
</p>
<div>
<p>NetworkPolicyEgressRule describes a rule to regulate egress traffic with.</p>
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
<code>ports</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyPort">
[]NetworkPolicyPort
</a>
</em>
</td>
<td>
<p>Ports specifies the list of destination ports that can be called with
this rule. Each item in this list is combined using a logical OR. Empty matches all ports.
As soon as a single item is present, only these ports are allowed.</p>
</td>
</tr>
<tr>
<td>
<code>to</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyPeer">
[]NetworkPolicyPeer
</a>
</em>
</td>
<td>
<p>To specifies the list of destinations which the selected network interfaces should be
able to send traffic to. Fields are combined using a logical OR. Empty matches all destinations.
As soon as a single item is present, only these peers are allowed.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicyIngressRule">NetworkPolicyIngressRule
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicySpec">NetworkPolicySpec</a>)
</p>
<div>
<p>NetworkPolicyIngressRule describes a rule to regulate ingress traffic with.</p>
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
<code>ports</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyPort">
[]NetworkPolicyPort
</a>
</em>
</td>
<td>
<p>Ports specifies the list of ports which should be made accessible for
this rule. Each item in this list is combined using a logical OR. Empty matches all ports.
As soon as a single item is present, only these ports are allowed.</p>
</td>
</tr>
<tr>
<td>
<code>from</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyPeer">
[]NetworkPolicyPeer
</a>
</em>
</td>
<td>
<p>From specifies the list of sources which should be able to send traffic to the
selected network interfaces. Fields are combined using a logical OR. Empty matches all sources.
As soon as a single item is present, only these peers are allowed.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicyPeer">NetworkPolicyPeer
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyEgressRule">NetworkPolicyEgressRule</a>, <a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyIngressRule">NetworkPolicyIngressRule</a>)
</p>
<div>
<p>NetworkPolicyPeer describes a peer to allow traffic to / from.</p>
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
<code>objectSelector</code><br/>
<em>
<a href="../core/#core.ironcore.dev/v1alpha1.ObjectSelector">
github.com/ironcore-dev/ironcore/api/core/v1alpha1.ObjectSelector
</a>
</em>
</td>
<td>
<p>ObjectSelector selects peers with the given kind matching the label selector.
Exclusive with other peer specifiers.</p>
</td>
</tr>
<tr>
<td>
<code>ipBlock</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.IPBlock">
IPBlock
</a>
</em>
</td>
<td>
<p>IPBlock specifies the ip block from or to which network traffic may come.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicyPort">NetworkPolicyPort
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyEgressRule">NetworkPolicyEgressRule</a>, <a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyIngressRule">NetworkPolicyIngressRule</a>)
</p>
<div>
<p>NetworkPolicyPort describes a port to allow traffic on</p>
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
<code>protocol</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#protocol-v1-core">
Kubernetes core/v1.Protocol
</a>
</em>
</td>
<td>
<p>Protocol (TCP, UDP, or SCTP) which traffic must match. If not specified, this
field defaults to TCP.</p>
</td>
</tr>
<tr>
<td>
<code>port</code><br/>
<em>
int32
</em>
</td>
<td>
<p>The port on the given protocol. If this field is not provided, this matches
all port names and numbers.
If present, only traffic on the specified protocol AND port will be matched.</p>
</td>
</tr>
<tr>
<td>
<code>endPort</code><br/>
<em>
int32
</em>
</td>
<td>
<p>EndPort indicates that the range of ports from Port to EndPort, inclusive,
should be allowed by the policy. This field cannot be defined if the port field
is not defined. The endPort must be equal or greater than port.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicySpec">NetworkPolicySpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicy">NetworkPolicy</a>)
</p>
<div>
<p>NetworkPolicySpec defines the desired state of NetworkPolicy.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkRef is the network to regulate using this policy.</p>
</td>
</tr>
<tr>
<td>
<code>networkInterfaceSelector</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>NetworkInterfaceSelector selects the network interfaces that are subject to this policy.</p>
</td>
</tr>
<tr>
<td>
<code>ingress</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyIngressRule">
[]NetworkPolicyIngressRule
</a>
</em>
</td>
<td>
<p>Ingress specifies rules for ingress traffic.</p>
</td>
</tr>
<tr>
<td>
<code>egress</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyEgressRule">
[]NetworkPolicyEgressRule
</a>
</em>
</td>
<td>
<p>Egress specifies rules for egress traffic.</p>
</td>
</tr>
<tr>
<td>
<code>policyTypes</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.PolicyType">
[]PolicyType
</a>
</em>
</td>
<td>
<p>PolicyTypes specifies the types of policies this network policy contains.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkPolicyStatus">NetworkPolicyStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicy">NetworkPolicy</a>)
</p>
<div>
<p>NetworkPolicyStatus defines the observed state of NetworkPolicy.</p>
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
<a href="#networking.ironcore.dev/v1alpha1.NetworkPolicyCondition">
[]NetworkPolicyCondition
</a>
</em>
</td>
<td>
<p>Conditions are various conditions of the NetworkPolicy.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkSpec">NetworkSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.Network">Network</a>)
</p>
<div>
<p>NetworkSpec defines the desired state of Network</p>
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
<code>providerID</code><br/>
<em>
string
</em>
</td>
<td>
<p>ProviderID is the provider-internal ID of the network.</p>
</td>
</tr>
<tr>
<td>
<code>peerings</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPeering">
[]NetworkPeering
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Peerings are the network peerings with this network.</p>
</td>
</tr>
<tr>
<td>
<code>incomingPeerings</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPeeringClaimRef">
[]NetworkPeeringClaimRef
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>PeeringClaimRefs are the peering claim references of other networks.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkState">NetworkState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkStatus">NetworkStatus</a>)
</p>
<div>
<p>NetworkState is the state of a network.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Available&#34;</p></td>
<td><p>NetworkStateAvailable means the network is ready to use.</p>
</td>
</tr><tr><td><p>&#34;Error&#34;</p></td>
<td><p>NetworkStateError means the network is in an error state.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>NetworkStatePending means the network is being provisioned.</p>
</td>
</tr></tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.NetworkStatus">NetworkStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.Network">Network</a>)
</p>
<div>
<p>NetworkStatus defines the observed state of Network</p>
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
<code>state</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkState">
NetworkState
</a>
</em>
</td>
<td>
<p>State is the state of the machine.</p>
</td>
</tr>
<tr>
<td>
<code>peerings</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.NetworkPeeringStatus">
[]NetworkPeeringStatus
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Peerings contains the states of the network peerings for the network.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.PolicyType">PolicyType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkPolicySpec">NetworkPolicySpec</a>)
</p>
<div>
<p>PolicyType is a type of policy.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Egress&#34;</p></td>
<td><p>PolicyTypeEgress is a policy that describes egress traffic.</p>
</td>
</tr><tr><td><p>&#34;Ingress&#34;</p></td>
<td><p>PolicyTypeIngress is a policy that describes ingress traffic.</p>
</td>
</tr></tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.PrefixSource">PrefixSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceSpec">NetworkInterfaceSpec</a>)
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
<a href="../common/#common.ironcore.dev/v1alpha1.IPPrefix">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.IPPrefix
</a>
</em>
</td>
<td>
<p>Value specifies a static prefix to use.</p>
</td>
</tr>
<tr>
<td>
<code>ephemeral</code><br/>
<em>
<a href="#networking.ironcore.dev/v1alpha1.EphemeralPrefixSource">
EphemeralPrefixSource
</a>
</em>
</td>
<td>
<p>Ephemeral specifies a prefix by creating an ephemeral ipam.Prefix to allocate the prefix with.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.VirtualIPSource">VirtualIPSource
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.NetworkInterfaceSpec">NetworkInterfaceSpec</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
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
<a href="#networking.ironcore.dev/v1alpha1.EphemeralVirtualIPSource">
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
<h3 id="networking.ironcore.dev/v1alpha1.VirtualIPSpec">VirtualIPSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.VirtualIP">VirtualIP</a>, <a href="#networking.ironcore.dev/v1alpha1.VirtualIPTemplateSpec">VirtualIPTemplateSpec</a>)
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
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPType">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
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
<a href="../common/#common.ironcore.dev/v1alpha1.LocalUIDReference">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>TargetRef references the target for this VirtualIP (currently only NetworkInterface).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.VirtualIPStatus">VirtualIPStatus
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.VirtualIP">VirtualIP</a>)
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
<a href="../common/#common.ironcore.dev/v1alpha1.IP">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.IP
</a>
</em>
</td>
<td>
<p>IP is the allocated IP, if any.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="networking.ironcore.dev/v1alpha1.VirtualIPTemplateSpec">VirtualIPTemplateSpec
</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.EphemeralVirtualIPSource">EphemeralVirtualIPSource</a>)
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">
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
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPSpec">
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
<a href="#networking.ironcore.dev/v1alpha1.VirtualIPType">
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#ipfamily-v1-core">
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
<a href="../common/#common.ironcore.dev/v1alpha1.LocalUIDReference">
github.com/ironcore-dev/ironcore/api/common/v1alpha1.LocalUIDReference
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
<h3 id="networking.ironcore.dev/v1alpha1.VirtualIPType">VirtualIPType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#networking.ironcore.dev/v1alpha1.VirtualIPSpec">VirtualIPSpec</a>)
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
