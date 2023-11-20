# IRI - IronCore Runtime Interface

## Introduction

The IronCore Runtime Interface (IRI) is a GRPC-based abstraction layer
introduced to ease the implementation of a `poollet` and `pool provider`. 

A `poollet` does not have any knowledge how the resources are materialized and where the `pool provider` runs.
The responsibility of the `poollet` is to collect and resolve the needed dependencies to materialize a resource.

A `pool provider` implements the IRI, where the IRI defines the correct creation and management of resources 
handled by a `pool provider`. A `pool provider` of the IRI should follow the interface defined in the
[IRI APIs](https://github.com/ironcore-dev/ironcore/tree/main/iri/apis). 

```mermaid
graph LR
    P[poollet] --> IRI
    IRI{IRI} --> B
    B[pool provider]
```

## `pool provider`

A `pool provider` represents a specific implementation of resources managed by
a Pool. The implementation details of the `pool provider` depend on the type of
resource it handles, such as Compute or Storage resources.

Based on the implementation of a `pool provider` it can serve multiple use-cases: 
- to broker resources between different clusters e.g. [volume-broker](https://github.com/ironcore-dev/ironcore/tree/main/broker/volumebroker)
- to materialize resources e.g. block devices created in a Ceph cluster via the [cephlet](https://github.com/ironcore-dev/cephlet)

## Interface Methods

The IRI defines several interface methods categIRIzed into Compute, Storage,
and Bucket.

- [Compute Methods](https://github.com/ironcore-dev/ironcore/tree/main/iri/apis/machine)
- [Storage Methods](https://github.com/ironcore-dev/ironcore/tree/main/iri/apis/volume)
- [Bucket Methods](https://github.com/ironcore-dev/ironcore/tree/main/iri/apis/bucket)

The IRI definition can be extended in the future with new resource groups.

## Diagram

Below is a diagram illustrating the relationship between `poollets`,
IRI, and `pool providers` in the `ironcore` project.

```mermaid
graph TB
    A[Machine] -- scheduled on --> B[MachinePool]
    C[Volume] -- scheduled on --> D[VolumePool]
    B -- announced by --> E[machinepoollet]
    D -- announced by --> F[volumepoollet]
    E -- GRPC calls --> G[IRI compute provider]
    F -- GRPC calls --> H[IRI storage provider]
    G -.sidecar to.- E
    H -.sidecar to.- F
```

This diagram illustrates:

- `Machine` resources are scheduled on a `MachinePool` which is announced by the `machinepoollet`.
- Similarly, `Volume` resources are scheduled on a `VolumePool` which is announced by the `volumepoollet`.
- The `machinepoollet` and `volumepoollet` each have an IRI `provider` sidecar, which provides a GRPC interface for 
making calls to create, update, or delete resources.
- The IRI `provider` (Compute) is a sidecar to the `machinepoollet` and the IRI `provider` (Storage) is a sidecar to the 
`volumepoollet`. They handle GRPC calls from their respective `poollets` and interact with the actual resources.
