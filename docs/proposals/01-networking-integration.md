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

Networking is part of the MVP for the `onmetal` stack. From a user's perspective, it should be possible to

* Have `Machine`s communicate with each other.
* Have `Machine`s communicate with the internet.
* Have the internet communicate with a `Machine` (exposure).
* Orchestrate the traffic between the machines (i.e. security groups / policies).

### Goals

* Augment the `Machine` type with networking-relevant definitions.
* Define types for managing networks.
* Define types for managing public prefixes.
* Extension points for security concepts to be implemented in the future

### Non-Goals

* Define Load Balancer APIs (L4 upcoming soon after this draft, L7 only later)

* Define security policies for
  
  * Internet connections
  
  * Machine-to-machine connections
  
  * Machine-to-other-network connections

* Define multi-region interconnection

* Define multi-network interconnection (upcoming, but not in initial scope)

* Define multi-namespace interconnection (upcoming, but not in initial scope)

* Allow user-owned public IP pools

* Feature-creep beyond a simplistic MVP

## Proposal

By default, we assume machines can reach the internet. We will limit this via policies in the future. For the first iteration however, this is out of scope.

As part of machine isolation, users should be able to define networks. Machines can be part of a network by having a network interface & internal prefixes allocated in it.

A machine can be exposed to the public internet by associating a public prefix with an internal target (for this scope, we choose a machine as target).

Let's break this down:

### `Network` type

For defining a network, a new namespaced `Network` type will be introduced. A `Network` has a `prefix` of which members can allocate their own prefixes / IPs.

Example manifest:

```yaml
apiVersion: network.onmetal.de/v1alpha1
kind: Network
metadata:
  name: my-network
spec:
  prefix: 192.168.178.0/24
```

### `NetworkInterface` type

The network interface is the entrypoint for a `Machine` into a `Network`. For the first iteration, we will only allow
**1** network interface per machine. The `NetworkInterface` will be a separate object. In the future, we may allow templating out network interfaces in the `Machine` object directly.

Example manifest:

```yaml
apiVersion: compute.onmetal.de/v1alpha1
kind: NetworkInterface
metadata:
  name: my-nic
spec:
  targetRef:
    kind: Network
    name: my-network
status:
  prefixes: # We use prefix notation by default.
    - 192.168.178.1/32
```

The `Machine` type will be adapted to reference the `NetworkInterface`:

```yaml
apiVersion: compute.onmetal.de/v1alpha1
kind: Machine
metadata:
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
```

### The `PublicPrefix` type

A `PublicPrefix` controls how to allocate a public prefix for internal prefixes / network interfaces / machines.

As soon as a `PublicPrefix` is created and selects a `Machine` / `NetworkInterface`, the underlying routing has to be adapted to map the public prefix to the internal prefix of the status of the `Machine` / `NetworkInterface`.

Example manifest:

```yaml
apiVersion: compute.onmetal.de/v1alpha1
kind: PublicPrefix
metadata:
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