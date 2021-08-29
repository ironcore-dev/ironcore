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
<a href="#storage.onmetal.de/v1alpha1.StorageClassCapability">
[]StorageClassCapability
</a>
</em>
</td>
<td>
<p>Capabilities describes the capabilities of a storage class</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<p>Description is a human readable description of a storage class</p>
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
<h3 id="storage.onmetal.de/v1alpha1.StorageClassCapability">StorageClassCapability
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StorageClassSpec">StorageClassSpec</a>)
</p>
<div>
<p>StorageClassCapability describes one attribute of the StorageClass</p>
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
<p>Name is the name of a capability</p>
</td>
</tr>
<tr>
<td>
<code>value</code><br/>
<em>
k8s.io/apimachinery/pkg/util/intstr.IntOrString
</em>
</td>
<td>
<p>Value is the value of a capability</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.StorageClassCapacity">StorageClassCapacity
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.StoragePoolSpec">StoragePoolSpec</a>, <a href="#storage.onmetal.de/v1alpha1.StoragePoolStatus">StoragePoolStatus</a>)
</p>
<div>
<p>StorageClassCapacity defines capacity attribute of a storage class</p>
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
<p>Name is the name of a storage class capacity</p>
</td>
</tr>
<tr>
<td>
<code>capacity</code><br/>
<em>
<a href="https://pkg.go.dev/k8s.io/apimachinery/pkg/api/resource#Duration">
k8s.io/apimachinery/pkg/api/resource.Quantity
</a>
</em>
</td>
<td>
<p>Capacity is the quantity of a capacity</p>
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
<a href="#storage.onmetal.de/v1alpha1.StorageClassCapability">
[]StorageClassCapability
</a>
</em>
</td>
<td>
<p>Capabilities describes the capabilities of a storage class</p>
</td>
</tr>
<tr>
<td>
<code>description</code><br/>
<em>
string
</em>
</td>
<td>
<p>Description is a human readable description of a storage class</p>
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
<code>availability</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Availability">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.Availability
</a>
</em>
</td>
<td>
<p>Availability describes the regions and zones where this MachineClass is available</p>
</td>
</tr>
</tbody>
</table>
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
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region defines the region of the storage pool</p>
</td>
</tr>
<tr>
<td>
<code>privacy</code><br/>
<em>
string
</em>
</td>
<td>
<p>Privacy defines the privacy scope of the storage pool</p>
</td>
</tr>
<tr>
<td>
<code>replication</code><br/>
<em>
int
</em>
</td>
<td>
<p>Replication indicates the replication factor in the storage pool</p>
</td>
</tr>
<tr>
<td>
<code>capacity</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.StorageClassCapacity">
[]StorageClassCapacity
</a>
</em>
</td>
<td>
<p>Capacity list the available capacity of a storage pool</p>
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
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region defines the region of the storage pool</p>
</td>
</tr>
<tr>
<td>
<code>privacy</code><br/>
<em>
string
</em>
</td>
<td>
<p>Privacy defines the privacy scope of the storage pool</p>
</td>
</tr>
<tr>
<td>
<code>replication</code><br/>
<em>
int
</em>
</td>
<td>
<p>Replication indicates the replication factor in the storage pool</p>
</td>
</tr>
<tr>
<td>
<code>capacity</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.StorageClassCapacity">
[]StorageClassCapacity
</a>
</em>
</td>
<td>
<p>Capacity list the available capacity of a storage pool</p>
</td>
</tr>
</tbody>
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
<code>used</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.StorageClassCapacity">
[]StorageClassCapacity
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
<code>storage_class</code><br/>
<em>
string
</em>
</td>
<td>
<p>StorageClass is the storage class of a volume</p>
</td>
</tr>
<tr>
<td>
<code>storagepool</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>StoragePool indicates which storage pool to use for a volume</p>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Machine is a reference of the machine object which the volume should be attached to</p>
</td>
</tr>
<tr>
<td>
<code>device</code><br/>
<em>
string
</em>
</td>
<td>
<p>Device defines the device on the host for a volume</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.VolumeSource">
VolumeSource
</a>
</em>
</td>
<td>
<p>Source references either an image or a snapshot</p>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Machine is a reference of the machine object which the volume should be attached to</p>
</td>
</tr>
<tr>
<td>
<code>device</code><br/>
<em>
string
</em>
</td>
<td>
<p>Device defines the device on the host for a volume</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
<a href="#storage.onmetal.de/v1alpha1.VolumeSource">
VolumeSource
</a>
</em>
</td>
<td>
<p>Source references either an image or a snapshot</p>
</td>
</tr>
</tbody>
</table>
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
<code>device</code><br/>
<em>
string
</em>
</td>
<td>
<p>Device describes the device of the volume on the host</p>
</td>
</tr>
</tbody>
</table>
<h3 id="storage.onmetal.de/v1alpha1.VolumeSource">VolumeSource
</h3>
<p>
(<em>Appears on:</em><a href="#storage.onmetal.de/v1alpha1.VolumeAttachmentSpec">VolumeAttachmentSpec</a>)
</p>
<div>
<p>VolumeSource defines the source of a volume which can be either an image or a snapshot</p>
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
<code>image</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Image defines the image name of the referenced image</p>
</td>
</tr>
<tr>
<td>
<code>snapshot</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Snapshot defines the snapshot which should be used</p>
</td>
</tr>
</tbody>
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
<code>storage_class</code><br/>
<em>
string
</em>
</td>
<td>
<p>StorageClass is the storage class of a volume</p>
</td>
</tr>
<tr>
<td>
<code>storagepool</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>StoragePool indicates which storage pool to use for a volume</p>
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
</tbody>
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
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>aa5aba9</code>.
</em></p>
