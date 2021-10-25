<p>Packages:</p>
<ul>
<li>
<a href="#common.onmetal.de%2fv1alpha1">common.onmetal.de/v1alpha1</a>
</li>
</ul>
<h2 id="common.onmetal.de/v1alpha1">common.onmetal.de/v1alpha1</h2>
Resource Types:
<ul></ul>
<h3 id="common.onmetal.de/v1alpha1.CIDR">CIDR
</h3>
<div>
<p>CIDR represents a network CIDR.</p>
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
<code>-</code><br/>
<em>
inet.af/netaddr.IPPrefix
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.ConfigMapKeySelector">ConfigMapKeySelector
</h3>
<div>
<p>ConfigMapKeySelector is a reference to a specific &lsquo;key&rsquo; within a ConfigMap resource.
In some instances, <code>key</code> is a required field.</p>
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
<p>The name of the ConfigMap resource being referred to.</p>
</td>
</tr>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The key of the entry in the ConfigMap resource&rsquo;s <code>data</code> field to be used.
Some instances of this field may be defaulted, in others it may be
required.</p>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.IPAddr">IPAddr
</h3>
<p>
(<em>Appears on:</em><a href="#common.onmetal.de/v1alpha1.IPRange">IPRange</a>)
</p>
<div>
<p>IPAddr is an IP address.</p>
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
<code>-</code><br/>
<em>
inet.af/netaddr.IP
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.IPRange">IPRange
</h3>
<div>
<p>IPRange is an IP range.</p>
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
<code>from</code><br/>
<em>
<a href="#common.onmetal.de/v1alpha1.IPAddr">
IPAddr
</a>
</em>
</td>
<td>
</td>
</tr>
<tr>
<td>
<code>to</code><br/>
<em>
<a href="#common.onmetal.de/v1alpha1.IPAddr">
IPAddr
</a>
</em>
</td>
<td>
</td>
</tr>
</tbody>
</table>
<h3 id="common.onmetal.de/v1alpha1.SecretKeySelector">SecretKeySelector
</h3>
<div>
<p>SecretKeySelector is a reference to a specific &lsquo;key&rsquo; within a Secret resource.
In some instances, <code>key</code> is a required field.</p>
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
<p>The name of the Secret resource being referred to.</p>
</td>
</tr>
<tr>
<td>
<code>key</code><br/>
<em>
string
</em>
</td>
<td>
<em>(Optional)</em>
<p>The key of the entry in the Secret resource&rsquo;s <code>data</code> field to be used.
Some instances of this field may be defaulted, in others it may be
required.</p>
</td>
</tr>
</tbody>
</table>
<hr/>
<p><em>
Generated with <code>gen-crd-api-reference-docs</code>
on git commit <code>a435c36</code>.
</em></p>
