<p>Packages:</p>
<ul>
<li>
<a href="#common.onmetal.de%2fv1alpha1">common.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="common.onmetal.de/v1alpha1">common.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="common.onmetal.de/v1alpha1.Availability">Availability
(<code>[]..RegionAvailability</code> alias)</h3>
<div>
</div>
<h3 id="common.onmetal.de/v1alpha1.Cidr">Cidr
(<code>string</code> alias)</h3>
<div>
<p>TODO: create marshal/unmarshal functions</p>
</div>
<h3 id="common.onmetal.de/v1alpha1.IPAddr">IPAddr
(<code>string</code> alias)</h3>
<div>
<p>TODO: create marshal/unmarshal functions</p>
</div>
<h3 id="common.onmetal.de/v1alpha1.Location">Location
</h3>
<div>
<p>Location describes the location of a resource</p>
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
<p>Region defines the region of a resource</p>
</td>
</tr>
<tr>
<td>
<code>availabilityZone</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>AvailabilityZone is the availability zone of a resource</p>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.RegionAvailability">RegionAvailability
</h3>
<div>
<p>RegionAvailability defines a region with its availability zones</p>
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
<p>Region is the name of the region</p>
</td>
</tr>
<tr>
<td>
<code>availabilityZones</code><br/>
<em>
<a href="#common.onmetal.de/v1alpha1.ZoneAvailability">
[]ZoneAvailability
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Zones is a list of zones in this region</p>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.ScopedKindReference">ScopedKindReference
</h3>
<div>
<p>ScopedKindReference defines an object with its kind and API group and its scope reference</p>
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
<code>kind</code><br/>
<em>
string
</em>
</td>
<td>
<p>Kind is the kind of the object</p>
</td>
</tr>
<tr>
<td>
<code>apiGroup</code><br/>
<em>
string
</em>
</td>
<td>
<p>APIGroup is the API group of the object</p>
</td>
</tr>
<tr>
<td>
<code>ScopedReference</code><br/>
<em>
<a href="#common.onmetal.de/v1alpha1.ScopedReference">
ScopedReference
</a>
</em>
</td>
<td>
<p>
(Members of <code>ScopedReference</code> are embedded into this type.)
</p>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.ScopedReference">ScopedReference
</h3>
<p>
(<em>Appears on:</em><a href="#common.onmetal.de/v1alpha1.ScopedKindReference">ScopedKindReference</a>)
</p>
<div>
<p>ScopedReference refers to a scope and the scopes name</p>
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
<p>Name is the name of the scope</p>
</td>
</tr>
<tr>
<td>
<code>scope</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>Scope is the absolute scope path</p>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.StateFields">StateFields
</h3>
<div>
<p>StateFields defines the observed state of an object</p>
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
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>State indicates the state of a resource</p>
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
<em>(Optional)</em>
<p>Message contains a message for the corresponding state</p>
</td>
</tr>
<tr>
<td>
<code>conditions</code><br/>
<em>
<a href="https://v1-21.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#condition-v1-meta">
[]Kubernetes meta/v1.Condition
</a>
</em>
</td>
<td>
<em>(Optional)</em>
<p>Conditions represents the status for individual operators</p>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.ZoneAvailability">ZoneAvailability
</h3>
<p>
(<em>Appears on:</em><a href="#common.onmetal.de/v1alpha1.RegionAvailability">RegionAvailability</a>)
</p>
<div>
<p>ZoneAvailability defines the name of a zone</p>
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
<p>Name is the name of the availability zone</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>aa5aba9</code>.
</em></p>
