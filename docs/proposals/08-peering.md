---
title: VNet Peering

oep-number: 8

creation-date: 2023-02-24

status: implementable

authors:

- "@ManuStoessel"
- "@gehoern"

reviewers:

- "@afritzler"
- "@adracus"

---

# OEP-NNNN: Your short, descriptive title

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Alternatives](#alternatives)

## Summary

Network peering is a crucial feature to allow users to route traffic between two onmetal networks. This OEP introduces a new API that allows the owner of both networks to express the consent to peering the networks explicitely and extends the existing `Network` API to reflect an active peering.

- peering connects whole networks together 
(vf get both network route set announced)
- clear separation from transit gateways
- we need peering requests on both partners
- peering request should have a ttl so they get removed if not matched

## Motivation

Without peering, users of onmetal would need to run all services via public IPs or via complicated VPN solutions if they are located in seperate networks.
We want to implement the basic use case of allowing services running in two distinct onmetal networks to be able to reach eachother without any additional installation from the user and without exposing said services to the public internet.

### Goals

* Define API for requesting and acknowledging peering of two network
* Peering request should have a TTL that allows for garbage collection of unanswered requests
* Peering should result in one network's prefixes should be fully routeable from the other network's prefixes and vice versa (vf get both network route sets announced)

### Non-Goals

* We do not want to implement a transit gateway

## Proposal

We propose using a new API called `NetworkPeeringRequests` and extending the `Network` API with a `status` reflecting the peering. For a peering to succeed, both owners of the two `Networks` that should peer, have to create a matching `NetworkPeeringRequest`. On successful peering, the `status` of the corresponding `Network` objects should be updated with the other `Network`s details. A `NetworkPeeringRequest` will also have a set time-to-live (ttl), so the system will be able to automatically clean up vacant or unmatched `NetworkPeeringRequest`objects. Always two matching `NetworkPeeringRequests` are required to enable peering. Important: `localNetworkRef` must match to other `remoteNetworkRef` and vice versa.

### The `NetworkPeeringRequest`type

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NetworkPeeringRequest
metadata:
  namespace: customer-manuel
  name: network-sample-request
spec:
  localNetworkRef:
    name: manuels-network
  remoteNetworkRef:
    name: mysupernetwork
    namespace: customer-andre #optional could be another network of the same namespace
status:
  ttl: 2023-02-15T19:00:00Z # omit if empty
  state: -> pending because missing match
         -> failed overlapping ip range
         -> success when found a match and peered
```

* `localNetworkRef`: reference to the network that the creator of this `NetworkPeeringRequest` owns and intends to peer, needs to reside in the same namespace as the referenced `Network`object
* `remoteNetworkRef`: reference to the network that should be peered with the one referenced in the `localNetworkRef`, namespace can be omited if this `Network`object resides in the same namespace as the `NetworkPeeringRequest``
* `status.ttl`: ISO timestamp that reflects the point in time when this peering request is being deleted if the `status.state`is still `pending` or `failed` before. If `status.state`is `success`, this will be ignored and the object will not be deleted
* `status.state`: can be either `pending`, `success` or `failed`. State is `pending` as long as there is no matching `NetworkPeeringRequest` present. State is `success`when there is a matching `NetworkPeeringRequest` present. State is `failed` when certain parameters stop the matching from succeeding, like e.g. overlapping IP ranges of the two referenced networks

### The updated `Network` type

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: customer-manuel
  name: manuels-network
status:
  peeredNetworks:
    - name: mysupernetwork
      namespace: customer-andre
```

* `status.peeredNetworks`: is a list of successfully peered `Networks`

### Example

Let's suppose Manuel wants to peer his `Network` `manuels-network` with André's `Network` `mysupernetwork`. Let's also assume that Manuel's namespace is `customer-manuel` and André's namespace is called `customer-andre`. With that we get the following two `NetworkPeeringRequest`objects reflecting a successful peering of the two `Network`objects presented at the end.

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NetworkPeeringRequest
metadata:
  namespace: customer-manuel
  name: peer-with-andre
spec:
  localNetworkRef:
    name: manuels-network
  remoteNetworkRef:
    name: mysupernetwork
    namespace: customer-andre
status:
  ttl: 2023-02-15T19:00:00Z 
  state: success
```

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NetworkPeeringRequest
metadata:
  namespace: customer-andre
  name: peer-with-manuel
spec:
  localNetworkRef:
    name: mysupernetwork
  remoteNetworkRef:
    name: manuels-network
    namespace: customer-manuel
status:
  ttl: 2023-02-16T20:00:00Z 
  state: success
```

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: customer-manuel
  name: manuels-network
status:
  peeredNetworks:
    - name: mysupernetwork
      namespace: customer-andre
```

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: customer-andre
  name: mysupernetwork
status:
  peeredNetworks:
    - name: manuels-network
      namespace: customer-manuel
```

## Alternatives

Users could route all their traffic between services in different `Networks` through the public internet or a dedicated VPN tunnel, e.g. with IPSec