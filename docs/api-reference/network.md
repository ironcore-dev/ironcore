<p>Packages:</p>
<ul>
<li>
<a href="#network.onmetal.de%2fv1alpha1">network.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="network.onmetal.de/v1alpha1">network.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="network.onmetal.de/v1alpha1.AllocationState">AllocationState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.CIDRAllocationStatus">CIDRAllocationStatus</a>)
</p>
<div>
<p>AllocationState is a state an allocation can be in.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Allocated&#34;</p></td>
<td><p>AllocationStateAllocated reports that the allocation has been made successfully.</p>
</td>
</tr><tr><td><p>&#34;Busy&#34;</p></td>
<td><p>AllocationStateBusy reports that an allocation is busy.</p>
</td>
</tr><tr><td><p>&#34;Failed&#34;</p></td>
<td><p>AllocationStateFailed reports that an allocation has failed.</p>
</td>
</tr></tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.CIDRAllocation">CIDRAllocation
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.CIDRAllocationStatus">CIDRAllocationStatus</a>, <a href="#network.onmetal.de/v1alpha1.IPAMPendingRequest">IPAMPendingRequest</a>)
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
<code>request</code><br/>
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
<h3 id="network.onmetal.de/v1alpha1.CIDRAllocationStatus">CIDRAllocationStatus
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.IPAMRangeStatus">IPAMRangeStatus</a>)
</p>
<div>
<p>CIDRAllocationStatus is the result of a CIDR allocation request</p>
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
<code>CIDRAllocation</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.CIDRAllocation">
CIDRAllocation
</a>
</em>
</td>
<td>
<p>
(Members of <code>CIDRAllocation</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.AllocationState">
AllocationState
</a>
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
</tbody>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.CIDR">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.CIDR
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.CIDR">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.CIDR
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
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
<a href="#network.onmetal.de/v1alpha1.GatewayMode">
GatewayMode
</a>
</em>
</td>
<td>
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
<a href="#network.onmetal.de/v1alpha1.Target">
Target
</a>
</em>
</td>
<td>
<p>Uplink is a Target to route traffic to.</p>
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
<h3 id="network.onmetal.de/v1alpha1.GatewayCondition">GatewayCondition
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.GatewayStatus">GatewayStatus</a>)
</p>
<div>
<p>GatewayCondition is one of the conditions of a volume.</p>
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
<a href="#network.onmetal.de/v1alpha1.GatewayConditionType">
GatewayConditionType
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#conditionstatus-v1-core">
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
<code>lastUpdateTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastUpdateTime is the last time a condition has been updated.</p>
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
<p>LastTransitionTime is the last time the status of a condition has transitioned from one state to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.GatewayConditionType">GatewayConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.GatewayCondition">GatewayCondition</a>)
</p>
<div>
<p>GatewayConditionType is a type a GatewayCondition can have.</p>
</div>
<h3 id="network.onmetal.de/v1alpha1.GatewayMode">GatewayMode
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.GatewaySpec">GatewaySpec</a>)
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
<tbody><tr><td><p>&#34;NAT&#34;</p></td>
<td><p>NATMode is regular NAT (network address translation).</p>
</td>
</tr><tr><td><p>&#34;SNAT&#34;</p></td>
<td><p>SNATMode is stateless NAT / 1-1 NAT (network address translation).</p>
</td>
</tr><tr><td><p>&#34;Transparent&#34;</p></td>
<td><p>TransparentMode makes the gateway behave transparently.</p>
</td>
</tr></tbody>
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
<a href="#network.onmetal.de/v1alpha1.GatewayMode">
GatewayMode
</a>
</em>
</td>
<td>
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
<a href="#network.onmetal.de/v1alpha1.Target">
Target
</a>
</em>
</td>
<td>
<p>Uplink is a Target to route traffic to.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.GatewayState">GatewayState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.GatewayStatus">GatewayStatus</a>)
</p>
<div>
</div>
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
<code>state</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.GatewayState">
GatewayState
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.GatewayCondition">
[]GatewayCondition
</a>
</em>
</td>
<td>
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
<h3 id="network.onmetal.de/v1alpha1.IPAMPendingRequest">IPAMPendingRequest
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.IPAMRangeStatus">IPAMRangeStatus</a>)
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
<code>name</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>namespace</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>cidrs</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.CIDRAllocation">
[]CIDRAllocation
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>deletions</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.CIDRAllocation">
[]CIDRAllocation
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Parent is the reference of the Parent IPAMRange from which the Cidr or size should be derived</p>
</td>
</tr>
<tr>
<td>
<code>cidrs</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>CIDRs is a list of CIDR specs which are defined for this IPAMRange</p>
</td>
</tr>
<tr>
<td>
<code>mode</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IPAMRangeMode">
IPAMRangeMode
</a>
</em>
</td>
<td>
<p>Mode is the mode to request an IPAMRange.</p>
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
<h3 id="network.onmetal.de/v1alpha1.IPAMRangeCondition">IPAMRangeCondition
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.IPAMRangeStatus">IPAMRangeStatus</a>)
</p>
<div>
<p>IPAMRangeCondition is one of the conditions of a volume.</p>
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
<a href="#network.onmetal.de/v1alpha1.IPAMRangeConditionType">
IPAMRangeConditionType
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#conditionstatus-v1-core">
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
<code>lastUpdateTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastUpdateTime is the last time a condition has been updated.</p>
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
<p>LastTransitionTime is the last time the status of a condition has transitioned from one state to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.IPAMRangeConditionType">IPAMRangeConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.IPAMRangeCondition">IPAMRangeCondition</a>)
</p>
<div>
<p>IPAMRangeConditionType is a type a IPAMRangeCondition can have.</p>
</div>
<h3 id="network.onmetal.de/v1alpha1.IPAMRangeMode">IPAMRangeMode
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.IPAMRangeSpec">IPAMRangeSpec</a>)
</p>
<div>
<p>IPAMRangeMode is the mode to request IPAMRanges.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;FirstMatch&#34;</p></td>
<td><p>ModeFirstMatch requests IPAMRanges by using the first possible match.</p>
</td>
</tr><tr><td><p>&#34;RoundRobin&#34;</p></td>
<td><p>ModeRoundRobin requests IPAMRanges in a round-robin fashion, distributing evenly.</p>
</td>
</tr></tbody>
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
Otherwise, the status of the object will be set to &ldquo;Invalid&rdquo;.</p>
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Parent is the reference of the Parent IPAMRange from which the Cidr or size should be derived</p>
</td>
</tr>
<tr>
<td>
<code>cidrs</code><br/>
<em>
[]string
</em>
</td>
<td>
<p>CIDRs is a list of CIDR specs which are defined for this IPAMRange</p>
</td>
</tr>
<tr>
<td>
<code>mode</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IPAMRangeMode">
IPAMRangeMode
</a>
</em>
</td>
<td>
<p>Mode is the mode to request an IPAMRange.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.IPAMRangeState">IPAMRangeState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.IPAMRangeStatus">IPAMRangeStatus</a>)
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
<tbody><tr><td><p>&#34;Available&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Busy&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Error&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Invalid&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Ready&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Up&#34;</p></td>
<td></td>
</tr></tbody>
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
<code>state</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IPAMRangeState">
IPAMRangeState
</a>
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
<code>conditions</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IPAMRangeCondition">
[]IPAMRangeCondition
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>cidrs</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.CIDRAllocationStatus">
[]CIDRAllocationStatus
</a>
</em>
</td>
<td>
<p>CIDRs is a list of effective cidrs which belong to this IPAMRange</p>
</td>
</tr>
<tr>
<td>
<code>allocationState</code><br/>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>roundRobinState</code><br/>
<em>
[]string
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>pendingRequests</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.IPAMPendingRequest">
IPAMPendingRequest
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>pendingDeletions</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.CIDRAllocationStatus">
[]CIDRAllocationStatus
</a>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.CIDR">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.CIDR
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
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
<h3 id="network.onmetal.de/v1alpha1.MachineRouteTarget">MachineRouteTarget
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.Target">Target</a>)
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
<code>LocalObjectReference</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>
(Members of <code>LocalObjectReference</code> are embedded into this type.)
</p>
</td>
</tr>
<tr>
<td>
<code>interface</code><br/>
<em>
string
</em>
</td>
<td>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.CIDR">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.CIDR
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
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
<a href="#network.onmetal.de/v1alpha1.ReservedIPAssignment">
ReservedIPAssignment
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
<h3 id="network.onmetal.de/v1alpha1.ReservedIPAssignment">ReservedIPAssignment
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.ReservedIPBound">ReservedIPBound</a>, <a href="#network.onmetal.de/v1alpha1.ReservedIPSpec">ReservedIPSpec</a>)
</p>
<div>
<p>ReservedIPAssignment contains information that points to the resource being used.</p>
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
<code>apiGroup</code><br/>
<em>
string
</em>
</td>
<td>
<p>APIGroup is the group for the resource being referenced</p>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
<em>
string
</em>
</td>
<td>
<p>Kind is the type of resource being referenced</p>
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
<p>Name is the name of resource being referenced</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.ReservedIPBindMode">ReservedIPBindMode
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.ReservedIPBound">ReservedIPBound</a>)
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
<tbody><tr><td><p>&#34;Floating&#34;</p></td>
<td><p>BindModeFloating defines a ReservedIP which is dynamically assigned
as additional DNAT-ed IP for the target resource.</p>
</td>
</tr><tr><td><p>&#34;Static&#34;</p></td>
<td><p>BindModeStatic defines a ReservedIP which is directly assigned to an interface
of the target resource. This means the target is directly connected to the Subnet
of the reserved IP.</p>
</td>
</tr></tbody>
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
<a href="#network.onmetal.de/v1alpha1.ReservedIPBindMode">
ReservedIPBindMode
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>assignment</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.ReservedIPAssignment">
ReservedIPAssignment
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.ReservedIPCondition">ReservedIPCondition
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.ReservedIPStatus">ReservedIPStatus</a>)
</p>
<div>
<p>ReservedIPCondition is one of the conditions of a volume.</p>
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
<a href="#network.onmetal.de/v1alpha1.ReservedIPConditionType">
ReservedIPConditionType
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#conditionstatus-v1-core">
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
<code>lastUpdateTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastUpdateTime is the last time a condition has been updated.</p>
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
<p>LastTransitionTime is the last time the status of a condition has transitioned from one state to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.ReservedIPConditionType">ReservedIPConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.ReservedIPCondition">ReservedIPCondition</a>)
</p>
<div>
<p>ReservedIPConditionType is a type a ReservedIPCondition can have.</p>
</div>
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
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
<a href="#network.onmetal.de/v1alpha1.ReservedIPAssignment">
ReservedIPAssignment
</a>
</em>
</td>
<td>
<p>Assignment indicates to which resource this IP address should be assigned</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.ReservedIPState">ReservedIPState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.ReservedIPStatus">ReservedIPStatus</a>)
</p>
<div>
</div>
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
<code>state</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.ReservedIPState">
ReservedIPState
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.ReservedIPCondition">
[]ReservedIPCondition
</a>
</em>
</td>
<td>
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
<p>IP indicates the effective reserved IP address</p>
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
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
<a href="#network.onmetal.de/v1alpha1.Target">
Target
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
<h3 id="network.onmetal.de/v1alpha1.RoutingDomainCondition">RoutingDomainCondition
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.RoutingDomainStatus">RoutingDomainStatus</a>)
</p>
<div>
<p>RoutingDomainCondition is one of the conditions of a volume.</p>
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
<a href="#network.onmetal.de/v1alpha1.RoutingDomainConditionType">
RoutingDomainConditionType
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#conditionstatus-v1-core">
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
<code>lastUpdateTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastUpdateTime is the last time a condition has been updated.</p>
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
<p>LastTransitionTime is the last time the status of a condition has transitioned from one state to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.RoutingDomainConditionType">RoutingDomainConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.RoutingDomainCondition">RoutingDomainCondition</a>)
</p>
<div>
<p>RoutingDomainConditionType is a type a RoutingDomainCondition can have.</p>
</div>
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
<h3 id="network.onmetal.de/v1alpha1.RoutingDomainState">RoutingDomainState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.RoutingDomainStatus">RoutingDomainStatus</a>)
</p>
<div>
</div>
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
<code>state</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.RoutingDomainState">
RoutingDomainState
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.RoutingDomainCondition">
[]RoutingDomainCondition
</a>
</em>
</td>
<td>
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
<h3 id="network.onmetal.de/v1alpha1.SecurityGroupAction">SecurityGroupAction
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroupRule">SecurityGroupRule</a>)
</p>
<div>
<p>SecurityGroupAction describes the action of a SecurityGroupRule.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Allow&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Deny&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SecurityGroupCondition">SecurityGroupCondition
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroupStatus">SecurityGroupStatus</a>)
</p>
<div>
<p>SecurityGroupCondition is one of the conditions of a volume.</p>
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
<a href="#network.onmetal.de/v1alpha1.SecurityGroupConditionType">
SecurityGroupConditionType
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#conditionstatus-v1-core">
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
<code>lastUpdateTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastUpdateTime is the last time a condition has been updated.</p>
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
<p>LastTransitionTime is the last time the status of a condition has transitioned from one state to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SecurityGroupConditionType">SecurityGroupConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroupCondition">SecurityGroupCondition</a>)
</p>
<div>
<p>SecurityGroupConditionType is a type a SecurityGroupCondition can have.</p>
</div>
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>SecurityGroupRef is a reference to an existing SecurityGroup</p>
</td>
</tr>
<tr>
<td>
<code>action</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SecurityGroupAction">
SecurityGroupAction
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
<h3 id="network.onmetal.de/v1alpha1.SecurityGroupState">SecurityGroupState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SecurityGroupStatus">SecurityGroupStatus</a>)
</p>
<div>
<p>SecurityGroupState is the state of a SecurityGroup.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Invalid&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unused&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Used&#34;</p></td>
<td></td>
</tr></tbody>
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
<code>state</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SecurityGroupState">
SecurityGroupState
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SecurityGroupCondition">
[]SecurityGroupCondition
</a>
</em>
</td>
<td>
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Parent is a reference to a public parent Subnet.</p>
</td>
</tr>
<tr>
<td>
<code>machinePools</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachinePools defines in which pools this subnet should be available</p>
</td>
</tr>
<tr>
<td>
<code>routingDomain</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
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
<h3 id="network.onmetal.de/v1alpha1.SubnetCondition">SubnetCondition
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SubnetStatus">SubnetStatus</a>)
</p>
<div>
<p>SubnetCondition is one of the conditions of a volume.</p>
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
<a href="#network.onmetal.de/v1alpha1.SubnetConditionType">
SubnetConditionType
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#conditionstatus-v1-core">
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
<code>lastUpdateTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastUpdateTime is the last time a condition has been updated.</p>
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
<p>LastTransitionTime is the last time the status of a condition has transitioned from one state to another.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="network.onmetal.de/v1alpha1.SubnetConditionType">SubnetConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SubnetCondition">SubnetCondition</a>)
</p>
<div>
<p>SubnetConditionType is a type a SubnetCondition can have.</p>
</div>
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
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Parent is a reference to a public parent Subnet.</p>
</td>
</tr>
<tr>
<td>
<code>machinePools</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachinePools defines in which pools this subnet should be available</p>
</td>
</tr>
<tr>
<td>
<code>routingDomain</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
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
<h3 id="network.onmetal.de/v1alpha1.SubnetState">SubnetState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.SubnetStatus">SubnetStatus</a>)
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
<tbody><tr><td><p>&#34;Down&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Initial&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Invalid&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Up&#34;</p></td>
<td></td>
</tr></tbody>
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
<code>state</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SubnetState">
SubnetState
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.SubnetCondition">
[]SubnetCondition
</a>
</em>
</td>
<td>
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
<h3 id="network.onmetal.de/v1alpha1.Target">Target
</h3>
<p>
(<em>Appears on:</em><a href="#network.onmetal.de/v1alpha1.GatewaySpec">GatewaySpec</a>, <a href="#network.onmetal.de/v1alpha1.Route">Route</a>)
</p>
<div>
<p>Target is a target for network traffic.
It may be either
* a v1alpha1.Machine
* a Gateway
* a ReservedIP
* a raw IP
* a raw CIDR.</p>
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
<code>machine</code><br/>
<em>
<a href="#network.onmetal.de/v1alpha1.MachineRouteTarget">
MachineRouteTarget
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>gateway</code><br/>
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
<code>reservedIP</code><br/>
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
<code>ip</code><br/>
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
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>6fe95dc</code>.
</em></p>
