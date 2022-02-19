<p>Packages:</p>
<ul>
<li>
<a href="#storage.onmetal.de%2fv1alpha1">storage.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="storage.onmetal.de/v1alpha1">storage.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="storage.onmetal.de/v1alpha1.ClaimReference">ClaimReference
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeSpec">VolumeSpec</a>)
</p>
<div>
<p>ClaimReference points to a referenced VolumeClaim.</p>
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
<p>Name is the name of the referenced VolumeClaim.</p>
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
<p>UID is the UID of the referenced VolumeClaim.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.MVIVolumeSource">MVIVolumeSource
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeSpec">VolumeSpec</a>)
</p>
<div>
<p>MVIVolumeSource represents storage that is managed by an external MVI volume driver.</p>
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
<code>secretRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>SecretRef references the Secret containing the access credentials to consume a Volume.</p>
</td>
</tr>
<tr>
<td>
<code>driver</code><br/>
<em>
string
</em>
</td>
<td>
<p>Driver is the name of the drive to use for this volume. Required.</p>
</td>
</tr>
<tr>
<td>
<code>volumeAttributes</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>VolumeAttributes are attributes of the volume to use.</p>
</td>
</tr>
</tbody>
</table>
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
<p>Taints of the StoragePool. Only Volumes who tolerate all the taints
will land in the StoragePool.</p>
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
<p>Taints of the StoragePool. Only Volumes who tolerate all the taints
will land in the StoragePool.</p>
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
<code>availableStorageClasses</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>AvailableStorageClasses list the references of supported StorageClasses of this pool</p>
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
<code>storageClassRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>StorageClassRef is the storage class of a volume</p>
</td>
</tr>
<tr>
<td>
<code>storagePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>StoragePoolSelector selects a suitable StoragePool by the given labels.</p>
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
<p>StoragePool indicates which storage pool to use for a volume.
If unset, the scheduler will figure out a suitable StoragePool.</p>
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
<p>ProviderID is the unique id of the volume on provider side.</p>
</td>
</tr>
<tr>
<td>
<code>claimRef</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.ClaimReference">
ClaimReference
</a>
</em>
</td>
<td>
<p>ClaimRef is the reference to the VolumeClaim used by the Volume.</p>
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
<p>Tolerations define tolerations the Volume has. Only StoragePools whose taints
covered by Tolerations will be considered to host the Volume.</p>
</td>
</tr>
<tr>
<td>
<code>mvi</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.MVIVolumeSource">
MVIVolumeSource
</a>
</em>
</td>
<td>
<p>MVI specifies how to manage the volume via MVI.</p>
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
<h3 id="storage.onmetal.de/v1alpha1.VolumeClaim">VolumeClaim
</h3>
<div>
<p>VolumeClaim is the Schema for the volumeclaims API</p>
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
<a href="#storage.onmetal.de/v1alpha1.VolumeClaimSpec">
VolumeClaimSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>volumeRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumeRef is the reference to the Volume used by the VolumeClaim</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector is a label query over volumes to consider for binding.</p>
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
<p>Resources are the requested Volume resources.</p>
</td>
</tr>
<tr>
<td>
<code>storageClassRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>StorageClassRef references the StorageClass used by the Volume.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.VolumeClaimStatus">
VolumeClaimStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeClaimPhase">VolumeClaimPhase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeClaimStatus">VolumeClaimStatus</a>)
</p>
<div>
<p>VolumeClaimPhase represents the state a VolumeClaim can be in.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Bound&#34;</p></td>
<td><p>VolumeClaimBound is used for a VolumeClaim which is bound to a Volume.</p>
</td>
</tr><tr><td><p>&#34;Lost&#34;</p></td>
<td><p>VolumeClaimLost is used for a VolumeClaim that lost its underlying Volume. The claim was bound to a
Volume and this volume does not exist any longer and all data on it was lost.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>VolumeClaimPending is used for a VolumeClaim which is not yet bound.</p>
</td>
</tr></tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeClaimSpec">VolumeClaimSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeClaim">VolumeClaim</a>)
</p>
<div>
<p>VolumeClaimSpec defines the desired state of VolumeClaim</p>
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
<code>volumeRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumeRef is the reference to the Volume used by the VolumeClaim</p>
</td>
</tr>
<tr>
<td>
<code>selector</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#labelselector-v1-meta">
Kubernetes meta/v1.LabelSelector
</a>
</em>
</td>
<td>
<p>Selector is a label query over volumes to consider for binding.</p>
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
<p>Resources are the requested Volume resources.</p>
</td>
</tr>
<tr>
<td>
<code>storageClassRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>StorageClassRef references the StorageClass used by the Volume.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeClaimStatus">VolumeClaimStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeClaim">VolumeClaim</a>)
</p>
<div>
<p>VolumeClaimStatus defines the observed state of VolumeClaim</p>
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
<code>phase</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.VolumeClaimPhase">
VolumeClaimPhase
</a>
</em>
</td>
<td>
<p>Phase represents the state a VolumeClaim can be in.</p>
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
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Synced&#34;</p></td>
<td><p>VolumeSynced represents the condition of a volume being synced with its backing resources</p>
</td>
</tr></tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumePhase">VolumePhase
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus</a>)
</p>
<div>
<p>VolumePhase represents the VolumeClaim binding phase of a Volume</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Available&#34;</p></td>
<td><p>VolumeAvailable is used for Volumes that are not yet bound
Available volumes are held by the binder and matched to VolumeClaims.</p>
</td>
</tr><tr><td><p>&#34;Bound&#34;</p></td>
<td><p>VolumeBound is used for Volumes that are bound.</p>
</td>
</tr><tr><td><p>&#34;Failed&#34;</p></td>
<td><p>VolumeFailed is used for Volumes that failed to be correctly freed from a VolumeClaim.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>VolumePending is used for Volumes that are not available.</p>
</td>
</tr></tbody>
</table>
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
<code>storageClassRef</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>StorageClassRef is the storage class of a volume</p>
</td>
</tr>
<tr>
<td>
<code>storagePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>StoragePoolSelector selects a suitable StoragePool by the given labels.</p>
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
<p>StoragePool indicates which storage pool to use for a volume.
If unset, the scheduler will figure out a suitable StoragePool.</p>
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
<p>ProviderID is the unique id of the volume on provider side.</p>
</td>
</tr>
<tr>
<td>
<code>claimRef</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.ClaimReference">
ClaimReference
</a>
</em>
</td>
<td>
<p>ClaimRef is the reference to the VolumeClaim used by the Volume.</p>
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
<p>Tolerations define tolerations the Volume has. Only StoragePools whose taints
covered by Tolerations will be considered to host the Volume.</p>
</td>
</tr>
<tr>
<td>
<code>mvi</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.MVIVolumeSource">
MVIVolumeSource
</a>
</em>
</td>
<td>
<p>MVI specifies how to manage the volume via MVI.</p>
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
<tbody><tr><td><p>&#34;Available&#34;</p></td>
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
<p>State represents the infrastructure state of a Volume.</p>
</td>
</tr>
<tr>
<td>
<code>phase</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.VolumePhase">
VolumePhase
</a>
</em>
</td>
<td>
<p>Phase represents the VolumeClaim binding phase of a Volume.</p>
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
on git commit <code>7399651</code>.
</em></p>
