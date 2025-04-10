---
title: Reservation-Based Scheduling

oep-number: 11

creation-date: 2024-09-30

status: draft

authors:

- "@lukasfrank"


reviewers:

- "@afritzler"
- "@balpert89"

---

# OEP-11: Reservation-Based Scheduling

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Alternatives](#alternatives)

## Summary
Scheduling resources is a central process to increase utilization of cloud infrastructure. It not only involves the process to find any feasible `pool` but rather, if possible, to find an optimal compute `pool`.


## Motivation

With `ironcore` a multi-level hierarchy can be built to design availability/failure zones. `Machine`s can be created on every layer, which are subsequently scheduled on the compute `Node`s  in the hierarchy below. A scheduler on every layer should assign a `Machine` on a `Pool` which can fulfill it. This proposal addresses the known limitations of the current implementation:
- API changes needed to extend the scheduling if further resources should be considered
- Available resources are attached to `Pools` and aggregated. In higher levels a correct decision can't be made.

### Goals

- Dynamic (`NetworkInterfaces`, `LocalDiskStorage`) and static resources (resources defined by a `Class` like `CPU`, `Memory`) should be taken into the scheduling decision.
- Scheduling should be robust against data race conditions
- `Machine`s on every level should be scheduled on a `Pool`
- It should be possible to add resources influencing the scheduling decision without API change
- Resource updates of `Pool`s should be possible
- Adding/removing of compute `Pool`s should be possible

### Non-Goals
- The scheduler should not act on `Machine` updates like `NetworkInterface` attachments

## Proposal

For every created `Machine` a resource `Reservation` object will be created. The `poollet` will continue to broker only `Machine`s with a `.spec.machinePoolRef` set. The `Reservation` is a condensed derivation of a `Machine` containing the requested resources like `NetworkInterfaces` and the attached `MachineClass` resource list. The `IRI` will be extended to be able to broker the `Reservation`s. Once the `Reservation` hits a pool provider, the decision can be made if the `Reservation` can be fulfilled or not. In case of accepting the `Reservation`, resources needs to be blocked until a) the reservation is being deleted b) the corresponding `Machine` is being placed on the pool provider. The status of the `Reservation` is being propagated up to the layer where it was created. After some time it contains a list of possible pools where the scheduler can pick one, set the `Machine.spec.machinePoolRef` and the `poollet` will broker the `Machine` to the next level.

`Reservation` resource: 
```
apiVersion: core.ironcore.dev/v1alpha1
kind: Reservation
metadata:
  namespace: default
  name: reservation
spec:
  # Pools define each pool where resources should be reserved
  pools: 
    - poolA
    - poolB
    - poolC
  # Resources defines the requested resources in a ResourceList
  resources:
    localStorage: 10G
    nics: 2
    cpu: 2
status:
  # Pools describe where the resource reservation was possible
  pools: 
    - name: poolA
      rating: 2
      state: accepting
    - name: poolB
      ratring: 1
      state: rejecting
```

Added `IRI` methods:

```protobuf
rpc ListReservations(ListReservationsRequest) returns (ListReservationsResponse) {};
rpc CreateReservation(CreateReservationRequest) returns (CreateReservationResponse) {};
rpc DeleteReservation(DeleteReservationRequest) returns (DeleteReservationResponse) {};
```


### Detailed flow
1. If `.spec.machinePoolRef` is not set, the scheduler creates a `Reservation` that includes the required resources.
2. `poollet`s which match `.spec.pools` of `Reservation` pick it up and broker it one layer down
3. `poriver` evaluates the `Reservation` and sets its state to `accepting` or `rejecting`, which is then propagated up the hierarchy.
4. The scheduler will pick a `pool` on every layer and update the `.spec.machinePoolRef` of the related `Machine`
5. Once a `Machine` reaches a `provider` with a relating `Reservation`, the `Reservation` will be replaced through 
   the `Machine`. Deletion process of `Reservation` is initiated through setting state to `releasing`, which is 
   propagated up.
6. Scheduler will delete top `Reservation` which triggers deletion of all brokered `Reservation`s 



### Advantages
- `Reservation`s can also be used to block resources for a specific use-case
- In case of a partial outage/network issues `Machine`s can be placed on other `Pool`s
- `Pool`s do not leak resource information, owner of resources decides if `Reservation` can be fulfilled (less complexity for over provisioning)

### Disadvantages
- One more resource (`Reservations`) is being introduced

## Alternatives

Another solution: The providers (in the lowest layer) announce there resources to a central cluster. In these clusters a `shallow` pool represent the actual compute pool. In fact the problem of scheduling across multi-levels is transformed in a one-layer scheduling decision.  Pools between the root cluster and the leaf clusters are announced but only to represent the hierarchy and not for the actual scheduling.

If a `Machine` is created, a controller creates a related `Machine` in the central cluster where the `Pool` is being picked. The `Pool` is synced back and the`.spec.machinePoolRef` being set. The `poollet` syncs the `Machine` one layer down. In every subsequent layer the path is being looked up in the central cluster until the provider is hit.

### Advantages
- Scheduling in central place is simpler
- Availability/failure zones can be modeled in central cluster
- (Scheduling decision takes potentially less time?)

### Disadvantages
- Resources are populated to central place and consistency needs to be guaranteed
  - Updates are harder (failure of provider, overbooking)
- Bookkeeping of resources needs to happen twice: in provider and central place 
- Layered structure of hierarchy needs to be duplicated at central place
- If central cluster is not reachable, no `Machine` can be placed