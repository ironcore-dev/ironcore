<p>Packages:</p>
<ul>
<li>
<a href="#storage.api.onmetal.de%2fv1alpha1">storage.api.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="storage.api.onmetal.de/v1alpha1">storage.api.onmetal.de/v1alpha1</h2>
<div>
<p>Package v1alpha1 is the v1alpha1 version of the API.</p>
</div>
Resource Types:
<ul><li>
<a href="#storage.api.onmetal.de/v1alpha1.Bucket">Bucket</a>
</li><li>
<a href="#storage.api.onmetal.de/v1alpha1.BucketClass">BucketClass</a>
</li><li>
<a href="#storage.api.onmetal.de/v1alpha1.BucketPool">BucketPool</a>
</li><li>
<a href="#storage.api.onmetal.de/v1alpha1.Volume">Volume</a>
</li><li>
<a href="#storage.api.onmetal.de/v1alpha1.VolumeClass">VolumeClass</a>
</li><li>
<a href="#storage.api.onmetal.de/v1alpha1.VolumePool">VolumePool</a>
</li></ul>
<h3 id="storage.api.onmetal.de/v1alpha1.Bucket">Bucket
</h3>
<div>
<p>Bucket is the Schema for the buckets API</p>
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
storage.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>Bucket</code></td>
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
<a href="#storage.api.onmetal.de/v1alpha1.BucketSpec">
BucketSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>bucketClassRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>BucketClassRef is the BucketClass of a bucket
If empty, an external controller has to provision the bucket.</p>
</td>
</tr>
<tr>
<td>
<code>bucketPoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>BucketPoolSelector selects a suitable BucketPoolRef by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>bucketPoolRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>BucketPoolRef indicates which BucketPool to use for a bucket.
If unset, the scheduler will figure out a suitable BucketPoolRef.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Toleration">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations define tolerations the Bucket has. Only any BucketPool whose taints
covered by Tolerations will be considered to host the Bucket.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.BucketStatus">
BucketStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketClass">BucketClass
</h3>
<div>
<p>BucketClass is the Schema for the bucketclasses API</p>
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
storage.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>BucketClass</code></td>
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
<code>capabilities</code><br/>
<em>
<a href="../core/#core.api.onmetal.de/v1alpha1.ResourceList">
github.com/onmetal/onmetal-api/api/core/v1alpha1.ResourceList
</a>
</em>
</td>
<td>
<p>Capabilities describes the capabilities of a BucketClass.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketPool">BucketPool
</h3>
<div>
<p>BucketPool is the Schema for the bucketpools API</p>
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
storage.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>BucketPool</code></td>
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
<a href="#storage.api.onmetal.de/v1alpha1.BucketPoolSpec">
BucketPoolSpec
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
<p>ProviderID identifies the BucketPool on provider side.</p>
</td>
</tr>
<tr>
<td>
<code>taints</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Taint">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Taint
</a>
</em>
</td>
<td>
<p>Taints of the BucketPool. Only Buckets who tolerate all the taints
will land in the BucketPool.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.BucketPoolStatus">
BucketPoolStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.Volume">Volume
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
<code>apiVersion</code><br/>
string</td>
<td>
<code>
storage.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>Volume</code></td>
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
<a href="#storage.api.onmetal.de/v1alpha1.VolumeSpec">
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
<code>volumeClassRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumeClassRef is the VolumeClass of a volume
If empty, an external controller has to provision the volume.</p>
</td>
</tr>
<tr>
<td>
<code>volumePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>VolumePoolSelector selects a suitable VolumePoolRef by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>volumePoolRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumePoolRef indicates which VolumePool to use for a volume.
If unset, the scheduler will figure out a suitable VolumePoolRef.</p>
</td>
</tr>
<tr>
<td>
<code>claimRef</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.LocalUIDReference">
github.com/onmetal/onmetal-api/api/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>ClaimRef is the reference to the claiming entity of the Volume.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="../core/#core.api.onmetal.de/v1alpha1.ResourceList">
github.com/onmetal/onmetal-api/api/core/v1alpha1.ResourceList
</a>
</em>
</td>
<td>
<p>Resources is a description of the volume&rsquo;s resources and capacity.</p>
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
<p>Image is an optional image to bootstrap the volume with.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecretRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>ImagePullSecretRef is an optional secret for pulling the image of a volume.</p>
</td>
</tr>
<tr>
<td>
<code>unclaimable</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Unclaimable marks the volume as unclaimable.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Toleration">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations define tolerations the Volume has. Only any VolumePool whose taints
covered by Tolerations will be considered to host the Volume.</p>
</td>
</tr>
<tr>
<td>
<code>encryption</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.VolumeEncryption">
VolumeEncryption
</a>
</em>
</td>
<td>
<p>Encryption is an optional field which provides attributes to encrypt Volume.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.VolumeStatus">
VolumeStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeClass">VolumeClass
</h3>
<div>
<p>VolumeClass is the Schema for the volumeclasses API</p>
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
storage.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>VolumeClass</code></td>
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
<code>capabilities</code><br/>
<em>
<a href="../core/#core.api.onmetal.de/v1alpha1.ResourceList">
github.com/onmetal/onmetal-api/api/core/v1alpha1.ResourceList
</a>
</em>
</td>
<td>
<p>Capabilities describes the capabilities of a VolumeClass.</p>
</td>
</tr>
<tr>
<td>
<code>resizePolicy</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.ResizePolicy">
ResizePolicy
</a>
</em>
</td>
<td>
<p>ResizePolicy describes the supported expansion policy of a VolumeClass.
If not set default to Static expansion policy.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumePool">VolumePool
</h3>
<div>
<p>VolumePool is the Schema for the volumepools API</p>
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
storage.api.onmetal.de/v1alpha1
</code>
</td>
</tr>
<tr>
<td>
<code>kind</code><br/>
string
</td>
<td><code>VolumePool</code></td>
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
<a href="#storage.api.onmetal.de/v1alpha1.VolumePoolSpec">
VolumePoolSpec
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
<p>ProviderID identifies the VolumePool on provider side.</p>
</td>
</tr>
<tr>
<td>
<code>taints</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Taint">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Taint
</a>
</em>
</td>
<td>
<p>Taints of the VolumePool. Only Volumes who tolerate all the taints
will land in the VolumePool.</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.VolumePoolStatus">
VolumePoolStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketAccess">BucketAccess
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.BucketStatus">BucketStatus</a>)
</p>
<div>
<p>BucketAccess represents information on how to access a bucket.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>SecretRef references the Secret containing the access credentials to consume a Bucket.</p>
</td>
</tr>
<tr>
<td>
<code>endpoint</code><br/>
<em>
string
</em>
</td>
<td>
<p>Endpoint defines address of the Bucket REST-API.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketCondition">BucketCondition
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.BucketStatus">BucketStatus</a>)
</p>
<div>
<p>BucketCondition is one of the conditions of a bucket.</p>
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
<a href="#storage.api.onmetal.de/v1alpha1.BucketConditionType">
BucketConditionType
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
<h3 id="storage.api.onmetal.de/v1alpha1.BucketConditionType">BucketConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.BucketCondition">BucketCondition</a>)
</p>
<div>
<p>BucketConditionType is a type a BucketCondition can have.</p>
</div>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketPoolSpec">BucketPoolSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.BucketPool">BucketPool</a>)
</p>
<div>
<p>BucketPoolSpec defines the desired state of BucketPool</p>
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
<p>ProviderID identifies the BucketPool on provider side.</p>
</td>
</tr>
<tr>
<td>
<code>taints</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Taint">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Taint
</a>
</em>
</td>
<td>
<p>Taints of the BucketPool. Only Buckets who tolerate all the taints
will land in the BucketPool.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketPoolState">BucketPoolState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.BucketPoolStatus">BucketPoolStatus</a>)
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
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unavailable&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketPoolStatus">BucketPoolStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.BucketPool">BucketPool</a>)
</p>
<div>
<p>BucketPoolStatus defines the observed state of BucketPool</p>
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
<a href="#storage.api.onmetal.de/v1alpha1.BucketPoolState">
BucketPoolState
</a>
</em>
</td>
<td>
<p>State represents the infrastructure state of a BucketPool.</p>
</td>
</tr>
<tr>
<td>
<code>availableBucketClasses</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>AvailableBucketClasses list the references of any supported BucketClass of this pool</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketSpec">BucketSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.Bucket">Bucket</a>, <a href="#storage.api.onmetal.de/v1alpha1.BucketTemplateSpec">BucketTemplateSpec</a>)
</p>
<div>
<p>BucketSpec defines the desired state of Bucket</p>
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
<code>bucketClassRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>BucketClassRef is the BucketClass of a bucket
If empty, an external controller has to provision the bucket.</p>
</td>
</tr>
<tr>
<td>
<code>bucketPoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>BucketPoolSelector selects a suitable BucketPoolRef by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>bucketPoolRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>BucketPoolRef indicates which BucketPool to use for a bucket.
If unset, the scheduler will figure out a suitable BucketPoolRef.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Toleration">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations define tolerations the Bucket has. Only any BucketPool whose taints
covered by Tolerations will be considered to host the Bucket.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketState">BucketState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.BucketStatus">BucketStatus</a>)
</p>
<div>
<p>BucketState represents the infrastructure state of a Bucket.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Available&#34;</p></td>
<td><p>BucketStateAvailable reports whether a Bucket is available to be used.</p>
</td>
</tr><tr><td><p>&#34;Error&#34;</p></td>
<td><p>BucketStateError reports that a Bucket is in an error state.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>BucketStatePending reports whether a Bucket is about to be ready.</p>
</td>
</tr></tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketStatus">BucketStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.Bucket">Bucket</a>)
</p>
<div>
<p>BucketStatus defines the observed state of Bucket</p>
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
<a href="#storage.api.onmetal.de/v1alpha1.BucketState">
BucketState
</a>
</em>
</td>
<td>
<p>State represents the infrastructure state of a Bucket.</p>
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
<p>LastStateTransitionTime is the last time the State transitioned between values.</p>
</td>
</tr>
<tr>
<td>
<code>access</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.BucketAccess">
BucketAccess
</a>
</em>
</td>
<td>
<p>Access specifies how to access a Bucket.
This is set by the bucket provider when the bucket is provisioned.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.BucketCondition">
[]BucketCondition
</a>
</em>
</td>
<td>
<p>Conditions are the conditions of a bucket.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.BucketTemplateSpec">BucketTemplateSpec
</h3>
<div>
<p>BucketTemplateSpec is the specification of a Bucket template.</p>
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
<a href="#storage.api.onmetal.de/v1alpha1.BucketSpec">
BucketSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>bucketClassRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>BucketClassRef is the BucketClass of a bucket
If empty, an external controller has to provision the bucket.</p>
</td>
</tr>
<tr>
<td>
<code>bucketPoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>BucketPoolSelector selects a suitable BucketPoolRef by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>bucketPoolRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>BucketPoolRef indicates which BucketPool to use for a bucket.
If unset, the scheduler will figure out a suitable BucketPoolRef.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Toleration">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations define tolerations the Bucket has. Only any BucketPool whose taints
covered by Tolerations will be considered to host the Bucket.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.ResizePolicy">ResizePolicy
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumeClass">VolumeClass</a>)
</p>
<div>
<p>ResizePolicy is a type of policy.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;ExpandOnly&#34;</p></td>
<td><p>ResizePolicyExpandOnly is a policy that only allows the expansion of a Volume.</p>
</td>
</tr><tr><td><p>&#34;Static&#34;</p></td>
<td><p>ResizePolicyStatic is a policy that does not allow the expansion of a Volume.</p>
</td>
</tr></tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeAccess">VolumeAccess
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus</a>)
</p>
<div>
<p>VolumeAccess represents information on how to access a volume.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
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
<code>handle</code><br/>
<em>
string
</em>
</td>
<td>
<p>Handle is the unique handle of the volume.</p>
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
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeCondition">VolumeCondition
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus</a>)
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
<a href="#storage.api.onmetal.de/v1alpha1.VolumeConditionType">
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
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeConditionType">VolumeConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumeCondition">VolumeCondition</a>)
</p>
<div>
<p>VolumeConditionType is a type a VolumeCondition can have.</p>
</div>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeEncryption">VolumeEncryption
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumeSpec">VolumeSpec</a>)
</p>
<div>
<p>VolumeEncryption represents information to encrypt a volume.</p>
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
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>SecretRef references the Secret containing the encryption key to encrypt a Volume.
This secret is created by user with encryptionKey as Key and base64 encoded 256-bit encryption key as Value.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumePoolCondition">VolumePoolCondition
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumePoolStatus">VolumePoolStatus</a>)
</p>
<div>
<p>VolumePoolCondition is one of the conditions of a volume.</p>
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
<a href="#storage.api.onmetal.de/v1alpha1.VolumePoolConditionType">
VolumePoolConditionType
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
<h3 id="storage.api.onmetal.de/v1alpha1.VolumePoolConditionType">VolumePoolConditionType
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumePoolCondition">VolumePoolCondition</a>)
</p>
<div>
<p>VolumePoolConditionType is a type a VolumePoolCondition can have.</p>
</div>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumePoolSpec">VolumePoolSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumePool">VolumePool</a>)
</p>
<div>
<p>VolumePoolSpec defines the desired state of VolumePool</p>
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
<p>ProviderID identifies the VolumePool on provider side.</p>
</td>
</tr>
<tr>
<td>
<code>taints</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Taint">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Taint
</a>
</em>
</td>
<td>
<p>Taints of the VolumePool. Only Volumes who tolerate all the taints
will land in the VolumePool.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumePoolState">VolumePoolState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumePoolStatus">VolumePoolStatus</a>)
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
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td></td>
</tr><tr><td><p>&#34;Unavailable&#34;</p></td>
<td></td>
</tr></tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumePoolStatus">VolumePoolStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumePool">VolumePool</a>)
</p>
<div>
<p>VolumePoolStatus defines the observed state of VolumePool</p>
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
<a href="#storage.api.onmetal.de/v1alpha1.VolumePoolState">
VolumePoolState
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
<a href="#storage.api.onmetal.de/v1alpha1.VolumePoolCondition">
[]VolumePoolCondition
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>availableVolumeClasses</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
[]Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>AvailableVolumeClasses list the references of any supported VolumeClass of this pool</p>
</td>
</tr>
<tr>
<td>
<code>capacity</code><br/>
<em>
<a href="../core/#core.api.onmetal.de/v1alpha1.ResourceList">
github.com/onmetal/onmetal-api/api/core/v1alpha1.ResourceList
</a>
</em>
</td>
<td>
<p>Capacity represents the total resources of a machine pool.</p>
</td>
</tr>
<tr>
<td>
<code>allocatable</code><br/>
<em>
<a href="../core/#core.api.onmetal.de/v1alpha1.ResourceList">
github.com/onmetal/onmetal-api/api/core/v1alpha1.ResourceList
</a>
</em>
</td>
<td>
<p>Allocatable represents the resources of a machine pool that are available for scheduling.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeSpec">VolumeSpec
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.Volume">Volume</a>, <a href="#storage.api.onmetal.de/v1alpha1.VolumeTemplateSpec">VolumeTemplateSpec</a>)
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
<code>volumeClassRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumeClassRef is the VolumeClass of a volume
If empty, an external controller has to provision the volume.</p>
</td>
</tr>
<tr>
<td>
<code>volumePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>VolumePoolSelector selects a suitable VolumePoolRef by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>volumePoolRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumePoolRef indicates which VolumePool to use for a volume.
If unset, the scheduler will figure out a suitable VolumePoolRef.</p>
</td>
</tr>
<tr>
<td>
<code>claimRef</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.LocalUIDReference">
github.com/onmetal/onmetal-api/api/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>ClaimRef is the reference to the claiming entity of the Volume.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="../core/#core.api.onmetal.de/v1alpha1.ResourceList">
github.com/onmetal/onmetal-api/api/core/v1alpha1.ResourceList
</a>
</em>
</td>
<td>
<p>Resources is a description of the volume&rsquo;s resources and capacity.</p>
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
<p>Image is an optional image to bootstrap the volume with.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecretRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>ImagePullSecretRef is an optional secret for pulling the image of a volume.</p>
</td>
</tr>
<tr>
<td>
<code>unclaimable</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Unclaimable marks the volume as unclaimable.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Toleration">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations define tolerations the Volume has. Only any VolumePool whose taints
covered by Tolerations will be considered to host the Volume.</p>
</td>
</tr>
<tr>
<td>
<code>encryption</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.VolumeEncryption">
VolumeEncryption
</a>
</em>
</td>
<td>
<p>Encryption is an optional field which provides attributes to encrypt Volume.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeState">VolumeState
(<code>string</code> alias)</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus</a>)
</p>
<div>
<p>VolumeState represents the infrastructure state of a Volume.</p>
</div>
<table>
<thead>
<tr>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody><tr><td><p>&#34;Available&#34;</p></td>
<td><p>VolumeStateAvailable reports whether a Volume is available to be used.</p>
</td>
</tr><tr><td><p>&#34;Error&#34;</p></td>
<td><p>VolumeStateError reports that a Volume is in an error state.</p>
</td>
</tr><tr><td><p>&#34;Pending&#34;</p></td>
<td><p>VolumeStatePending reports whether a Volume is about to be ready.</p>
</td>
</tr></tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeStatus">VolumeStatus
</h3>
<p>
(<em>Appears on:</em><a href="#storage.api.onmetal.de/v1alpha1.Volume">Volume</a>)
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
<a href="#storage.api.onmetal.de/v1alpha1.VolumeState">
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
<code>lastStateTransitionTime</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>LastStateTransitionTime is the last time the State transitioned between values.</p>
</td>
</tr>
<tr>
<td>
<code>access</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.VolumeAccess">
VolumeAccess
</a>
</em>
</td>
<td>
<p>Access specifies how to access a Volume.
This is set by the volume provider when the volume is provisioned.</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.VolumeCondition">
[]VolumeCondition
</a>
</em>
</td>
<td>
<p>Conditions are the conditions of a volume.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.api.onmetal.de/v1alpha1.VolumeTemplateSpec">VolumeTemplateSpec
</h3>
<div>
<p>VolumeTemplateSpec is the specification of a Volume template.</p>
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
<a href="#storage.api.onmetal.de/v1alpha1.VolumeSpec">
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
<code>volumeClassRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumeClassRef is the VolumeClass of a volume
If empty, an external controller has to provision the volume.</p>
</td>
</tr>
<tr>
<td>
<code>volumePoolSelector</code><br/>
<em>
map[string]string
</em>
</td>
<td>
<p>VolumePoolSelector selects a suitable VolumePoolRef by the given labels.</p>
</td>
</tr>
<tr>
<td>
<code>volumePoolRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>VolumePoolRef indicates which VolumePool to use for a volume.
If unset, the scheduler will figure out a suitable VolumePoolRef.</p>
</td>
</tr>
<tr>
<td>
<code>claimRef</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.LocalUIDReference">
github.com/onmetal/onmetal-api/api/common/v1alpha1.LocalUIDReference
</a>
</em>
</td>
<td>
<p>ClaimRef is the reference to the claiming entity of the Volume.</p>
</td>
</tr>
<tr>
<td>
<code>resources</code><br/>
<em>
<a href="../core/#core.api.onmetal.de/v1alpha1.ResourceList">
github.com/onmetal/onmetal-api/api/core/v1alpha1.ResourceList
</a>
</em>
</td>
<td>
<p>Resources is a description of the volume&rsquo;s resources and capacity.</p>
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
<p>Image is an optional image to bootstrap the volume with.</p>
</td>
</tr>
<tr>
<td>
<code>imagePullSecretRef</code><br/>
<em>
<a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#localobjectreference-v1-core">
Kubernetes core/v1.LocalObjectReference
</a>
</em>
</td>
<td>
<p>ImagePullSecretRef is an optional secret for pulling the image of a volume.</p>
</td>
</tr>
<tr>
<td>
<code>unclaimable</code><br/>
<em>
bool
</em>
</td>
<td>
<p>Unclaimable marks the volume as unclaimable.</p>
</td>
</tr>
<tr>
<td>
<code>tolerations</code><br/>
<em>
<a href="../common/#common.api.onmetal.de/v1alpha1.Toleration">
[]github.com/onmetal/onmetal-api/api/common/v1alpha1.Toleration
</a>
</em>
</td>
<td>
<p>Tolerations define tolerations the Volume has. Only any VolumePool whose taints
covered by Tolerations will be considered to host the Volume.</p>
</td>
</tr>
<tr>
<td>
<code>encryption</code><br/>
<em>
<a href="#storage.api.onmetal.de/v1alpha1.VolumeEncryption">
VolumeEncryption
</a>
</em>
</td>
<td>
<p>Encryption is an optional field which provides attributes to encrypt Volume.</p>
</td>
</tr>
</table>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
</em></p>
