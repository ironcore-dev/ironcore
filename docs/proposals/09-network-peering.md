---
title: Network Peering

iep-number: 9

creation-date: 2023-03-17

status: implementable

authors:

- "@adracus"

reviewers:

- "@afritzler"
- "@gehoern"

---

# IEP-9: Network Peering

## Table of Contents

- [IEP-9: Network Peering](#IEP-9-network-peering)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
  - [Alternatives](#alternatives)

## Summary

Network peering is a technique used to interleave two isolated networks, allowing
members of both networks to communicate with each other as if they were in the same
networking domain. This proposal describes how to introduce network peering to
`ironcore`, building upon the existing concepts that were proposed in the
[Networking Integration OEP](01-networking-integration.md).

## Motivation

Network peering allows members of two networks to communicate with each other
without exposing them publicly and without a single point of failure. The
networking fabric underneath is used to enable the actual routing. `ironcore`'s
`networking` API should offer a way to define such a peering.

### Goals

* Allow two `Network`s to communicate with each other without exposing their
  routes / addresses publicly and by only using network routing to do so
  (i.e. no load-balancing).
* Be less resource-intensive than comparable solutions like load-balancing /
  VPN overlays.

### Non-Goals

* Map / transform IP addresses: The peered networks will be interleaved with
  each other without any transformation. The owners of the networks are
  responsible for keeping the addresses conflict-free. If there are conflicts,
  it is up to the networking implementation how to resolve them (e.g. use
  a 'local-first' approach for which address / route to use).

## Proposal

Extend the `networking.ironcore.dev.Network` resource with a `spec.peerings`
field that specifies the desired network peerings and a `status.peerings` that
reflects the status of these peerings.

A peering has a `name` as handle & primary key (used in `StrategicMergePatch` and
`Apply`). It references the network to peer with via a `networkRef` field. This
`networkRef` contains the `name` and the `uid` of the target network. If the `uid`
is unset, the `Network` controller sets this to the `uid` of the corresponding
network upon first reconciliation. This ensures that the same object instances
are peered together by verifying the object identity.

Both `Network`s have to specify a matching peering item (i.e. reference each
other via `networkRef`) to mutually accept the peering.

The (binding) `phase` of a `spec.peerings` item is reflected in a corresponding
`status.peerings` item with the same `name`. The `phase` can either be `Pending`,
meaning there is no active peering, or `Bound`, meaning the peering as described
in the `spec.peerings` item is in place. The `lastTransitionTime` field is updated
every time there is a change in the `phase`, allowing users and external
controllers to determine whether a binding is hanging and to manually delete it
if necessary.

Example Manifests:

```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: Network
metadata:
  name: my-network-1
  uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
spec:
  peerings:
  # name is the name of the peering configuration
  - name: my-network-peering
    networkRef:
      name: my-network-2
      # If unset, uid will be filled in by the Network controller
      # uid: 3030dcf9-f031-801b-f0f0-4fec2016e73a
status:
  peerings:
  - name: my-network-peering
    # The phase shows the binding progress between two networks.
    # The initial state is 'Pending' until both peers accept.
    phase: Bound
    lastPhaseTransitionTime: "2023-02-16T15:06:58Z"
---
apiVersion: networking.ironcore.dev/v1alpha1
kind: Network
metadata:
  name: my-network-2
  uid: 3030dcf9-f031-801b-f0f0-4fec2016e73a
spec:
  # Both networks have to have the semantically same peering configuration.
  peerings:
  - name: my-network-peering
    networkRef:
      name: my-network-1
      # If unset, uid will be filled in by the Network controller
      # uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
status:
  peerings:
  - name: my-network-peering
    phase: Bound
    lastPhaseTransitionTime: "2023-02-16T15:06:58Z"
```

Network peering can also specify `Network`s from different namespaces.
For this, the `networkRef` field includes a `namespace` field (default empty if in
the same namespace). Both `Network`s need to reference the other network and
namespace correctly, otherwise the peering will stay in `phase: Pending` indefinitely.

Example Manifests:

```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: Network
metadata:
  namespace: ns-1
  name: my-network-1
spec:
  peerings:
  - name: my-network-peering
    networkRef:
      namespace: ns-2
      name: my-network-2
---
apiVersion: networking.ironcore.dev/v1alpha1
kind: Network
metadata:
  namespace: ns-2
  name: my-network-2
spec:
  peerings:
  - name: my-network-peering
    networkRef:
      namespace: ns-1
      name: my-network-1
```

## Alternatives

* Create a VPN overlay between two networks. However, this requires a (potentially
  public endpoint and introduces points of failure in form of the VPN server(s)
  and client(s). Additionally, the VPN components have to be maintained manually.
* Use hole-punching to create a bidirectional tunnel. This cannot always be done,
  as it depends on the network fabric, requires a publicly available
  rendezvous-point, introduces potential points of failure and requires
  maintenance for its components as for the VPN-based solution.
* Depending on the use-case, services can be exposed behind an internal
  load-balancer, providing failure-safe and scalable communication channels.
