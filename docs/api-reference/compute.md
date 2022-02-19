<p>Packages:</p>
<ul>
<li>
<a href="#compute.onmetal.de%2fv1alpha1">compute.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="compute.onmetal.de/v1alpha1">compute.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="compute.onmetal.de/v1alpha1.Console">Console
</h3>
<div>
<p>Console is the Schema for the consoles API</p>
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
<a href="#compute.onmetal.de/v1alpha1.ConsoleSpec">
ConsoleSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>machineRef</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachineRef references the machine to open a console to.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.ConsoleStatus">
ConsoleStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.ConsoleClientConfig">ConsoleClientConfig
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.ConsoleStatus">ConsoleStatus</a>)
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
<code>service</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.ServiceReference">
ServiceReference
</a>
</em>
</td>
<td>
<p>Service is the service to connect to.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.ConsoleSpec">ConsoleSpec
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.Console">Console</a>)
</p>
<div>
<p>ConsoleSpec defines the desired state of Console</p>
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
<code>machineRef</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>MachineRef references the machine to open a console to.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.ConsoleState">ConsoleState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.ConsoleStatus">ConsoleStatus</a>)
</p>
<div>
<p>ConsoleState is a state a Console can be in.</p>
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
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Ready&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.ConsoleStatus">ConsoleStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.Console">Console</a>)
</p>
<div>
<p>ConsoleStatus defines the observed state of Console</p>
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
<a href="#compute.onmetal.de/v1alpha1.ConsoleState">
ConsoleState
</a>
</em>
</td>
<td>
<p>State is the state of a Console.</p>
</td>
</tr>
<tr>
<td>
<code>clientConfig</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.ConsoleClientConfig">
ConsoleClientConfig
</a>
</em>
</td>
<td>
<p>ClientConfig is the client configuration to connect to a console.
Only usable if the ConsoleStatus.State is ConsoleStateReady.</p>
</td>
</tr>
</tbody>
</table>
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Target is the referenced resource of this interface.</p>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.IP">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.IP
</a>
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
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
<code>machinePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>MachinePoolSelector selects a suitable MachinePool by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>machinePool</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
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
<code>volumeAttachments</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.VolumeAttachment">
[]VolumeAttachment
</a>
</em>
</td>
<td>
<p>VolumeAttachments are volumes attached to this machine.</p>
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcelist-v1-core">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#resourcelist-v1-core">
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
<h3 id="compute.onmetal.de/v1alpha1.MachineConditionType">MachineConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineCondition">MachineCondition</a>)
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
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
<code>machinePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>MachinePoolSelector selects a suitable MachinePool by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>machinePool</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
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
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
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
<code>volumeAttachments</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.VolumeAttachment">
[]VolumeAttachment
</a>
</em>
</td>
<td>
<p>VolumeAttachments are volumes attached to this machine.</p>
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
<code>volumeAttachments</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.VolumeAttachmentStatus">
[]VolumeAttachmentStatus
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
<h3 id="compute.onmetal.de/v1alpha1.ServiceReference">ServiceReference
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.ConsoleClientConfig">ConsoleClientConfig</a>)
</p>
<div>
<p>ServiceReference is a reference to a Service in the same namespace as the referent.</p>
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
<p>Name of the referenced service.</p>
</td>
</tr>
<tr>
<td>
<code>path</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p><code>path</code> is an optional URL path which will be sent in any request to
this service.</p>
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
<p>Port on the service hosting the console.
Defaults to 443 for backward compatibility.
<code>port</code> should be a valid port number (1-65535, inclusive).</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.VolumeAttachment">VolumeAttachment
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineSpec">MachineSpec</a>)
</p>
<div>
<p>VolumeAttachment defines a volume attachment of a machine</p>
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
<p>Name is the name of the VolumeAttachment</p>
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
<code>VolumeAttachmentSource</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.VolumeAttachmentSource">
VolumeAttachmentSource
</a>
</em>
</td>
<td>
<p>
(Members of <code>VolumeAttachmentSource</code> are embedded into this type.)
</p>
<p>VolumeAttachmentSource is the source where the storage for the VolumeAttachment resides at.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.VolumeAttachmentSource">VolumeAttachmentSource
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.VolumeAttachment">VolumeAttachment</a>)
</p>
<div>
<p>VolumeAttachmentSource specifies the source to use for a VolumeAttachment.</p>
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
<code>volumeClaim</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.VolumeClaimAttachmentSource">
VolumeClaimAttachmentSource
</a>
</em>
</td>
<td>
<p>VolumeClaim instructs the VolumeAttachment to use a VolumeClaim as source for the attachment.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.VolumeAttachmentStatus">VolumeAttachmentStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineStatus">MachineStatus</a>)
</p>
<div>
<p>VolumeAttachmentStatus is the status of a VolumeAttachment.</p>
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
<h3 id="compute.onmetal.de/v1alpha1.VolumeClaimAttachmentSource">VolumeClaimAttachmentSource
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.VolumeAttachmentSource">VolumeAttachmentSource</a>)
</p>
<div>
<p>VolumeClaimAttachmentSource references a VolumeClaim as VolumeAttachment source.</p>
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
<code>ref</code><br/>
<em>
<a href="https://v1-23.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Ref is a reference to the VolumeClaim.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>7399651</code>.
</em></p>
