<p>Packages:</p>
<ul>
<li>
<a href="#compute.onmetal.de%2fv1alpha1">compute.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="compute.onmetal.de/v1alpha1">compute.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="compute.onmetal.de/v1alpha1.EFIVar">EFIVar
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineSpec">MachineSpec</a>)
</p>
<div>
<p>EFIVar is a variable to pass to EFI while booting up.</p>
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
<code>uuid</code><br/>
<em>
string
</em>
</td>
<td>
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
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.Interface">Interface
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineSpec">MachineSpec</a>)
</p>
<div>
<p>Interface is the definition of a single interface</p>
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
<p>Name is the name of the interface</p>
</td>
</tr>
<tr>
<td>
<code>target</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Target is the referenced resource of this interface</p>
</td>
</tr>
<tr>
<td>
<code>priority</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Priority is the priority level of this interface</p>
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
<p>IP specifies a concrete IP address which should be allocated from a Subnet</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.InterfaceStatus">InterfaceStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
</p>
<div>
<p>InterfaceStatus reports the status of an Interface.</p>
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
<p>Name is the name of an interface.</p>
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
<p>IP is the IP allocated for an interface.</p>
</td>
</tr>
<tr>
<td>
<code>priority</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Priority is the OS priority of the interface.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.Machine">Machine
</h3>
<div>
<p>Machine is the Schema for the machines API</p>
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
<a href="#compute.onmetal.de/v1alpha1.MachineSpec">
MachineSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>hostname</code><br/>
<em>
string
</em>
</td>
<td>
<p>Hostname is the hostname of the machine</p>
</td>
</tr>
<tr>
<td>
<code>machineClass</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachineClass is a reference to the machine class/flavor of the machine.</p>
</td>
</tr>
<tr>
<td>
<code>machinePool</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachinePool defines machine pool to run the machine in.
If empty, a scheduler will figure out an appropriate pool to run the machine in.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<p>Image is the URL providing the operating system image of the machine.</p>
</td>
</tr>
<tr>
<td>
<code>sshPublicKeys</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.SecretKeySelector">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>SSHPublicKeys is a list of SSH public key secret references of a machine.</p>
</td>
</tr>
<tr>
<td>
<code>interfaces</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.Interface">
[]Interface
</a>
</em>
</td>
<td>
<p>Interfaces define a list of network interfaces present on the machine</p>
</td>
</tr>
<tr>
<td>
<code>securityGroups</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>SecurityGroups is a list of security groups of a machine</p>
</td>
</tr>
<tr>
<td>
<code>volumeClaims</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.VolumeClaim">
[]VolumeClaim
</a>
</em>
</td>
<td>
<p>VolumeClaims</p>
</td>
</tr>
<tr>
<td>
<code>ignition</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ConfigMapKeySelector">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ConfigMapKeySelector
</a>
</em>
</td>
<td>
<p>Ignition is a reference to a config map containing the ignition YAML for the machine to boot up.
If key is empty, DefaultIgnitionKey will be used as fallback.</p>
</td>
</tr>
<tr>
<td>
<code>efiVars</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.EFIVar">
[]EFIVar
</a>
</em>
</td>
<td>
<p>EFIVars are variables to pass to EFI while booting up.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.MachineStatus">
MachineStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachineClass">MachineClass
</h3>
<div>
<p>MachineClass is the Schema for the machineclasses API</p>
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
<a href="#compute.onmetal.de/v1alpha1.MachineClassSpec">
MachineClassSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>capabilities</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<p>Capabilities describes the resources a machine class can provide.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.MachineClassStatus">
MachineClassStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachineClassSpec">MachineClassSpec
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineClass">MachineClass</a>)
</p>
<div>
<p>MachineClassSpec defines the desired state of MachineClass</p>
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
<code>capabilities</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<p>Capabilities describes the resources a machine class can provide.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachineClassStatus">MachineClassStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineClass">MachineClass</a>)
</p>
<div>
<p>MachineClassStatus defines the observed state of MachineClass</p>
</div>
<h3 id="compute.onmetal.de/v1alpha1.MachineCondition">MachineCondition
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
</p>
<div>
<p>MachineCondition is one of the conditions of a volume.</p>
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
<a href="#compute.onmetal.de/v1alpha1.MachineConditionType">
MachineConditionType
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
<h3 id="compute.onmetal.de/v1alpha1.MachineConditionType">MachineConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineCondition">MachineCondition</a>)
</p>
<div>
<p>MachineConditionType is a type a MachineCondition can have.</p>
</div>
<h3 id="compute.onmetal.de/v1alpha1.MachinePool">MachinePool
</h3>
<div>
<p>MachinePool is the Schema for the machinepools API</p>
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
<a href="#compute.onmetal.de/v1alpha1.MachinePoolSpec">
MachinePoolSpec
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
<p>ProviderID identifies the MachinePool on provider side.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.MachinePoolStatus">
MachinePoolStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachinePoolCondition">MachinePoolCondition
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachinePoolStatus">MachinePoolStatus</a>)
</p>
<div>
<p>MachinePoolCondition is one of the conditions of a volume.</p>
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
<a href="#compute.onmetal.de/v1alpha1.MachinePoolConditionType">
MachinePoolConditionType
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
<h3 id="compute.onmetal.de/v1alpha1.MachinePoolConditionType">MachinePoolConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachinePoolCondition">MachinePoolCondition</a>)
</p>
<div>
<p>MachinePoolConditionType is a type a MachinePoolCondition can have.</p>
</div>
<h3 id="compute.onmetal.de/v1alpha1.MachinePoolSpec">MachinePoolSpec
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachinePool">MachinePool</a>)
</p>
<div>
<p>MachinePoolSpec defines the desired state of MachinePool</p>
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
<p>ProviderID identifies the MachinePool on provider side.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachinePoolState">MachinePoolState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachinePoolStatus">MachinePoolStatus</a>)
</p>
<div>
<p>MachinePoolState is a state a MachinePool can be in.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Error&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Offline&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Ready&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachinePoolStatus">MachinePoolStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachinePool">MachinePool</a>)
</p>
<div>
<p>MachinePoolStatus defines the observed state of MachinePool</p>
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
<a href="#compute.onmetal.de/v1alpha1.MachinePoolState">
MachinePoolState
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
<a href="#compute.onmetal.de/v1alpha1.MachinePoolCondition">
[]MachinePoolCondition
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachineSpec">MachineSpec
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.Machine">Machine</a>)
</p>
<div>
<p>MachineSpec defines the desired state of Machine</p>
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
<code>hostname</code><br/>
<em>
string
</em>
</td>
<td>
<p>Hostname is the hostname of the machine</p>
</td>
</tr>
<tr>
<td>
<code>machineClass</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachineClass is a reference to the machine class/flavor of the machine.</p>
</td>
</tr>
<tr>
<td>
<code>machinePool</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachinePool defines machine pool to run the machine in.
If empty, a scheduler will figure out an appropriate pool to run the machine in.</p>
</td>
</tr>
<tr>
<td>
<code>image</code><br/>
<em>
string
</em>
</td>
<td>
<p>Image is the URL providing the operating system image of the machine.</p>
</td>
</tr>
<tr>
<td>
<code>sshPublicKeys</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.SecretKeySelector">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.SecretKeySelector
</a>
</em>
</td>
<td>
<p>SSHPublicKeys is a list of SSH public key secret references of a machine.</p>
</td>
</tr>
<tr>
<td>
<code>interfaces</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.Interface">
[]Interface
</a>
</em>
</td>
<td>
<p>Interfaces define a list of network interfaces present on the machine</p>
</td>
</tr>
<tr>
<td>
<code>securityGroups</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>SecurityGroups is a list of security groups of a machine</p>
</td>
</tr>
<tr>
<td>
<code>volumeClaims</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.VolumeClaim">
[]VolumeClaim
</a>
</em>
</td>
<td>
<p>VolumeClaims</p>
</td>
</tr>
<tr>
<td>
<code>ignition</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ConfigMapKeySelector">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ConfigMapKeySelector
</a>
</em>
</td>
<td>
<p>Ignition is a reference to a config map containing the ignition YAML for the machine to boot up.
If key is empty, DefaultIgnitionKey will be used as fallback.</p>
</td>
</tr>
<tr>
<td>
<code>efiVars</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.EFIVar">
[]EFIVar
</a>
</em>
</td>
<td>
<p>EFIVars are variables to pass to EFI while booting up.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachineState">MachineState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
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
<tbody><tr><td><p>&#34;Error&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Initial&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Running&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Shutdown&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.MachineStatus">MachineStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.Machine">Machine</a>)
</p>
<div>
<p>MachineStatus defines the observed state of Machine</p>
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
<a href="#compute.onmetal.de/v1alpha1.MachineState">
MachineState
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
<a href="#compute.onmetal.de/v1alpha1.MachineCondition">
[]MachineCondition
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>interfaces</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.InterfaceStatus">
[]InterfaceStatus
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>volumeClaims</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.VolumeClaimStatus">
[]VolumeClaimStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.RetainPolicy">RetainPolicy
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.VolumeClaim">VolumeClaim</a>)
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
<tbody><tr><td><p>&#34;DeleteOnTermination&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Persistent&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.VolumeClaim">VolumeClaim
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineSpec">MachineSpec</a>)
</p>
<div>
<p>VolumeClaim defines a volume claim of a machine</p>
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
<p>Name is the name of the VolumeClaim</p>
</td>
</tr>
<tr>
<td>
<code>priority</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Priority is the OS priority of the volume.</p>
</td>
</tr>
<tr>
<td>
<code>retainPolicy</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.RetainPolicy">
RetainPolicy
</a>
</em>
</td>
<td>
<p>RetainPolicy defines what should happen when the machine is being deleted</p>
</td>
</tr>
<tr>
<td>
<code>storageClass</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>StorageClass describes the storage class of the volumes</p>
</td>
</tr>
<tr>
<td>
<code>size</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Duration">
k8s.io/apimachinery/pkg/api/resource.Quantity
</a>
</em>
</td>
<td>
<p>Size defines the size of the volume</p>
</td>
</tr>
<tr>
<td>
<code>volume</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Volume is a reference to an existing volume</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.VolumeClaimStatus">VolumeClaimStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
</p>
<div>
<p>VolumeClaimStatus is the status of a VolumeClaim.</p>
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
<p>Name is the name of a volume claim.</p>
</td>
</tr>
<tr>
<td>
<code>priority</code><br/>
<em>
int32
</em>
</td>
<td>
<p>Priority is the OS priority of the volume.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>6fe95dc</code>.
</em></p>
