<p>Packages:</p>
<ul>
<li>
<a href="#compute.api.onmetal.de%2fv1alpha1">compute.api.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="compute.api.onmetal.de/v1alpha1">compute.api.onmetal.de/v1alpha1</h2>
<div>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</div>
Resource Types:
<ul><li>
<a href="#compute.api.onmetal.de/v1alpha1.Machine">Machine</a>
</li><li>
<a href="#compute.api.onmetal.de/v1alpha1.MachineClass">MachineClass</a>
</li><li>
<a href="#compute.api.onmetal.de/v1alpha1.MachinePool">MachinePool</a>
</li></ul>
<h3 id="compute.api.onmetal.de/v1alpha1.Machine">Machine
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
<code>apiVersion</code><br/>
string</td>
<td>
<code>
compute.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>Machine</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta">
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
<a href="#compute.api.onmetal.de/v1alpha1.MachineSpec">
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
<code>machineClassRef</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachineClassRef is a reference to the machine class/flavor of the machine.</p>
</td>
</tr>
<tr>
<td>
<code>machinePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>MachinePoolSelector selects a suitable MachinePoolRef by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>machinePoolRef</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachinePoolRef defines machine pool to run the machine in.
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
<code>interfaces</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.Interface">
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
<code>volumes</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.Volume">
[]Volume
</a>
</em>
</td>
<td>
<p>Volumes are volumes attached to this machine.</p>
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
<p>IgnitionRef is a reference to a config map containing the ignition YAML for the machine to boot up.
If key is empty, DefaultIgnitionKey will be used as fallback.</p>
</td>
</tr>
<tr>
<td>
<code>efiVars</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.EFIVar">
[]EFIVar
</a>
</em>
</td>
<td>
<p>EFIVars are variables to pass to EFI while booting up.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Toleration">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations define tolerations the Machine has. Only MachinePools whose taints
covered by Tolerations will be considered to run the Machine.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.MachineStatus">
MachineStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.MachineClass">MachineClass
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
<code>apiVersion</code><br/>
string</td>
<td>
<code>
compute.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>MachineClass</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta">
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
<code>capabilities</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.MachinePool">MachinePool
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
<code>apiVersion</code><br/>
string</td>
<td>
<code>
compute.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>MachinePool</code></td>
</tr>
<tr>
<td>
<code>metadata</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#objectmeta-v1-meta">
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
<a href="#compute.api.onmetal.de/v1alpha1.MachinePoolSpec">
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
<tr>
<td>
<code>taints</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Taint">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.Taint
</a>
</em>
</td>
<td>
<p>Taints of the MachinePool. Only Machines who tolerate all the taints
will land in the MachinePool.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.MachinePoolStatus">
MachinePoolStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.EFIVar">EFIVar
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachineSpec">MachineSpec</a>)
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
<h3 id="compute.api.onmetal.de/v1alpha1.EphemeralNetworkInterfaceSource">EphemeralNetworkInterfaceSource
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.InterfaceSource">InterfaceSource</a>)
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
<code>networkInterfaceTemplate</code><br/>
<em>
github.com/onmetal/onmetal-api/apis/networking/v1alpha1.NetworkInterfaceTemplateSpec
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.Interface">Interface
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachineSpec">MachineSpec</a>)
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
<p>Name is the name of the network interface.</p>
</td>
</tr>
<tr>
<td>
<code>InterfaceSource</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.InterfaceSource">
InterfaceSource
</a>
</em>
</td>
<td>
<p>
(Members of <code>InterfaceSource</code> are embedded into this type.)
</p>
<p>InterfaceSource is where to obtain the interface from.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.InterfaceSource">InterfaceSource
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.Interface">Interface</a>)
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
<code>networkInterfaceRef</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>NetworkInterfaceRef instructs to use the NetworkInterface at the target reference.</p>
</td>
</tr>
<tr>
<td>
<code>ephemeral</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.EphemeralNetworkInterfaceSource">
EphemeralNetworkInterfaceSource
</a>
</em>
</td>
<td>
<p>Ephemeral instructs to create an ephemeral (i.e. coupled to the lifetime of the surrounding object)
NetworkInterface to use.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.InterfaceStatus">InterfaceStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
</p>
<div>
<p>InterfaceStatus reports the status of an InterfaceSource.</p>
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
<code>virtualIP</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.MachineCondition">MachineCondition
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
</p>
<div>
<p>MachineCondition is one of the conditions of a Machine.</p>
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
<a href="#compute.api.onmetal.de/v1alpha1.MachineConditionType">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#conditionstatus-v1-core">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#time-v1-meta">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#time-v1-meta">
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
<h3 id="compute.api.onmetal.de/v1alpha1.MachineConditionType">MachineConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachineCondition">MachineCondition</a>)
</p>
<div>
<p>MachineConditionType is a type a MachineCondition can have.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Synced&#34;</p></td>
<td><p>MachineSynced represents the condition of a machine being synced with its backing resources</p>
</td>
</tr></tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.MachinePoolCondition">MachinePoolCondition
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachinePoolStatus">MachinePoolStatus</a>)
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
<a href="#compute.api.onmetal.de/v1alpha1.MachinePoolConditionType">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#conditionstatus-v1-core">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#time-v1-meta">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#time-v1-meta">
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
<h3 id="compute.api.onmetal.de/v1alpha1.MachinePoolConditionType">MachinePoolConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachinePoolCondition">MachinePoolCondition</a>)
</p>
<div>
<p>MachinePoolConditionType is a type a MachinePoolCondition can have.</p>
</div>
<h3 id="compute.api.onmetal.de/v1alpha1.MachinePoolSpec">MachinePoolSpec
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachinePool">MachinePool</a>)
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
<tr>
<td>
<code>taints</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Taint">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.Taint
</a>
</em>
</td>
<td>
<p>Taints of the MachinePool. Only Machines who tolerate all the taints
will land in the MachinePool.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.MachinePoolState">MachinePoolState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachinePoolStatus">MachinePoolStatus</a>)
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
<h3 id="compute.api.onmetal.de/v1alpha1.MachinePoolStatus">MachinePoolStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachinePool">MachinePool</a>)
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
<a href="#compute.api.onmetal.de/v1alpha1.MachinePoolState">
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
<a href="#compute.api.onmetal.de/v1alpha1.MachinePoolCondition">
[]MachinePoolCondition
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>availableMachineClasses</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.MachineSpec">MachineSpec
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.Machine">Machine</a>)
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
<code>machineClassRef</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachineClassRef is a reference to the machine class/flavor of the machine.</p>
</td>
</tr>
<tr>
<td>
<code>machinePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>MachinePoolSelector selects a suitable MachinePoolRef by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>machinePoolRef</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachinePoolRef defines machine pool to run the machine in.
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
<code>interfaces</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.Interface">
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
<code>volumes</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.Volume">
[]Volume
</a>
</em>
</td>
<td>
<p>Volumes are volumes attached to this machine.</p>
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
<p>IgnitionRef is a reference to a config map containing the ignition YAML for the machine to boot up.
If key is empty, DefaultIgnitionKey will be used as fallback.</p>
</td>
</tr>
<tr>
<td>
<code>efiVars</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.EFIVar">
[]EFIVar
</a>
</em>
</td>
<td>
<p>EFIVars are variables to pass to EFI while booting up.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Toleration">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations define tolerations the Machine has. Only MachinePools whose taints
covered by Tolerations will be considered to run the Machine.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.MachineState">MachineState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
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
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>MachineStatePending means the Machine has been accepted by the system, but not yet completely started.
This includes time before being bound to a MachinePool, as well as time spent setting up the Machine on that
MachinePool.</p>
</td>
</tr><tr><td><p>&#34;Running&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Shutdown&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.MachineStatus">MachineStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.Machine">Machine</a>)
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
<a href="#compute.api.onmetal.de/v1alpha1.MachineState">
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
<a href="#compute.api.onmetal.de/v1alpha1.MachineCondition">
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
<a href="#compute.api.onmetal.de/v1alpha1.InterfaceStatus">
[]InterfaceStatus
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>volumes</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.VolumeStatus">
[]VolumeStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.RetainPolicy">RetainPolicy
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
<tbody><tr><td><p>&#34;DeleteOnTermination&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Persistent&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.Volume">Volume
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachineSpec">MachineSpec</a>)
</p>
<div>
<p>Volume defines a volume attachment of a machine</p>
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
<p>Name is the name of the Volume</p>
</td>
</tr>
<tr>
<td>
<code>VolumeSource</code><br/>
<em>
<a href="#compute.api.onmetal.de/v1alpha1.VolumeSource">
VolumeSource
</a>
</em>
</td>
<td>
<p>
(Members of <code>VolumeSource</code> are embedded into this type.)
</p>
<p>VolumeSource is the source where the storage for the Volume resides at.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.VolumeSource">VolumeSource
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.Volume">Volume</a>)
</p>
<div>
<p>VolumeSource specifies the source to use for a Volume.</p>
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
<code>volumeClaimRef</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumeClaim instructs the Volume to use a VolumeClaim as source for the attachment.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.api.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.api.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
</p>
<div>
<p>VolumeStatus is the status of a Volume.</p>
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
<p>Name is the name of a volume attachment.</p>
</td>
</tr>
<tr>
<td>
<code>deviceID</code><br/>
<em>
string
</em>
</td>
<td>
<p>DeviceID is the disk device ID on the host.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
</em></p>
