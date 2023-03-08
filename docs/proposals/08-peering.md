---
title: Network Peering

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

# OEP-8: Peering of OnMetal Networks

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Alternatives](#alternatives)

## Summary

Network peering is a crucial feature to allow users to route traffic between two onmetal networks. This OEP introduces a new API that allows the users that created of the networks to express the consent to peering the networks explicitely and extends the existing `Network` API to reflect an active peering.

- peering connects whole networks together
- clear separation from transit gateways
- we need peering requests for both networks to ensure consent on both sides
- peering request should have a ttl so they get removed if not matched

## Motivation

Without peering, users of onmetal would need to run all services via public IPs or via complicated VPN solutions if they are located in separate networks.
We want to implement the basic use case of allowing services running in two distinct onmetal networks to be able to reach eachother without any additional installation from the user and without exposing said services to the public internet.

### Goals

* Define API for requesting and acknowledging peering of two networks
* Peering request should have a TTL that allows for garbage collection of unanswered requests to mitigate flooding of non-matching peering requests
* Peering should result in one network's prefixes should be fully routeable from the other network's prefixes and vice versa, meaning all IP addresses are abailable in all peered networks
* On conflicts (e.g. overlapping IP ranges) both peering requests should enter a failed state

### Non-Goals

We do not want to implement a transit gateway. When we peer networks we expect to make all IP addresses available in the peered networks. With a transit gateway we would be able to only share networks selectively. When peered networks have overlapping IP address ranges we can either fail the peering or have some precedence to routing rules (e.g. local first) but the transit gateway would be able to e.g. define a transit IP range to enable routing between overlapping IP ranges.

## Proposal

We propose using a new API called `NetworkPeering` and extending the `Network` API with a `status` reflecting the peering. The change in the `Network` API will need the enhancement of the `Network`controller to check for matching `NetworkPeering`resources.
For a peering to succeed, both owners of the two `Networks` that should peer, have to create a matching `NetworkPeering`. On successful peering, the `status` of the corresponding `Network` objects should be updated with the other `Network`s details.
A `NetworkPeering` will also have a set time-to-live (ttl), so the system will be able to automatically clean up vacant or unmatched `NetworkPeering`objects. The `status-ttl` field will reflect the end date before the object gets deleted. The actual trigger for garbage collection will be inferred from the `LastTransistionTime` plus a sensible amount of time (e.g. 7 days). Also the garbage collection can only happen if the `NetworkPeering` does NOT have `status.state` set to `Success`.
Always two matching `NetworkPeering` are required to enable peering. Important: `localNetworkRef` must match to other `remoteNetworkRef` and vice versa.

### The `NetworkPeering`type

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NetworkPeering
metadata:
  namespace: my-namespace
  name: network-sample-peering
spec:
  localNetworkRef:
    name: network-a
  remoteNetworkRef:
    name: network-b
    namespace: other-namespace #optional could be another network of the same namespace
status:
  ttl: 2023-02-15T19:00:00Z # omit if empty
  state: -> Pending because missing match
         -> Failed overlapping ip range
         -> Success when found a match and peered
```

* `localNetworkRef`: reference to the network that the creator of this `NetworkPeering` owns and intends to peer, needs to reside in the same namespace as the referenced `Network`object
* `remoteNetworkRef`: reference to the network that should be peered with the one referenced in the `localNetworkRef`, namespace can be omited if this `Network`object resides in the same namespace as the `NetworkPeeringRequest``
* `status.ttl`: ISO timestamp that reflects the point in time when this peering request is being deleted if the `status.state`is still `pending` or `failed` before. If `status.state`is `success`, this will be ignored and the object will not be deleted
* `status.state`: can be either `Pending`, `Success` or `Failed`. State is `Pending` as long as there is no matching `NetworkPeering` present. State is `Success`when there is a matching `NetworkPeering` present. State is `Failed` when certain parameters stop the matching from succeeding, like e.g. overlapping IP ranges of the two referenced networks

### The updated `Network` type

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: my-namespace
  name: network-a
status:
  peeredNetworks:
    - name: network-b
      namespace: other-namespace
```

* `status.peeredNetworks`: is a list of successfully peered `Networks`

### Example

Let's suppose User-1 wants to peer his `Network` `network-1` with User-2's `Network` `network-2`. Let's also assume that User-1's namespace is `namespace-1` and User-2's namespace is called `namespace-2`. With that we get the following two `NetworkPeering`objects reflecting a successful peering of the two `Network` objects presented at the end.

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NetworkPeering
metadata:
  namespace: namespace-1
  name: peer-with-network-2
spec:
  localNetworkRef:
    name: network-1
  remoteNetworkRef:
    name: network-2
    namespace: namespace-2
status:
  ttl: 2023-02-15T19:00:00Z 
  state: success
```

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NetworkPeering
metadata:
  namespace: namespace-2
  name: peer-with-network-1
spec:
  localNetworkRef:
    name: network-2
  remoteNetworkRef:
    name: network-1
    namespace: namespace-1
status:
  ttl: 2023-02-16T20:00:00Z 
  state: success
```

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: namespace-1
  name: network-1
status:
  peeredNetworks:
    - name: network-2
      namespace: namespace-2
```

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: Network
metadata:
  namespace: namespace-2
  name: network-2
status:
  peeredNetworks:
    - name: network-1
      namespace: namespace-1
```

## Alternatives

Users could route all their traffic between services in different `Networks` through the public internet or a dedicated VPN tunnel, e.g. with IPSec