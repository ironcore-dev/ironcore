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
* It should be possible to extend the API in the future to achieve the following (listed by decreasing priority):
    * Regulate Communication within a subnet (plus security concepts)
    * Subnet-to-subnet communication (plus security concepts)
    * Isolated network-to-network communication (plus security concepts)
    * Cross-region isolated network-to-network communicaction (plus security concepts)

### Non-Goals

* Define Load Balancer APIs (L4 sooner in the future, L7 later)
* Implement any of the future API extensions listed above
* Allow a user to bring own public IP prefixes
* Feature-creep beyond a simplistic MVP

## Proposal

### `Network` type

The namespaced `Network` type defines a `Network` bracket. Traffic from, to and within the `Network` can be managed.
A `Network` has to specify the ip families it wants to allow (same is design as in the Kubernetes `Service` type).

IP address space in a `Network` is not dictated in any way. A `Network` however has to accept any claimed IP address
space within it. For initial design, a `Network` will only accept non-overlapping space. In a later version, this may be
regulated with e.g. a `.spec.ipPolicy`.

Example manifest:

```yaml
apiVersion: network.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: default
  name: my-network
spec:
  ipFamilies:
    - IPv4
    - IPv6
status:
  used:
    - 192.168.178.1/32
    - 2607:f0d0:1002:51::4/128
    - 192.168.179.0/24
    - 10.5.3.7/32
  conditions:
    - type: Available # Available may be the name for this condition, though this has to be refined.
      status: True
      reason: PrefixAllocated
```

### The `NetworkInterface` type

A `NetworkInterface` lets a `Machine` join a `Network`. As such, in its `spec`, it references a `Network` and
the `Prefix` it shall allocate an IP from.

The binding between a `NetworkInterface` and a `Machine` is bidirectional via `NetworkInterface.spec.machineRef.name` /
`Machine.spec.interfaces[*].name`. For the mvp, we will only allow exactly **1** `NetworkInterface` per `Machine`.

Example usage:

```yaml
apiVersion: compute.onmetal.de/v1alpha1
kind: NetworkInterface
metadata:
  namespace: default
  name: my-machine-interface
spec:
  networkRef:
    name: my-network
  prefixRef:
    name: my-node-prefix
  machineRef:
    name: my-machine
status:
  primaryIPs:
    - 192.168.178.1
    - 2607:f0d0:1002:51::4
  conditions:
    - type: Available
      status: True
      reason: JoinedNetwork
---
apiVersion: compute.onmetal.de/v1alpha1
kind: Machine
metadata:
  namespace: default
  name: my-machine
  labels:
    app: web
spec:
  interfaces:
    - name: my-interface
      ref:
        name: my-machine-interface
  ...
status:
  interfaces:
    - name: my-interface
      primaryIPs: # The machine reports all ips available via its interfaces
        - 192.168.178.1
        - 2607:f0d0:1002:51::4
  ...
```

### The `AliasPrefix` type.

An `AliasPrefix` allows routing a sub-prefix of a network to multiple `NetworkInterface`s. It thus requires a reference
to a `Network` as well as a reference to a `Prefix`.

Example manifest:

```yaml
apiVersion: network.onmetal.de
kind: AliasPrefix
metadata:
  namespace: default
  name: my-alias-prefix
spec:
  networkRef:
    name: my-network
  prefixRef:
    sizes: [ 24, 120 ]
    name: my-prefix
status:
  prefixes:
    - 192.168.0.0/24
    - 2607:f0d0:1002:51::4/120
```

This only allocates the `AliasPrefix`. To establish the binding between `AliasPrefix` and `NetworkInterface`, the
`NetworkInterface` has to reference the desired `AliasPrefix`es:

```yaml
apiVersion: compute.onmetal.de
kind: NetworkInterface
metadata:
  namespace: default
  name: my-machine-interface
spec:
  aliasPrefixes:
    - name: my-alias-prefix
      ref:
        name: my-alias-prefix
...
status:
  aliasPrefixes:
    - name: my-alias-prefix
      prefixes:
        - 192.168.0.0/24
        - 2607:f0d0:1002:51::4/120
...
```

### The `VirtualIP` type

A `VirtualIP` requests a stable public IP for multiple `NetworkInterface`s. We also have a
`type` field that currently only can be `type: Public` in order to support future `VirtualIP` types (most prominently
here `VirtualIP`s in other networks).

As this type manages public IPs, no `prefixRef` can be specified.

Example manifest:

```yaml
apiVersion: compute.onmetal.de
kind: VirtualIP
metadata:
  namespace: default
  name: my-virtual-ip
spec:
  type: Public
status:
  ips:
    - 45.86.152.88
    - 2607:f0d0:1002:51::4
```

Again, to assign such a `VirtualIP` to a `NetworkInterface`, the `NetworkInterface` has to reference it:

```yaml
apiVersion: compute.onmetal.de
kind: NetworkInterface
metadata:
  namespace: default
  name: my-machine-interface
spec:
  virtualIPs:
    - name: my-virtual-ip
      ref:
        name: my-virtual-ip
...
status:
  virtualIPs:
    - name: my-virtualIP
      ips:
        - 45.86.152.88
        - 2607:f0d0:1002:51::4
...
```

## Alternatives

None discussed so far.
