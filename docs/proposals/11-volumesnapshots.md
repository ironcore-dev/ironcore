---
title: Introducing VolumeSnapshot and VolumeSnapshotContent types and Restoration Paths

iep-number: 11

creation-date: 2025-07-23

status: implementable

authors:

- "@afritzler"

reviewers:

- "@afritzler"
- "@lukasfrank"
- "@balpert89"

---

# IEP-11: Introducing `VolumeSnapshot` and `VolumeSnapshotContent` types and Restoration Paths

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [API Design](#api-design)
  - [`VolumeSnapshot` Resource](#volumesnapshot-resource)
  - [`VolumeSnapshotContent` Resource](#volumesnapshotcontent-resource)
  - [`Volume` Resource Enhancement](#volume-resource-enhancement)
- [`VolumeRuntime` enhancement](#volumeruntime-enhancement)
- [Implementation Strategy](#implementation-strategy)
- [Alternatives](#alternatives)

## Summary

We propose the introduction of two new resources to the IronCore project: `VolumeSnapshot` and `VolumeSnapshotContent`. 
The `VolumeSnapshot` resource would allow users to take point-in-time snapshots of the content of a `Volume`. 
The `VolumeSnapshotContent` represents the actual storage-provider-specific snapshot content. We also propose an 
enhancement to the `Volume` resource to provide a path for restoring from a snapshot. These new resources and 
changes would provide users with increased capabilities for data protection and disaster recovery.

## Motivation

The IronCore project currently provides interfaces for managing resources such as `Volumes`, `VolumeClasses` and `VolumePools`. 
A `VolumePool` is a representation of a `VolumePool` provider that materializes the requested `Volume` in an underlying
storage implementation. A `VolumePool` provider here announces the supported `VolumeClasses` to indicate to the user
which performance characteristics of `Volumes` this particular `VolumePool` can offer.
There is currently no direct way to snapshot the content of a `Volume`, nor is there a method 
for defining different types of snapshots or restoring from a snapshot, which potentially puts users' data at risk.

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
  phase: Pending/Processing/Ready/Failed/Deleting
  restoreSize: 10Gi
```  

#### Fields:

- `volumeRef`: Reference to the `Volume` to be snapshot. The `VolumeSnapshot` and `Volume` should be in the same namespace.
- `status.phase`: The current phase of the snapshot. It can be `Pending`, `Processing`, `Ready`, `Failed`, or `Deleting`.
- `status.restoreSize`: The size of the data in the snapshot. This will be populated by the system once the snapshot is ready.

### `VolumeSnapshotContent` Resource

The `VolumeSnapshotContent` resource will have the following structure:

```yaml
apiVersion: storage.ironcore.dev/v1alpha1
kind: VolumeSnapshotContent
metadata:
  name: example-volumesnapshotcontent
spec:
  deletionPolicy: Delete
  source:
    snapshotHandle: 1334353-234234-45435435 # Unique identifier for the snapshot in the storage provider
  volumeSnapshotRef:
    name: example-snapshot
    namespace: default
    uid: 12345678-1234-5678-1234-123456789012
```
    
#### Fields:

- `deletionPolicy`: Defines what happens to the snapshot content when the `VolumeSnapshot` is deleted. It can be `Delete` or `Retain`.
- `source`: Contains the `snapshotHandle`, which is a unique identifier for the snapshot in the storage provider.
- `volumeSnapshotRef`: Reference to the `VolumeSnapshot` that this content belongs to. It includes the name, namespace, and UID of the `VolumeSnapshot`.

### Volume Resource Enhancement

The `Volume` resource will have a new `dataSource` field under `spec`:

```yaml
apiVersion: storage.ironcore.dev/v1alpha1
kind: Volume
metadata:
  name: restored-volume
spec:
  volumeClassRef:
    name: example-volumeclass
  dataSource:
    apiGroup: storage.ironcore.dev/v1alpha1
    kind: VolumeSnapshot
    name: example-snapshot
```

#### Fields:

`dataSource`: Information regarding the snapshot to be used as the source for the restoration. Contains:
- `apiGroup`: The group to which the referenced resource belongs, which would be `storage.ironcore.dev/v1alpha1`.
- `kind`: Kind of the source, which should be `VolumeSnapshot`.
- `name`: Name of the snapshot.

## `VolumeRuntime` enhancement

To support the creation and restoration of `VolumeSnapshot` resources, the `VolumeRuntime` interface will be 
enhanced to include methods for:

- Creating a `VolumeSnapshot` from a `Volume`.
- Deleting a `VolumeSnapshot` and its associated `VolumeSnapshotContent`.
- Restoring a `Volume` from a `VolumeSnapshot`.

```protobuf
service VolumeRuntime {
  // CreateVolumeRequest needs to be updated to include a dataSource field to support restoration from a snapshot 
  // when creating a new Volume.
  rpc CreateVolume(CreateVolumeRequest) returns (CreateVolumeResponse) {};

  // CreateSnapshot will be used to create a snapshot of a Volume.
  rpc CreateSnapshot(CreateSnapshotRequest) returns (CreateSnapshotResponse) {};
  // DeleteSnapshot will be used to delete a snapshot and its associated content.
  rpc DeleteSnapshot(DeleteSnapshotRequest) returns (DeleteSnapshotResponse) {};
  // ListSnapshots will be used to list all snapshots managed by the volume provider
  rpc ListSnapshots(ListSnapshotsRequest) returns (ListSnapshotsResponse) {};
}
```

### Implementation Strategy

Once the proposal is approved, the implementation will follow these stages:

1. Define the `VolumeSnapshot` and `VolumeSnapshotContent` resources in the `ironcore` project.
2. Extend the `VolumeRuntime` interface to support snapshot creation and restoration.
3. Implement the runtime interface methods in the `volume-broker` component
4. Implement the runtime interface methods in the respective storage provider components.
5. Implement the `volumepoollet` to react on the new API types and handle the creation, deletion, and restoration of 
snapshots by invoking the appropriate methods in the `VolumeRuntime` interface.

## Alternatives

Potential alternatives to the proposed changes could include:

1. **Third-Party Snapshot Tools**: One alternative to implementing a `VolumeSnapshot` and `VolumeSnapshotContent` resource could be 
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
