<p>Packages:</p>
<ul>
<li>
<a href="#network.onmetal.de%2fv1alpha1">network.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="network.onmetal.de/v1alpha1">network.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="network.onmetal.de/v1alpha1.ActionType">ActionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroupRule">SecurityGroupRule</a>)
</p>
<div>
<p>ActionType describes the action type of a SecurityGroupRule</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;allowed&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;deny&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.CIDRStatus">CIDRStatus
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SubnetStatus">SubnetStatus</a>)
</p>
<div>
<p>CIDRStatus is the status of a CIDR</p>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Cidr">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.Cidr
</a>
</em>
</td>
<td>
<p>CIDR defines the cidr</p>
</td>
</tr>
<tr>
<td>
<code>blockedRanges</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Cidr">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.Cidr
</a>
</em>
</td>
<td>
<p>BlockedRanges is a list of blocked cidr ranges</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.EgressSecurityGroupRule">EgressSecurityGroupRule
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroupSpec">SecurityGroupSpec</a>)
</p>
<div>
<p>EgressSecurityGroupRule is an egress rule of a security group</p>
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
<code>SecurityGroupRule</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SecurityGroupRule">
SecurityGroupRule
</a>
</em>
</td>
<td>
<p>
(Members of <code>SecurityGroupRule</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>destination</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IPSetSpec">
IPSetSpec
</a>
</em>
</td>
<td>
<p>Destination is either the cird or a reference to another security group</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.FilterRule">FilterRule
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.GatewaySpec">GatewaySpec</a>)
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
<code>securityGroup</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.Gateway">Gateway
</h3>
<div>
<p>Gateway is the Schema for the gateways API</p>
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
<a href="#network.onmetal.de/v1alpha1.GatewaySpec">
GatewaySpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>mode</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>regions</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Regions is a list of regions where this Gateway should be available</p>
</td>
</tr>
<tr>
<td>
<code>filterRules</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.FilterRule">
[]FilterRule
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>uplink</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedKindReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedKindReference
</a>
</em>
</td>
<td>
<p>Uplink is either a ReservedIP or a Subnet</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.GatewayStatus">
GatewayStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.GatewaySpec">GatewaySpec
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.Gateway">Gateway</a>)
</p>
<div>
<p>GatewaySpec defines the desired state of Gateway</p>
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
<code>mode</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>regions</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>Regions is a list of regions where this Gateway should be available</p>
</td>
</tr>
<tr>
<td>
<code>filterRules</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.FilterRule">
[]FilterRule
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>uplink</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedKindReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedKindReference
</a>
</em>
</td>
<td>
<p>Uplink is either a ReservedIP or a Subnet</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.GatewayStatus">GatewayStatus
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.Gateway">Gateway</a>)
</p>
<div>
<p>GatewayStatus defines the observed state of Gateway</p>
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
<code>ips</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IPAddr">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPAddr
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.IPAMRange">IPAMRange
</h3>
<div>
<p>IPAMRange is the Schema for the ipamranges API</p>
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
<a href="#network.onmetal.de/v1alpha1.IPAMRangeSpec">
IPAMRangeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>parent</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>size</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>cidr</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IPAMRangeStatus">
IPAMRangeStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.IPAMRangeSpec">IPAMRangeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.IPAMRange">IPAMRange</a>)
</p>
<div>
<p>IPAMRangeSpec defines the desired state of IPAMRange
Either parent and size or a give CIDR must be specified. If parent is specified,
the effective range of the given size is allocated from the parent IP range. If parent and CIDR
is defined, the given CIDR must be in the parent range and unused. It will be allocated if possible.
Otherwise the status of the object will be set to &ldquo;Failed&rdquo;.</p>
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
<code>parent</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>size</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>cidr</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.IPAMRangeStatus">IPAMRangeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.IPAMRange">IPAMRange</a>)
</p>
<div>
<p>IPAMRangeStatus defines the observed state of IPAMRange</p>
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
<code>bound</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedKindReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedKindReference
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>cidr</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>freeBlocks</code><br/>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.IPSetSpec">IPSetSpec
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.EgressSecurityGroupRule">EgressSecurityGroupRule</a>, <a href="#network.onmetal.de/v1alpha1.IngressSecurityGroupRule">IngressSecurityGroupRule</a>)
</p>
<div>
<p>IPSetSpec defines either a cidr or a security group reference</p>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Cidr">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.Cidr
</a>
</em>
</td>
<td>
<p>CIDR block for source/destination</p>
</td>
</tr>
<tr>
<td>
<code>securityGroupref</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>SecurityGroupRef references a security group</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.IngressSecurityGroupRule">IngressSecurityGroupRule
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroupSpec">SecurityGroupSpec</a>)
</p>
<div>
<p>IngressSecurityGroupRule is an ingress rule of a security group</p>
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
<code>SecurityGroupRule</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SecurityGroupRule">
SecurityGroupRule
</a>
</em>
</td>
<td>
<p>
(Members of <code>SecurityGroupRule</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IPSetSpec">
IPSetSpec
</a>
</em>
</td>
<td>
<p>Source is either the cird or a reference to another security group</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.PortRange">PortRange
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroupRule">SecurityGroupRule</a>)
</p>
<div>
<p>PortRange defines the start and end of a port range</p>
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
<code>startPort</code><br/>
<em>
int
</em>
</td>
<td>
<p>StartPort is the start port of the port range</p>
</td>
</tr>
<tr>
<td>
<code>endPort</code><br/>
<em>
int
</em>
</td>
<td>
<p>EndPort is the end port of the port range</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.RangeType">RangeType
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SubnetSpec">SubnetSpec</a>)
</p>
<div>
<p>RangeType defines the range/size of a subnet</p>
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
<code>ipam</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>IPAM is a reference to the an range block of a subnet</p>
</td>
</tr>
<tr>
<td>
<code>size</code><br/>
<em>
string
</em>
</td>
<td>
<p>Size defines the size of a subnet e.g. &ldquo;/12&rdquo;</p>
</td>
</tr>
<tr>
<td>
<code>cidr</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Cidr">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.Cidr
</a>
</em>
</td>
<td>
<p>CIDR is the CIDR block</p>
</td>
</tr>
<tr>
<td>
<code>blockedRanges</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>BlockedRanges specifies which part of the subnet should be used for static IP assignment
e.g. 0/14 means the first /14 subnet is blocked in the allocated /12 subnet</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.ReservedIP">ReservedIP
</h3>
<div>
<p>ReservedIP is the Schema for the reservedips API</p>
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
<a href="#network.onmetal.de/v1alpha1.ReservedIPSpec">
ReservedIPSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>subnet</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Subnet references the subnet where an IP address should be reserved</p>
</td>
</tr>
<tr>
<td>
<code>ip</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IPAddr">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPAddr
</a>
</em>
</td>
<td>
<p>IP specifies an IP address which should be reserved. Must be in the CIDR of the
associated Subnet</p>
</td>
</tr>
<tr>
<td>
<code>assignment</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedKindReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedKindReference
</a>
</em>
</td>
<td>
<p>Assignment indicates to which resource this IP address should be assigned</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.ReservedIPStatus">
ReservedIPStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.ReservedIPBound">ReservedIPBound
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.ReservedIPStatus">ReservedIPStatus</a>)
</p>
<div>
<p>ReservedIPBound describes the binding state of a ReservedIP</p>
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
<code>mode</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>assignment</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedKindReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedKindReference
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.ReservedIPSpec">ReservedIPSpec
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.ReservedIP">ReservedIP</a>)
</p>
<div>
<p>ReservedIPSpec defines the desired state of ReservedIP</p>
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
<code>subnet</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Subnet references the subnet where an IP address should be reserved</p>
</td>
</tr>
<tr>
<td>
<code>ip</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IPAddr">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPAddr
</a>
</em>
</td>
<td>
<p>IP specifies an IP address which should be reserved. Must be in the CIDR of the
associated Subnet</p>
</td>
</tr>
<tr>
<td>
<code>assignment</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedKindReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedKindReference
</a>
</em>
</td>
<td>
<p>Assignment indicates to which resource this IP address should be assigned</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.ReservedIPStatus">ReservedIPStatus
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.ReservedIP">ReservedIP</a>)
</p>
<div>
<p>ReservedIPStatus defines the observed state of ReservedIP</p>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IPAddr">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IPAddr
</a>
</em>
</td>
<td>
<p>IP indicates the effective reserved IP address</p>
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
<tr>
<td>
<code>bound</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.ReservedIPBound">
ReservedIPBound
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.Route">Route
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.RoutingDomainSpec">RoutingDomainSpec</a>)
</p>
<div>
<p>Route describes a single route definition</p>
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
<code>subnetRef</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>SubnetRef is a reference to Subnet</p>
</td>
</tr>
<tr>
<td>
<code>cidr</code><br/>
<em>
string
</em>
</td>
<td>
<p>CIDR is the matching CIDR of a Route</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedKindReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedKindReference
</a>
</em>
</td>
<td>
<p>Target is the target object of a Route</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.RoutingDomain">RoutingDomain
</h3>
<div>
<p>RoutingDomain is the Schema for the RoutingDomain API</p>
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
<a href="#network.onmetal.de/v1alpha1.RoutingDomainSpec">
RoutingDomainSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>routes</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.Route">
[]Route
</a>
</em>
</td>
<td>
<p>Routes is a list of routing instructions</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.RoutingDomainStatus">
RoutingDomainStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.RoutingDomainSpec">RoutingDomainSpec
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.RoutingDomain">RoutingDomain</a>)
</p>
<div>
<p>RoutingDomainSpec defines the desired state of RoutingDomain
Subnets associated with a RoutingDomain are routed implicitly and don&rsquo;t
need explicit routing instructions.</p>
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
<code>routes</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.Route">
[]Route
</a>
</em>
</td>
<td>
<p>Routes is a list of routing instructions</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.RoutingDomainStatus">RoutingDomainStatus
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.RoutingDomain">RoutingDomain</a>)
</p>
<div>
<p>RoutingDomainStatus defines the observed state of RoutingDomain</p>
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
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SecurityGroup">SecurityGroup
</h3>
<div>
<p>SecurityGroup is the Schema for the securitygroups API</p>
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
<a href="#network.onmetal.de/v1alpha1.SecurityGroupSpec">
SecurityGroupSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>ingress</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IngressSecurityGroupRule">
[]IngressSecurityGroupRule
</a>
</em>
</td>
<td>
<p>Ingress is a list of inbound rules</p>
</td>
</tr>
<tr>
<td>
<code>egress</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.EgressSecurityGroupRule">
[]EgressSecurityGroupRule
</a>
</em>
</td>
<td>
<p>Egress is a list of outbound rules</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SecurityGroupStatus">
SecurityGroupStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SecurityGroupRule">SecurityGroupRule
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.EgressSecurityGroupRule">EgressSecurityGroupRule</a>, <a href="#network.onmetal.de/v1alpha1.IngressSecurityGroupRule">IngressSecurityGroupRule</a>)
</p>
<div>
<p>SecurityGroupRule is a single access rule</p>
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
<p>Name is the name of the SecurityGroupRule</p>
</td>
</tr>
<tr>
<td>
<code>securityGroupRef</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>SecurityGroupRef is a scoped reference to an existing SecurityGroup</p>
</td>
</tr>
<tr>
<td>
<code>action</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.ActionType">
ActionType
</a>
</em>
</td>
<td>
<p>Action defines the action type of a SecurityGroupRule</p>
</td>
</tr>
<tr>
<td>
<code>protocol</code><br/>
<em>
string
</em>
</td>
<td>
<p>Protocol defines the protocol of a SecurityGroupRule</p>
</td>
</tr>
<tr>
<td>
<code>portRange</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.PortRange">
PortRange
</a>
</em>
</td>
<td>
<p>PortRange is the port range of the SecurityGroupRule</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SecurityGroupSpec">SecurityGroupSpec
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroup">SecurityGroup</a>)
</p>
<div>
<p>SecurityGroupSpec defines the desired state of SecurityGroup</p>
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
<code>ingress</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IngressSecurityGroupRule">
[]IngressSecurityGroupRule
</a>
</em>
</td>
<td>
<p>Ingress is a list of inbound rules</p>
</td>
</tr>
<tr>
<td>
<code>egress</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.EgressSecurityGroupRule">
[]EgressSecurityGroupRule
</a>
</em>
</td>
<td>
<p>Egress is a list of outbound rules</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SecurityGroupStatus">SecurityGroupStatus
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroup">SecurityGroup</a>)
</p>
<div>
<p>SecurityGroupStatus defines the observed state of SecurityGroup</p>
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
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.Subnet">Subnet
</h3>
<div>
<p>Subnet is the Schema for the subnets API</p>
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
<a href="#network.onmetal.de/v1alpha1.SubnetSpec">
SubnetSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>parent</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Parent is a reference to a public parent Subnet without regional manifestation. The direct children
then represent the regional incarnations of this public subnet.</p>
</td>
</tr>
<tr>
<td>
<code>locations</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.RegionAvailability">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.RegionAvailability
</a>
</em>
</td>
<td>
<p>Locations defines in which regions and availability zone this subnet should be available</p>
</td>
</tr>
<tr>
<td>
<code>routingDomain</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>RoutingDomain is the reference to the routing domain this SubNet should be associated with</p>
</td>
</tr>
<tr>
<td>
<code>ranges</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.RangeType">
[]RangeType
</a>
</em>
</td>
<td>
<p>Ranges defines the size of the subnet</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SubnetStatus">
SubnetStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SubnetSpec">SubnetSpec
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.Subnet">Subnet</a>)
</p>
<div>
<p>SubnetSpec defines the desired state of Subnet</p>
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
<code>parent</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Parent is a reference to a public parent Subnet without regional manifestation. The direct children
then represent the regional incarnations of this public subnet.</p>
</td>
</tr>
<tr>
<td>
<code>locations</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.RegionAvailability">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.RegionAvailability
</a>
</em>
</td>
<td>
<p>Locations defines in which regions and availability zone this subnet should be available</p>
</td>
</tr>
<tr>
<td>
<code>routingDomain</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>RoutingDomain is the reference to the routing domain this SubNet should be associated with</p>
</td>
</tr>
<tr>
<td>
<code>ranges</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.RangeType">
[]RangeType
</a>
</em>
</td>
<td>
<p>Ranges defines the size of the subnet</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SubnetStatus">SubnetStatus
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.Subnet">Subnet</a>)
</p>
<div>
<p>SubnetStatus defines the observed state of Subnet</p>
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
<code>cidrs</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.CIDRStatus">
[]CIDRStatus
</a>
</em>
</td>
<td>
<p>CIDRs is a list of CIDR status</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>aa5aba9</code>.
</em></p>
