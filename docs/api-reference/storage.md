<p>Packages:</p>
<ul>
<li>
<a href="#storage.onmetal.de%2fv1alpha1">storage.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="storage.onmetal.de/v1alpha1">storage.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="storage.onmetal.de/v1alpha1.StorageClass">StorageClass
</h3>
<div>
<p>StorageClass is the Schema for the storageclasses API</p>
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
<a href="#storage.onmetal.de/v1alpha1.StorageClassSpec">
StorageClassSpec
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
<p>Capabilities describes the capabilities of a storage class</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.StorageClassStatus">
StorageClassStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.StorageClassSpec">StorageClassSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StorageClass">StorageClass</a>)
</p>
<div>
<p>StorageClassSpec defines the desired state of StorageClass</p>
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
<p>Capabilities describes the capabilities of a storage class</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.StorageClassStatus">StorageClassStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StorageClass">StorageClass</a>)
</p>
<div>
<p>StorageClassStatus defines the observed state of StorageClass</p>
</div>
<h3 id="storage.onmetal.de/v1alpha1.StoragePool">StoragePool
</h3>
<div>
<p>StoragePool is the Schema for the storagepools API</p>
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
<a href="#storage.onmetal.de/v1alpha1.StoragePoolSpec">
StoragePoolSpec
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
<p>ProviderID identifies the StoragePool on provider side.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.StoragePoolStatus">
StoragePoolStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.StoragePoolCondition">StoragePoolCondition
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StoragePoolStatus">StoragePoolStatus</a>)
</p>
<div>
<p>StoragePoolCondition is one of the conditions of a volume.</p>
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
<a href="#storage.onmetal.de/v1alpha1.StoragePoolConditionType">
StoragePoolConditionType
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
<h3 id="storage.onmetal.de/v1alpha1.StoragePoolConditionType">StoragePoolConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StoragePoolCondition">StoragePoolCondition</a>)
</p>
<div>
<p>StoragePoolConditionType is a type a StoragePoolCondition can have.</p>
</div>
<h3 id="storage.onmetal.de/v1alpha1.StoragePoolSpec">StoragePoolSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StoragePool">StoragePool</a>)
</p>
<div>
<p>StoragePoolSpec defines the desired state of StoragePool</p>
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
<p>ProviderID identifies the StoragePool on provider side.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.StoragePoolState">StoragePoolState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StoragePoolStatus">StoragePoolStatus</a>)
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
</tr><tr><td><p>&#34;NotAvailable&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.StoragePoolStatus">StoragePoolStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StoragePool">StoragePool</a>)
</p>
<div>
<p>StoragePoolStatus defines the observed state of StoragePool</p>
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
<a href="#storage.onmetal.de/v1alpha1.StoragePoolState">
StoragePoolState
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
<a href="#storage.onmetal.de/v1alpha1.StoragePoolCondition">
[]StoragePoolCondition
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>available</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<p>Available list the available capacity of a storage pool</p>
</td>
</tr>
<tr>
<td>
<code>used</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<p>Used indicates how much capacity has been used in a storage pool</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.Volume">Volume
</h3>
<div>
<p>Volume is the Schema for the volumes API</p>
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
<a href="#storage.onmetal.de/v1alpha1.VolumeSpec">
VolumeSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
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
<p>StorageClass is the storage class of a volume</p>
</td>
</tr>
<tr>
<td>
<code>storagePool</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>StoragePool indicates which storage pool to use for a volume</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<p>Resources is a description of the volume&rsquo;s resources and capacity.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.VolumeStatus">
VolumeStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeAttachment">VolumeAttachment
</h3>
<div>
<p>VolumeAttachment is the Schema for the volumeattachments API</p>
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
<a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentSpec">
VolumeAttachmentSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
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
<p>Volume is a reference of the volume object which should be attached</p>
</td>
</tr>
<tr>
<td>
<code>machine</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Machine is a reference of the machine object which the volume should be attached to</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentStatus">
VolumeAttachmentStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeAttachmentCondition">VolumeAttachmentCondition
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentStatus">VolumeAttachmentStatus</a>)
</p>
<div>
<p>VolumeAttachmentCondition is one of the conditions of a volume.</p>
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
<a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentConditionType">
VolumeAttachmentConditionType
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
<h3 id="storage.onmetal.de/v1alpha1.VolumeAttachmentConditionType">VolumeAttachmentConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentCondition">VolumeAttachmentCondition</a>)
</p>
<div>
<p>VolumeAttachmentConditionType is a type a VolumeAttachmentCondition can have.</p>
</div>
<h3 id="storage.onmetal.de/v1alpha1.VolumeAttachmentSpec">VolumeAttachmentSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeAttachment">VolumeAttachment</a>)
</p>
<div>
<p>VolumeAttachmentSpec defines the desired state of VolumeAttachment</p>
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
<code>volume</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Volume is a reference of the volume object which should be attached</p>
</td>
</tr>
<tr>
<td>
<code>machine</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>Machine is a reference of the machine object which the volume should be attached to</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeAttachmentState">VolumeAttachmentState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentStatus">VolumeAttachmentStatus</a>)
</p>
<div>
<p>VolumeAttachmentState is a state a VolumeAttachment can be int.</p>
</div>
<h3 id="storage.onmetal.de/v1alpha1.VolumeAttachmentStatus">VolumeAttachmentStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeAttachment">VolumeAttachment</a>)
</p>
<div>
<p>VolumeAttachmentStatus defines the observed state of VolumeAttachment</p>
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
<a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentState">
VolumeAttachmentState
</a>
</em>
</td>
<td>
<p>State reports a VolumeAttachmentState a VolumeAttachment is in.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentCondition">
[]VolumeAttachmentCondition
</a>
</em>
</td>
<td>
<p>Conditions reports the conditions a VolumeAttachment may have.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeCondition">VolumeCondition
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus</a>)
</p>
<div>
<p>VolumeCondition is one of the conditions of a volume.</p>
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
<a href="#storage.onmetal.de/v1alpha1.VolumeConditionType">
VolumeConditionType
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
<h3 id="storage.onmetal.de/v1alpha1.VolumeConditionType">VolumeConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeCondition">VolumeCondition</a>)
</p>
<div>
<p>VolumeConditionType is a type a VolumeCondition can have.</p>
</div>
<h3 id="storage.onmetal.de/v1alpha1.VolumeSpec">VolumeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.Volume">Volume</a>)
</p>
<div>
<p>VolumeSpec defines the desired state of Volume</p>
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
<code>storageClass</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>StorageClass is the storage class of a volume</p>
</td>
</tr>
<tr>
<td>
<code>storagePool</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>StoragePool indicates which storage pool to use for a volume</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#resourcelist-v1-core">
Kubernetes core/v1.ResourceList
</a>
</em>
</td>
<td>
<p>Resources is a description of the volume&rsquo;s resources and capacity.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeState">VolumeState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus</a>)
</p>
<div>
<p>VolumeState is a possible state a volume can be in.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Attached&#34;</p></td>
<td><p>VolumeStateAttached reports that the volume is attached and in-use.</p>
</td>
</tr><tr><td><p>&#34;Available&#34;</p></td>
<td><p>VolumeStateAvailable reports whether the volume is available to be used.</p>
</td>
</tr><tr><td><p>&#34;Error&#34;</p></td>
<td><p>VolumeStateError reports that the volume is in an error state.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>VolumeStatePending reports whether the volume is about to be ready.</p>
</td>
</tr></tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.Volume">Volume</a>)
</p>
<div>
<p>VolumeStatus defines the observed state of Volume</p>
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
<a href="#storage.onmetal.de/v1alpha1.VolumeState">
VolumeState
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
<a href="#storage.onmetal.de/v1alpha1.VolumeCondition">
[]VolumeCondition
</a>
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
