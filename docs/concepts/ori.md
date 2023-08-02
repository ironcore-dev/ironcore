# ORI - Onmetal Runtime Interface

## Introduction

The Onmetal Runtime Interface (ORI) is a GRPC-based abstraction layer
introduced to ease the implementation of a `poollet` and `pool provider`. 

A `poollet` does not have any knowledge how the resources are materialized and where the `pool provider` runs.
The responsibility of the `poollet` is to collect and resolve the needed dependencies to materialize a resource.

A `pool provider` implements the ORI, where the ORI defines the correct creation and management of resources 
handled by a `pool provider`. A `pool provider` of the ORI should follow the interface defined in the
[ORI APIs](https://github.com/onmetal/onmetal-api/tree/main/ori/apis). 

```mermaid
graph LR
    P[poollet] --> ORI
    ORI{ORI} --> B
    B[pool provider]
```

## `pool provider`

A `pool provider` represents a specific implementation of resources managed by
a Pool. The implementation details of the `pool provider` depend on the type of
resource it handles, such as Compute or Storage resources.

Based on the implementation of a `pool provider` it can serve multiple use-cases: 
- to broker resources between different clusters e.g. [volume-broker](https://github.com/onmetal/onmetal-api/tree/main/broker/volumebroker)
- to materialize resources e.g. block devices created in a Ceph cluster via the [cephlet](https://github.com/onmetal/cephlet)

## Interface Methods

The ORI defines several interface methods categorized into Compute, Storage,
and Bucket.

- [Compute Methods](https://github.com/onmetal/onmetal-api/tree/main/ori/apis/machine)
- [Storage Methods](https://github.com/onmetal/onmetal-api/tree/main/ori/apis/volume)
- [Bucket Methods](https://github.com/onmetal/onmetal-api/tree/main/ori/apis/bucket)

The ORI definition can be extended in the future with new resource groups.

## Diagram

Below is a diagram illustrating the relationship between `poollets`,
ORI, and `pool providers` in the `onmetal-api` project.

```mermaid
graph TB
    A[Machine] -- scheduled on --> B[MachinePool]
    C[Volume] -- scheduled on --> D[VolumePool]
    B -- announced by --> E[machinepoollet]
    D -- announced by --> F[volumepoollet]
    E -- GRPC calls --> G[ORI compute provider]
    F -- GRPC calls --> H[ORI storage provider]
    G -.sidecar to.- E
    H -.sidecar to.- F
```

This diagram illustrates:

- `Machine` resources are scheduled on a `MachinePool` which is announced by the `machinepoollet`.
- Similarly, `Volume` resources are scheduled on a `VolumePool` which is announced by the `volumepoollet`.
- The `machinepoollet` and `volumepoollet` each have an ORI `provider` sidecar, which provides a GRPC interface for 
making calls to create, update, or delete resources.
- The ORI `provider` (Compute) is a sidecar to the `machinepoollet` and the ORI `provider` (Storage) is a sidecar to the 
`volumepoollet`. They handle GRPC calls from their respective `poollets` and interact with the actual resources.
