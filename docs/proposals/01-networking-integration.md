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

Without networking, any machine / process running inside a datacenter cannot interact / affect the outside
world. Networking is a crucial component that has to be implemented for onmetal to have business value.
In a full-fledged state, networking also enables security to the outside world and within a datacenter itself.

The basic use case we want to implement with onmetal is a machine that can access the internet and can be
reached from the internet.

### Goals

* Define APIs for managing isolated networks.
* Define APIs for managing splitting the isolated networks (subnetting).
* Define APIs for assigning / routing public IPs / prefixes to members of a network / subnet.
* Adapt the `compute.Machine` type to integrate with the network API.
* It should be possible to extend the API in the future to achieve the following (listed by decreasing priority):
  * Communication within a subnet (plus security concepts)
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

The `Network` type defines a private `Network`. Private meaning that it is isolated from the public
internet and communication within it can be freely managed. Traffic can be allowed or disallowed.
For defining a network, a new namespaced `Network` type will be introduced.
A `Network` has a `prefix` that define its boundaries.

Once a `Network` is up and ready, the prefixes available for allocation are reported in its `status`
as well as a condition indicating its availability / health.

Example manifest:

```yaml
apiVersion: network.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: default
  name: my-network
spec:
  prefix: 192.0.0.0/8
status:
  available:
  - 192.0.0.0/8
  conditions:
  - type: Available # Available may be the name for this condition, though this has to be refined.
    status: True
    reason: PrefixAllocated
```

### `Subnet` type

A `Network` can be sliced into multiple subparts by subnetting. For this, the `Subnet` type is introduced
that specifies the managed sub-prefix of a `Network`. The sub-prefix of can be specified explicitly by the user
or dynamically computed by specifying a desired prefix size.
When the prefix is dynamically computed, upon successful allocation, `Subnet.spec.prefix` is updated to the
computed prefix. Once set, the value is immutable.

As for the `Network`, a `Subnet` also reports both whether it is valid / available and the remaining address space
it has.

Example manifest:

```yaml
apiVersion: network.onmetal.de/v1alpha1
kind: Subnet
metadata
  namespace: default
  name: my-subnet
spec:
  networkRef:
    name: my-network
  prefixSize: 16
  # prefix: 192.1.0.0/16 # This will be set once successfully allocated.
status:
  available:
  - 192.1.0.0/16
  conditions:
  - type: Available # Available may be the name for this condition, though this has to be refined.
    status: True
    reason: PrefixAllocated
```

### `NetworkInterface` type

A `NetworkInterface` lets a `Machine` join a `Subnet`. For now, only **1** network interface per
`Machine` will be allowed. In contrast to the current state (`Machine` specifying multiple network interfaces
in its spec), a `NetworkInterface` is a separate, dedicated type that has to be referenced by
a `Machine`.

A subnet will allocate an ip for itself and report once it's ready.

Example manifest:

```yaml
apiVersion: compute.onmetal.de/v1alpha1
kind: NetworkInterface
metadata:
  name: my-nic
spec:
  subnetRef:
    name: my-subnet
status:
  prefixes: # We use prefix notation by default.
    - 192.168.178.1/32
  conditions:
  - type: Available # Available may be the name for this condition, though this has to be refined.
    status: true
    reason: PrefixAllocated
```

The `Machine` type will be modified to reference the `NetworkInterface`:

```yaml
apiVersion: compute.onmetal.de/v1alpha1
kind: Machine
metadata:
  namespace: default
  name: my-machine
  labels:
    app: web
spec:
  ...
  interfaces:
    - name: my-inteface
      ref:
        name: my-nic
  ...
status:
  prefixes: # The machine reports all prefixes available via its interfaces
    - 192.168.178.1/32
  ...
```

### The `PublicPrefix` type

The `PublicPrefix` type controls how to allocate a public prefix for a `NetworkInterface` / a `Machine`.
When a `PublicPrefix` is created, it allocates a stable public prefix from a provider-owned pool of
public prefixes. Successfully allocated prefixes are reported in the `status`.

Example manifest:

```yaml
apiVersion: compute.onmetal.de/v1alpha1
kind: PublicPrefix
metadata:
  namespace: default
  name: my-public-prefix
spec:
  selector:
    kind: Machine
    matchLabels:
      app: web
status:
  prefixes:
    - 13.14.15.1/32
```

## Alternatives

None discussed so far.
