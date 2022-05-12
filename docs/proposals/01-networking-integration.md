---
title: Networking Integration

oep-number: 1

creation-date: 2022-17-03

status: implementable

authors:

- @adracus
- @afritzler

reviewers:

- @adracus
- @afritzler
- @MalteJ
- @guvenc
- @gehoern

---

# OEP-1: Networking Integration

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Alternatives](#alternatives)

## Summary

Networking is a crucial part in a modern cloud system: It enables systems to communicate within themselves and to the
outside world. Orchestrating traffic, auditing it and gaining visibility of what is the desired state is key to a modern
network architecture.

Key of this OEP is to define the user-facing network API as well as its implications on any other type and the overall
structure of `onmetal`.

## Motivation

Without networking, any machine / process running inside a datacenter cannot interact / affect the outside world.
Networking is a crucial component that has to be implemented for onmetal to have business value. In a full-fledged
state, networking also enables security to the outside world and within a datacenter itself.

The basic use case we want to implement with onmetal is a machine that can access the internet and can be reached from
the internet.

### Goals

* Define APIs for managing isolated networks. It should be possible to do conflict-free peering of networks in the
  future.
* Define APIs for assigning / routing public IPs / prefixes to members of a network / subnet.
* Adapt the `compute.Machine` type to integrate with the network API.
* Have fully integrated IP address management for all resources (IPAM).
* It should be possible to extend the API in the future to achieve the following (listed by decreasing priority):
    * Regulate Communication within a subnet (plus security concepts)
    * Subnet-to-subnet communication (plus security concepts)
    * Isolated network-to-network communication (plus security concepts)
    * Cross-region isolated network-to-network communication (plus security concepts)

### Non-Goals

* Define Load Balancer APIs (L4 sooner in the future, L7 later)
* Implement any of the future API extensions listed above
* Allow a user to bring own public IP prefixes
* Feature-creep beyond a simplistic MVP

## Proposal

### Preface

As onmetal is Kubernetes-API, it should integrate nicely within the existing ecosystem. Some API design choices are made
in that regard. For further information about Kubernetes, see
[the Kubernetes reference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#ephemeralvolumesource-v1-core)
.

* Kubernetes specifies multiple ip types using `IPFamily`. This means, that instead of e.g. an object having
  ```yaml
  ipv4: 10.0.0.1
  ipv6: ffff::
  ```
  Kubernetes specifies it as
  ```yaml
  ipFamilies: [IPv4, IPv6]
  ips:
  - 10.0.0.1
  - ffff::
  ```
  The proposal should integrate into Kubernetes by using the same notation.
* Resources that are created, managed, and deleted in scope of another resource are called `ephemeral`. An example in
  Kubernetes is the `Pod.spec.volumes.ephemeralVolume` that creates a volume just before a
  `Pod` is created and deletes it alongside the `Pod` after usage.
* `1:1` binding between two resources is achieved by both resources referencing each other. This can be seen in
  Kubernetes'
  `PersistentVolumeClaim.Spec.volumeName` - `PersistentVolume.spec.claimRef`.
* `1:n` binding between two resources is achieved by the resource on `n` side having a reference to the resource on
  the `1` side. This can be seen in `n` `Pod.spec.nodeName` referencing a `Node`.
* `m:n` binding between two resources is achieved by using `selector`s and a 'binding' resource that usually gets
  created on-the-fly, though this also usually can be modified. An example can be seen in the relation between
  `Service`s and `Pod`s. A `Service` selects multiple `Pod`s via its `.spec.selector`. The resulting manifested binding
  resource is realized via the `Endpoints` kind that contains the current target list.

The proposal is divided into two parts: The first part purely focuses on IP address management. The second part defines
the actual networking types while allowing the user to use the IP address management features of the first part.

### `Prefix` type

The `Prefix` simplifies management of IP prefixes (v4 and v6 are both supported).

An `Prefix` may be a root prefix by specifying no parent / parent selector and a prefix it manages. If an `Prefix`
specifies a parent / parent selector, the requested prefix / prefix length is allocated from the parent (that matches,
if selector is used). This means, prefixes can both be allocated dynamically by specifying only a desired prefix length
or 'statically' by specifying the desired prefix.

Example manifests:

[//]: # (@formatter:off)
```yaml
apiVersion: ipam.onmetal.de/v1alpha1
kind: Prefix
metadata:
  namespace: default
  name: my-root-prefix
spec:
  prefix: 10.0.0.0/8
status:
  phase: Allocated
---
apiVersion: ipam.onmetal.de/v1alpha1
kind: Prefix
metadata:
  namespace: default
  name: my-sub-prefix
spec:
  parentRef:
    name: my-root-prefix
#  parentSelector: # A metav1.LabelSelector can be used to select the parent.
#    matchLabels:
#      foo: bar
  prefixLength: 16
  # prefix: 10.0.0.0/16 # Once successfully allocated, the spec is patched.
status:
  phase: Pending # This will become `Allocated` once the controller approves it.
```
[//]: # (@formatter:on)

### `Network` type

The namespaced `Network` type defines a `Network` bracket. Traffic from, to and within the `Network` can be managed.
A `Network` has to specify the ip families it wants to allow (same is design as in the Kubernetes `Service` type).

IP address space in a `Network` is not dictated in any way. A `Network` however has to accept any claimed IP address
space within it. For initial design, a `Network` will only accept non-overlapping space. In a later version, this may be
regulated with a field / policy of some kind.

Example manifest:

```yaml
apiVersion: networking.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: default
  name: my-network
```

### The `NetworkInterface` type

A `NetworkInterface` is the binding piece between a `Machine` and a `Network`. A `NetworkInterface` references the
`Network` it wants to join as well as the IPs it should use in that `Network`.

The IPs (v4 / v6) can be specified in multiple ways:

* Without IPAM by specifying an IP literal
* As `ephemeral`, creating an `ipam.Prefix` with the prefix length of the specified ip family (32 / 128) that will be
  owned and also deleted alongside the surrounding `NetworkInterface`. The name of the
  created `ipam.Prefix` will be `<nic-name>-<index>`, where `<index>` is the index of the `ephemeral` in the `ips`
  list. An existing `Prefix` with that name will *not* be used for the `NetworkInterface` to avoid using an unrelated
  `Prefix` by mistake.

When specifying IPs, a user should also specify `ipFamilies`. `ipFamilies` validates that there can be either a
single `IPv4` / `IPv6`, or an ordered list of an `IPv4` / `IPv6` address. If left empty and it can be deducted
deterministically from the `ips`, it will be defaulted. Same applies vice versa.

The binding between a `NetworkInterface` and a `Machine` is bidirectional via `NetworkInterface.spec.machineRef.name` /
`Machine.spec.networkInterfaces[*].name`. For the mvp, we will only allow exactly **1** `NetworkInterface` per `Machine`
.

Example usage:

[//]: # (@formatter:off)
```yaml
apiVersion: networking.onmetal.de/v1alpha1
kind: NetworkInterface
metadata:
  namespace: default
  name: my-machine-interface
spec:
  networkRef:
    name: my-network
  ipFamilies: [IPv4, IPv6]
  ips:
#    - value: 10.0.0.1 # It is also possible to directly specify IPs without IPAM 
#    - value: 2607:f0d0:1002:51::4 # Same applies for v6 addresses
    - ephemeral:
        prefixTemplate:
          spec:
            prefixRef:
              name: my-node-prefix-v4
    - ephemeral:
        prefixTemplate:
          spec:
            prefixRef:
              name: my-node-prefix-v6
  machineRef:
    name: my-machine
status:
  ips: # This will be updated with the allocated addresses.
    - 10.0.0.1
    - 2607:f0d0:1002:51::4
---
apiVersion: compute.onmetal.de/v1alpha1
kind: Machine
metadata:
  namespace: default
  name: my-machine
  labels:
    app: web
spec:
  networkInterfaces:
    - name: my-interface
      networkInterfaceRef:
        name: my-machine-interface
  ...
status:
  networkInterfaces:
    - name: my-interface
      ips: # The machine reports all ips available via its interfaces
        - 10.0.0.1
        - 2607:f0d0:1002:51::4
  ...
```
[//]: # (@formatter:on)

To simplify managing the creation of a `NetworkInterface` per `Machine`, a `Machine` can specify a `NetworkInterface`
as `ephemeral`, creating and owning it before the `Machine` becomes available. The name of the
`NetworkInterface` will be `<machine-name>-<name>` where `<name>` is the `name:` value in the `networkInterfaces` list.
Existing `NetworkInterface`s will not be adopted by the `Machine`.

Sample manifest:

[//]: # (@formatter:off)
```yaml
apiVersion: compute.onmetal.de/v1alpha1
kind: Machine
spec:
  interfaces:
    - name: my-interface
      ephemeral:
        networkInterfaceTemplate:
          spec:
            ipFamilies: [IPv4, IPv6]
            networkRef:
              name: my-network
            ips:
              - ephemeral:
                  prefixTemplate:
                    spec:
                      prefixRef:
                        name: my-node-prefix-v4
              - ephemeral:
                  prefixTemplate:
                    spec:
                      prefixRef:
                        name: my-node-prefix-v6
  ...
```
[//]: # (@formatter:on)

### The `AliasPrefix` type.

An `AliasPrefix` allows routing a sub-prefix of a network to multiple targets (in our case, `NetworkInterface`s). It
references its target `Network` and selects the `NetworkInterfaces` via its `selector` to apply the alias to.

The `AliasPrefix` creates an `AliasPrefixRouting` object with the same name as itself where it maintains a list of
the `NetworkInterface`s matching its `selector`. If the `selector` is empty it is assumed that an external process
manages the `AliasPrefixRouting` belonging to that `AliasPrefix`.

Example manifest:

[//]: # (@formatter:off)
```yaml
apiVersion: networking.onmetal.de/v1alpha1
kind: AliasPrefix
metadata:
  namespace: default
  name: my-pod-prefix-1
spec:
  ipFamily: IPv4
  networkRef:
    name: my-network
  networkInterfaceSelector:
    matchLabels:
      foo: bar
  prefix:
#    value: 10.0.0.0/24 # It's possible to directly specify the AliasPrefix value
    ephemeral:
      prefixTemplate:
        spec:
          prefixRef:
            name: my-pod-prefix
          prefixLength: 24
status:
  prefix: 10.0.0.0/24
```
[//]: # (@formatter:on)

This could manifest in the following `AliasPrefixRouting`:

[//]: # (@formatter:off)
```yaml
apiVersion: networking.onmetal.de/v1alpha1
kind: AliasPrefixRouting
metadata:
  namespace: default
  name: my-pod-prefix-1
networkRef:
  name: my-network
destinations:
  - name: my-machine-interface-1
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
  - name: my-machine-interface-2
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73d
```
[//]: # (@formatter:on)

To simplify the creation and use of an `AliasPrefix` per `NetworkInterface`, a `NetworkInterface`
allows the creation via `ephemeralAliasPrefixes`. The resulting `AliasPrefix` name will be
`<nic-name>-<name>` where `<name>` is the name in the `ephemeralAliasPrefixes` list.

It will also automatically be set in the same network and only target the hosting `NetworkInterface`.
`selector` and `networkRef` in `spec` thus cannot be specified.

[//]: # (@formatter:off)
```yaml
apiVersion: networking.onmetal.de/v1alpha1
kind: NetworkInterface
metadata:
  namespace: default
  name: my-machine-interface
spec:
  ephemeralAliasPrefixes:
    - name: podrange-v4
      spec:
        prefix:
#          value: 10.0.0.0/24 # It's possible to directly specify the AliasPrefix value
          ephemeralPrefix:
            spec:
              prefixRef:
                name: my-pod-prefix
              prefixLength: 24
```
[//]: # (@formatter:on)

### The `VirtualIP` type

A `VirtualIP` requests a stable public IP for a single targets (`NetworkInterface`s). There is a
`type` field that currently only can be `type: Public` in order to support other future `VirtualIP` types (for instance,
`VirtualIP`s in other networks).

As the public prefixes are provider-managed and custom public IP pools are not in scope of this draft, the IP allocation
cannot be influenced and thus no construct like `prefixRef` is possible for `VirtualIP`s.

To disambiguate between IPv4 and IPv6, the `VirtualIP` requires an `ipFamily` (same enum type as in Kubernetes'
`Service.spec.ipFamilies`).

The `VirtualIP` references the claiming `NetworkInterface` using `targetRef`.

Example manifest:

[//]: # (@formatter:off)
```yaml
apiVersion: networking.onmetal.de
kind: VirtualIP
metadata:
  namespace: default
  name: my-virtual-ip
spec:
  type: Public
  ipFamily: IPv4
  targetRef:
    name: my-nic
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73d
status:
  ip: 45.86.152.88
  phase: Bound
```
[//]: # (@formatter:on)

To simplify the creation and use of a `VirtualIP` per `NetworkInterface`, a `NetworkInterface`
allows the creation via `virtualIP.ephemeral`. The resulting `VirtualPrefix` name will be
`<nic-name>` It will also automatically be set up to reference the creating `NetworkInterface`.
A `networkInterfaceRef` in the `spec` thus cannot be specified.

[//]: # (@formatter:off)
```yaml
apiVersion: networking.onmetal.de/v1alpha1
kind: NetworkInterface
metadata:
  namespace: default
  name: my-machine-interface
spec:
  virtualIP:
    ephemeral:
      virtualIPTemplate:
        spec:
          type: Public
          ipFamily: IPv4
```
[//]: # (@formatter:on)

## Scenarios

### Kubernetes (Gardener) integration on top of onmetal

For a Kubernetes integration, multiple worker nodes should be created in the same network. For each worker node, a
separate pod prefix should be allocated. For internet-facing requests, each node should get a distinct
public `VirtualIP` (in the future, outgoing requests will be solved via `SNAT`, but for the initial version of the MVP
a `VirtualIP` is chosen).

Additionally, to show it's possible, an `AliasPrefix` that is shared across different nodes is created.

These are the required manifests:

[//]: # (@formatter:off)
```yaml
# IPAM Setup:
# Create a root prefix and a pod / node sub-prefix.
apiVersion: ipam.onmetal.de/v1alpha1
kind: Prefix
metadata:
  namespace: default
  name: root
spec:
  prefix: 10.0.0.0/8
---
apiVersion: ipam.onmetal.de/v1alpha1
kind: Prefix
metadata:
  namespace: default
  name: pods
spec:
  prefixLength: 11
  parentRef:
    name: root
---
apiVersion: ipam.onmetal.de/v1alpha1
kind: Prefix
metadata:
  namespace: default
  name: nodes
spec:
  prefixLength: 16
  parentRef:
    name: root
---
# Once IPAM is done, the concrete networking is defined
# The Network is the bracket around all resources
apiVersion: networking.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: default
  name: k8s
---
# Create one prefix that should be shared across all machines
apiVersion: networking.onmetal.de/v1alpha1
kind: AliasPrefix
metadata:
  namespace: default
  name: shared
spec:
  networkRef:
    name: k8s
  networkInterfaceSelector:
    matchLabels:
      type: k8s-worker
  prefix:
    ephemeral:
      prefixTemplate:
        spec:
          prefixRef:
            name: k8s
          prefixLength: 16
---
# Create the actual machine
apiVersion: compute.onmetal.de/v1alpha1
kind: Machine
metadata:
  namespace: default
  name: worker-1
  labels:
    type: k8s-worker
spec:
  image: gardenlinux-k8s-worker:v0.23.5
  networkInterfaces:
    - name: primary
      ephemeral:
        networkInterfaceTemplate:
          spec:
            # Let the nic join the network
            networkRef:
              name: k8s
            # The IP should be allocated from the node range
            ips:
              - ephemeral:
                  prefixTemplate:
                    spec:
                      ipFamily: IPv4
                      prefixRef:
                        name: nodes
            # Create a pod alias range exclusively for this machine
            ephemeralAliasPrefixes:
              - name: pods
                spec:
                  prefix:
                    ephemeral:
                      prefixTemplate:
                        ipFamily: IPv4
                        spec:
                          prefixRef:
                            name: pods
                          prefixLength: 24
            # Create a virtual IP for this machine
            virtualIP:
              - ephemeral:
                  virtualIPTemplate:
                    spec:
                      type: Public
                      ipFamily: IPv4
```
[//]: # (@formatter:on)

## Alternatives

None discussed so far.
