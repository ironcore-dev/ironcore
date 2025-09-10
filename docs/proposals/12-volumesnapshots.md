---
title: Introducing VolumeSnapshot type and Restoration Paths

iep-number: 12

creation-date: 2025-07-23

status: implementable

authors:

- "@afritzler"

reviewers:

- "@afritzler"
- "@lukasfrank"
- "@balpert89"

---

# IEP-12: Introducing `VolumeSnapshot` type and Restoration Paths

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [API Design](#api-design)
  - [`VolumeSnapshot` Resource](#volumesnapshot-resource)
  - [`Volume` Resource Enhancement](#volume-resource-enhancement)
- [`VolumeRuntime` enhancement](#volumeruntime-enhancement)
- [Implementation Strategy](#implementation-strategy)
- [Alternatives](#alternatives)

## Summary

We propose the introduction of new resource to the IronCore project: `VolumeSnapshot`. The `VolumeSnapshot` 
resource would allow users to take point-in-time snapshots of the content of a `Volume`. We also propose an 
enhancement to the `Volume` resource to provide a path for restoring from a `VolumeSnapshot`. These new resources 
and changes would provide users with increased capabilities for data protection and disaster recovery.

## Motivation

The IronCore project currently provides interfaces for managing resources such as `Volumes`, `VolumeClasses` and `VolumePools`. 
A `VolumePool` is a representation of a `VolumePool` provider that materializes the requested `Volume` in an underlying
storage implementation. A `VolumePool` provider here announces the supported `VolumeClasses` to indicate to the user
which performance characteristics of `Volumes` this particular `VolumePool` can offer.
There is currently no direct way to snapshot the content of a `Volume`, nor is there a method 
for defining different types of snapshots or restoring from a snapshot, which potentially puts user's data at risk.

### Goals

- Allow the user to take point-in-time snapshots of the content of a `Volume`.
- Allow the user to restore a `Volume` from a snapshot.
- Hide provider-specific details of snapshot management from the user.
- For encrypted `Volumes`, reuse the existing encryption key of the `Volume` when restoring from a snapshot.

### Non-Goals

- Incremental snapshots of `Volumes`.
- Restore an existing `Volume` from a snapshot.

## API Design

### `VolumeSnapshot` Resource

The `VolumeSnapshot` resource will have the following structure:

```yaml
apiVersion: storage.ironcore.dev/v1alpha1
kind: VolumeSnapshot
metadata:
  name: example-snapshot
  namespace: default
spec:
  volumeRef:
    name: example-volume
status:
  snapshotID: 029321790472c133de32f05944bc2543629fe4b4264dc42f8d95a87adbba265
  state: Pending/Ready/Failed
  size: 10Gi
  lastStateTransitionTime: 2025-08-20T08:24:25Z
  
```  

#### Fields:

- `volumeRef`: Reference to the `Volume` to be snapshot. The `VolumeSnapshot` and `Volume` should be in the same namespace.
- `snapshotID`: Reference to the storage-provider specific snapshot object.
- `status.state`: The current phase of the snapshot. It can be `Pending`, `Ready` or `Failed`.
  - `Pending`: The snapshot resource has been created, but the snapshot has not yet been initiated.
  - `Ready`: The snapshot has been successfully created and is ready for use.
  - `Failed`: The snapshot creation has failed.
- `status.size`: The size of the data in the snapshot. This will be populated by the system once the snapshot is ready.

### Volume Resource Enhancement

The `Volume` resource will have 2 new volume data source inline fields as `volumeSnapshotRef` and `osImage` under `spec`:

```yaml
apiVersion: storage.ironcore.dev/v1alpha1
kind: Volume
metadata:
  name: restored-volume
  namespace: default
spec:
  volumeClassRef:
    name: example-volumeclass
  volumeSnapshotRef:
    name: example-snapshot
  osImage: test-image
```

#### Fields:

- `volumeSnapshotRef`: Indicates to use the specified `VolumeSnapshot` as the data source.
- `osImage`: It is an os image to bootstrap the volume.

## `VolumeRuntime` enhancement

To support the creation and restoration of `VolumeSnapshot` resources, the `VolumeRuntime` interface will be 
enhanced to include methods for:

- Creating a `VolumeSnapshot` from a `Volume`.
- Deleting a `VolumeSnapshot`.
- Restoring a `Volume` from a `VolumeSnapshot`.

```protobuf
service VolumeRuntime {
  // CreateVolumeRequest needs to be updated to include a dataSource field to support restoration from a snapshot 
  // when creating a new Volume.
  rpc CreateVolume(CreateVolumeRequest) returns (CreateVolumeResponse) {};

  // CreateVolumeSnapshot will be used to create a snapshot of a Volume.
  rpc CreateVolumeSnapshot(CreateVolumeSnapshotRequest) returns (CreateVolumeSnapshotResponse) {};
  // DeleteVolumeSnapshot will be used to delete a snapshot and its associated content.
  rpc DeleteVolumeSnapshot(DeleteVolumeSnapshotRequest) returns (DeleteVolumeSnapshotResponse) {};
  // ListVolumeSnapshots will be used to list all snapshots managed by the volume provider
  rpc ListVolumeSnapshots(ListVolumeSnapshotsRequest) returns (ListVolumeSnapshotsResponse) {};
}
```

### Implementation Strategy

Once the proposal is approved, the implementation will follow these stages:

1. Define the `VolumeSnapshot` resource in the `ironcore` project.
2. Extend the `VolumeRuntime` interface to support snapshot creation and restoration.
3. Implement the runtime interface methods in the `volume-broker` component
4. Implement the runtime interface methods in the respective storage provider components.
5. Implement the `volumepoollet` to react on the new API types and handle the creation, deletion, and restoration of 
snapshots by invoking the appropriate methods in the `VolumeRuntime` interface.

## Alternatives

Potential alternatives to the proposed changes could include:

1. **Third-Party Snapshot Tools**: One alternative to implementing a `VolumeSnapshot` resource could be 
to recommend users leverage existing third-party tools or cloud-provider services for managing snapshots. However, 
this approach might lack the integration and ease of use that come with an inbuilt solution.
2. **Application-Level Snapshots**: Another alternative could be to allow applications to manage their own snapshots
and backups on an application level.
3. **Extend `Volume` Resource**: Rather than introducing new resources, we could extend the current `Volume` resource 
with snapshot-related functionalities. This could include adding fields to the `Volume` spec that would trigger a 
snapshot when certain conditions are met. However, this could potentially complicate the `Volume` resource and would 
limit the flexibility to manage snapshots independently.
4. **Storage Provider Level Snapshots**: Leave the task of snapshot creation and management to individual storage 
providers. Each storage provider has its own way of managing snapshots, and they could provide their own APIs for 
managing them. This might complicate things for users who have to work with different APIs for different 
storage providers.

Please note that these alternatives come with their own sets of advantages and disadvantages, and a thorough evaluation 
should be conducted to select the most appropriate solution.
