<p>Packages:</p>
<ul>
<li>
<a href="#compute.onmetal.de%2fv1alpha1">compute.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="compute.onmetal.de/v1alpha1">compute.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="compute.onmetal.de/v1alpha1.AvailabilityZoneQuantity">AvailabilityZoneQuantity
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachinePoolSpec">MachinePoolSpec</a>, <a href="#compute.onmetal.de/v1alpha1.MachinePoolStatus">MachinePoolStatus</a>)
</p>
<div>
<p>AvailabilityZoneQuantity defines the quantity of available MachineClasses in a given AZ</p>
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
<code>availabilityZone</code><br/>
<em>
string
</em>
</td>
<td>
<p>AvailabilityZone is the name of the availability zone</p>
</td>
</tr>
<tr>
<td>
<code>classes</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.MachineClassQuantity">
[]MachineClassQuantity
</a>
</em>
</td>
<td>
<p>Classes defines a list of machine classes and their corresponding quantities</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.Capability">Capability
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineClassSpec">MachineClassSpec</a>)
</p>
<div>
<p>Capability describes a single feature of a MachineClass</p>
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
<p>Name is the name of the capability</p>
</td>
</tr>
<tr>
<td>
<code>type</code><br/>
<em>
string
</em>
</td>
<td>
<p>Type defines the type of the capability</p>
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
<p>Value is the effective value of the capability</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.Flag">Flag
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.ImageSpec">ImageSpec</a>)
</p>
<div>
<p>Flag is a single value pair</p>
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
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<p>Key is the key name</p>
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
<p>Value contains the value for a key</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.Hash">Hash
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.SourceAttribute">SourceAttribute</a>)
</p>
<div>
<p>Hash describes a hash value and it&rsquo;s corresponding algorithm</p>
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
<code>algorithm</code><br/>
<em>
string
</em>
</td>
<td>
<p>Algorithm indicates the algorithm with which the hash should be computed</p>
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
<p>Value is the computed hash value</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.HashStatus">HashStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.ImageStatus">ImageStatus</a>)
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
<code>hash</code><br/>
<em>
string
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.Image">Image
</h3>
<div>
<p>Image is the Schema for the images API</p>
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
<a href="#compute.onmetal.de/v1alpha1.ImageSpec">
ImageSpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>arch</code><br/>
<em>
string
</em>
</td>
<td>
<p>Arch describes the architecture the Image is built for</p>
</td>
</tr>
<tr>
<td>
<code>maturity</code><br/>
<em>
string
</em>
</td>
<td>
<p>Maturity defines the maturity of an Image. It indicates whether this Image is e.g. a stable or preview version.</p>
</td>
</tr>
<tr>
<td>
<code>expirationTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>ExpirationTime defines when the support for this image will expire</p>
</td>
</tr>
<tr>
<td>
<code>os</code><br/>
<em>
string
</em>
</td>
<td>
<p>OS defines the operating system name of the image</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version defines the operating system version</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.SourceAttribute">
[]SourceAttribute
</a>
</em>
</td>
<td>
<p>Source defines the source artefacts and their corresponding location</p>
</td>
</tr>
<tr>
<td>
<code>imageRef</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>ImageRef is a scoped reference to an existing Image</p>
</td>
</tr>
<tr>
<td>
<code>flags</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.Flag">
[]Flag
</a>
</em>
</td>
<td>
<p>Flags is a generic key value pair used for defining Image hints</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.ImageStatus">
ImageStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.ImageSpec">ImageSpec
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.Image">Image</a>)
</p>
<div>
<p>ImageSpec defines the desired state of Image</p>
<p>Either a Source or an ImageRef should be defined to describe the content of an Image</p>
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
<code>arch</code><br/>
<em>
string
</em>
</td>
<td>
<p>Arch describes the architecture the Image is built for</p>
</td>
</tr>
<tr>
<td>
<code>maturity</code><br/>
<em>
string
</em>
</td>
<td>
<p>Maturity defines the maturity of an Image. It indicates whether this Image is e.g. a stable or preview version.</p>
</td>
</tr>
<tr>
<td>
<code>expirationTime</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>ExpirationTime defines when the support for this image will expire</p>
</td>
</tr>
<tr>
<td>
<code>os</code><br/>
<em>
string
</em>
</td>
<td>
<p>OS defines the operating system name of the image</p>
</td>
</tr>
<tr>
<td>
<code>version</code><br/>
<em>
string
</em>
</td>
<td>
<p>Version defines the operating system version</p>
</td>
</tr>
<tr>
<td>
<code>source</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.SourceAttribute">
[]SourceAttribute
</a>
</em>
</td>
<td>
<p>Source defines the source artefacts and their corresponding location</p>
</td>
</tr>
<tr>
<td>
<code>imageRef</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>ImageRef is a scoped reference to an existing Image</p>
</td>
</tr>
<tr>
<td>
<code>flags</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.Flag">
[]Flag
</a>
</em>
</td>
<td>
<p>Flags is a generic key value pair used for defining Image hints</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.ImageStatus">ImageStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.Image">Image</a>)
</p>
<div>
<p>ImageStatus defines the observed state of Image</p>
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
<code>hashes</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.HashStatus">
[]HashStatus
</a>
</em>
</td>
<td>
<p>Hashes lists all hashes for all included artefacts</p>
</td>
</tr>
<tr>
<td>
<code>regions</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.RegionState">
[]RegionState
</a>
</em>
</td>
<td>
<p>Regions indicates the availability of the image in the corresponding regions</p>
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>MachineClass is the machine class/flavor of the machine</p>
</td>
</tr>
<tr>
<td>
<code>machinePool</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>MachinePool defines the compute pool of the machine</p>
</td>
</tr>
<tr>
<td>
<code>location</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Location">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.Location
</a>
</em>
</td>
<td>
<p>Location is the physical location of the machine</p>
</td>
</tr>
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
<p>Image is the operating system image of the machine</p>
</td>
</tr>
<tr>
<td>
<code>sshPublicKeys</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.SSHPublicKeyEntry">
[]SSHPublicKeyEntry
</a>
</em>
</td>
<td>
<p>SSHPublicKeys is a list of SSH public keys of a machine</p>
</td>
</tr>
<tr>
<td>
<code>securityGroups</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Interfaces define a list of network interfaces present on the machine
TODO: define interfaces/network references
SecurityGroups is a list of security groups of a machine</p>
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
<code>userData</code><br/>
<em>
string
</em>
</td>
<td>
<p>UserData defines the ignition file</p>
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
<a href="#compute.onmetal.de/v1alpha1.Capability">
[]Capability
</a>
</em>
</td>
<td>
<p>Capabilities describes the features of the MachineClass</p>
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
<h3 id="compute.onmetal.de/v1alpha1.MachineClassQuantity">MachineClassQuantity
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.AvailabilityZoneQuantity">AvailabilityZoneQuantity</a>)
</p>
<div>
<p>MachineClassQuantity defines the quantity of a given MachineClass</p>
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
<p>Name is the name of the machine class quantity</p>
</td>
</tr>
<tr>
<td>
<code>quantity</code><br/>
<em>
int
</em>
</td>
<td>
<p>Quantity is an absolut number of the available machine class</p>
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
<a href="#compute.onmetal.de/v1alpha1.Capability">
[]Capability
</a>
</em>
</td>
<td>
<p>Capabilities describes the features of the MachineClass</p>
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
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region defines the region where this machine pool is available</p>
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
<p>Privacy indicates the privacy scope of the machine pool</p>
</td>
</tr>
<tr>
<td>
<code>capacity</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.AvailabilityZoneQuantity">
[]AvailabilityZoneQuantity
</a>
</em>
</td>
<td>
<p>Capacity defines the quantity of this machine pool per availability zone</p>
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
<code>region</code><br/>
<em>
string
</em>
</td>
<td>
<p>Region defines the region where this machine pool is available</p>
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
<p>Privacy indicates the privacy scope of the machine pool</p>
</td>
</tr>
<tr>
<td>
<code>capacity</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.AvailabilityZoneQuantity">
[]AvailabilityZoneQuantity
</a>
</em>
</td>
<td>
<p>Capacity defines the quantity of this machine pool per availability zone</p>
</td>
</tr>
</tbody>
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
<a href="#compute.onmetal.de/v1alpha1.AvailabilityZoneQuantity">
AvailabilityZoneQuantity
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>MachineClass is the machine class/flavor of the machine</p>
</td>
</tr>
<tr>
<td>
<code>machinePool</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>MachinePool defines the compute pool of the machine</p>
</td>
</tr>
<tr>
<td>
<code>location</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.Location">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.Location
</a>
</em>
</td>
<td>
<p>Location is the physical location of the machine</p>
</td>
</tr>
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
<p>Image is the operating system image of the machine</p>
</td>
</tr>
<tr>
<td>
<code>sshPublicKeys</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.SSHPublicKeyEntry">
[]SSHPublicKeyEntry
</a>
</em>
</td>
<td>
<p>SSHPublicKeys is a list of SSH public keys of a machine</p>
</td>
</tr>
<tr>
<td>
<code>securityGroups</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
[]github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Interfaces define a list of network interfaces present on the machine
TODO: define interfaces/network references
SecurityGroups is a list of security groups of a machine</p>
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
<code>userData</code><br/>
<em>
string
</em>
</td>
<td>
<p>UserData defines the ignition file</p>
</td>
</tr>
</tbody>
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
<h3 id="compute.onmetal.de/v1alpha1.RegionState">RegionState
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.ImageStatus">ImageStatus</a>)
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
<h3 id="compute.onmetal.de/v1alpha1.SSHPublicKey">SSHPublicKey
</h3>
<div>
<p>SSHPublicKey is the Schema for the sshpublickeys API</p>
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
<a href="#compute.onmetal.de/v1alpha1.SSHPublicKeySpec">
SSHPublicKeySpec
</a>
</em>
</td>
<td>
<br/>
<br/>
<table>
<tr>
<td>
<code>sshPublicKey</code><br/>
<em>
string
</em>
</td>
<td>
<p>SSHPublicKey is the SSH public key string</p>
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
<p>Description describes the purpose of the ssh key</p>
</td>
</tr>
<tr>
<td>
<code>expirationDate</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>ExpirationDate indicates until when this public key is valid</p>
</td>
</tr>
</table>
</td>
</tr>
<tr>
<td>
<code>status</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.SSHPublicKeyStatus">
SSHPublicKeyStatus
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.SSHPublicKeyEntry">SSHPublicKeyEntry
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.MachineSpec">MachineSpec</a>)
</p>
<div>
<p>SSHPublicKeyEntry describes either a reference to a SSH public key or a selector
to filter for a public key</p>
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
<code>scope</code><br/>
<em>
string
</em>
</td>
<td>
<p>Scope is the scope of a SSH public key</p>
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
<p>Name is the name of the SSH public key</p>
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
<p>Selector defines a LabelSelector to filter for a public key</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.SSHPublicKeySpec">SSHPublicKeySpec
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.SSHPublicKey">SSHPublicKey</a>)
</p>
<div>
<p>SSHPublicKeySpec defines the desired state of SSHPublicKey</p>
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
<code>sshPublicKey</code><br/>
<em>
string
</em>
</td>
<td>
<p>SSHPublicKey is the SSH public key string</p>
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
<p>Description describes the purpose of the ssh key</p>
</td>
</tr>
<tr>
<td>
<code>expirationDate</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#time-v1-meta">
Kubernetes meta/v1.Time
</a>
</em>
</td>
<td>
<p>ExpirationDate indicates until when this public key is valid</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.SSHPublicKeyStatus">SSHPublicKeyStatus
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.SSHPublicKey">SSHPublicKey</a>)
</p>
<div>
<p>SSHPublicKeyStatus defines the observed state of SSHPublicKey</p>
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
<code>fingerPrint</code><br/>
<em>
string
</em>
</td>
<td>
<p>FingerPrint is the finger print of the ssh public key</p>
</td>
</tr>
<tr>
<td>
<code>keyLength</code><br/>
<em>
int
</em>
</td>
<td>
<p>KeyLength is the byte length of the ssh key</p>
</td>
</tr>
<tr>
<td>
<code>algorithm</code><br/>
<em>
string
</em>
</td>
<td>
<p>Algorithm is the algorithm used to generate the ssh key</p>
</td>
</tr>
<tr>
<td>
<code>publicKey</code><br/>
<em>
string
</em>
</td>
<td>
<p>PublicKey is the PEM encoded public key</p>
</td>
</tr>
</tbody>
</table>
<h3 id="compute.onmetal.de/v1alpha1.SourceAttribute">SourceAttribute
</h3>
<p>
(<em>Appears on:</em><a href="#compute.onmetal.de/v1alpha1.ImageSpec">ImageSpec</a>)
</p>
<div>
<p>SourceAttribute describes the source components of an Image</p>
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
<p>Name defines the name of a source element</p>
</td>
</tr>
<tr>
<td>
<code>url</code><br/>
<em>
string
</em>
</td>
<td>
<p>URL defines the location of the image artefact</p>
</td>
</tr>
<tr>
<td>
<code>hash</code><br/>
<em>
<a href="#compute.onmetal.de/v1alpha1.Hash">
Hash
</a>
</em>
</td>
<td>
<p>Hash is the computed hash value of the artefacts content</p>
</td>
</tr>
</tbody>
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
<code>device</code><br/>
<em>
string
</em>
</td>
<td>
<p>Device defines the device for a volume on the machine</p>
</td>
</tr>
<tr>
<td>
<code>storageClass</code><br/>
<em>
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
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
<a href="/api-reference/common/#common.onmetal.de/v1alpha1.ScopedReference">
github.com/onmetal/onmetal-api/apis/common/v1alpha1.ScopedReference
</a>
</em>
</td>
<td>
<p>Volume is a reference to an existing volume</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>fbe0128</code>.
</em></p>
