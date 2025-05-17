---
title: Introducing Snapshot, SnapshotClass, and Restoration Paths

oep-number: 11

creation-date: 2023-07-24

status: implementable

authors:

- "@afritzler"

reviewers:

- "@afritzler"
- "@lukasfrank"

---

# OEP-11: Introducing `Snapshot`, `SnapshotClass`, and Restoration Paths

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Alternatives](#alternatives)

- [Overview](#overview)
- [Background](#background)
- [Goals](#goals)
- [Non-Goals](#non-goals)
- [API Design](#api-design)
  - [`Snapshot` Resource](#snapshot-resource)
  - [`SnapshotClass` Resource](#snapshotclass-resource)
  - [`Volume` Resource Enhancement](#volume-resource-enhancement)
- [Implementation Strategy](#implementation-strategy)
- [Potential Alternatives](#potential-alternatives)

## Overview

We propose the introduction of two new resources to the onmetal-api project: `Snapshot` and `SnapshotClass`. 
The `Snapshot` resource would allow users to take point-in-time snapshots of the content of a `Volume`. 
The `SnapshotClass` would define the different snapshot types supported by a `VolumePool`. We also propose an 
enhancement to the `Volume` resource to provide a path for restoring from a snapshot. These new resources and 
changes would provide users with increased capabilities for data protection and disaster recovery.

## Background

The `onmetal-api` project currently provides interfaces for managing resources such as `Volumes` and `VolumePools`. 
A `VolumePool` is a representation of a `VolumePool` provider that materializes the requested `Volume` in an underlying
storage implementation. However, there is no direct way to snapshot the content of a `Volume`, nor is there a method 
for defining different types of snapshots or restoring from a snapshot, which potentially puts users' data at risk.

## Goals

1. **Introduce `Snapshot` Resource**: Enable users to take point-in-time snapshots of the content of a `Volume`.
2. **Introduce `SnapshotClass` Resource**: Define the different types of snapshots that a `VolumePool` can support.
3. **Implement Restoration Path**: Enhance the `Volume` resource to include a `dataSource` field, providing a path for 
restoring data from a `Snapshot`.
4. **Improve Data Protection**: By creating snapshots and allowing data restoration, users can protect their data and 
restore it in case of loss or corruption.
5. **Support for Different Types of Snapshots**: With the introduction of `SnapshotClass`, support for different types 
of snapshots based on `VolumePool` capabilities will be provided.
6. **Enhance Flexibility**: Provide users with the ability to choose the type of snapshot they want to use based on 
their specific needs.

## Non-Goals

1. **Modify Existing `VolumePool` Resource**: This proposal aims to introduce new resources and an enhancement to 
the `Volume` resource and doesn't intend to modify the behavior of the existing `VolumePool` resources.
3. **Provide Backup or Archive Functionality**: The `Snapshot` resource will facilitate data protection through 
snapshots, but it's not designed to provide a complete backup or archival solution for `Volumes`.
4. **Ensure Immediate Availability of Snapshot**: While snapshots aim to provide point-in-time data protection, 
the actual creation and readiness of a snapshot can depend on various factors such as `Volume` size, 
`VolumePool` capabilities, and infrastructure performance.
5. **Replace Existing Disaster Recovery Solutions**: The `Snapshot` resource and restoration path are meant to be 
additional tools in the data protection toolbox, not to replace existing comprehensive disaster recovery and 
data protection solutions.

## API Design

### `Snapshot` Resource

The `Snapshot` resource will have the following structure:

```yaml
apiVersion: storage.api.onmetal.de/v1alpha1
kind: Snapshot
metadata:
  name: example-snapshot
spec:
  volumeRef:
    name: example-volume
  snapshotClassRef:
    name: example-snapshotclass
status:
  phase: Pending
  restoreSize: 10Gi
```  

#### Fields:

- `volumeRef`: Reference to the `Volume` to be snapshot. The `Snapshot` and `Volume` should be in the same namespace.
- `snapshotClassRef`: Reference to the `SnapshotClass` to define the type of snapshot to be created. The `SnapshotClass` 
should be in the same namespace as the Snapshot.
- `status.phase`: The current phase of the snapshot. It can be `Pending`, `Ready`, `Failed`, or `Deleting`.
- `status.restoreSize`: The size of the data in the snapshot. This will be populated by the system once the snapshot is ready.

### `SnapshotClass` Resource

The `SnapshotClass` resource will have the following structure:

```yaml
apiVersion: storage.api.onmetal.de/v1alpha1
kind: SnapshotClass
metadata:
  name: example-snapshotclass
spec:
  capabilities:
    custom.capability: "true"
```
    
#### Fields:

`spec.capabilities`: This field is of type `corev1.ResourceList` and will store the capabilities of the `SnapshotClass`. 
The capabilities are defined by the `VolumePool` provider and specify the unique features or restrictions 
of the `SnapshotClass`.

### Volume Resource Enhancement

The `Volume` resource will have a new `dataSource` field under `spec`:

```
apiVersion: storage.api.onmetal.de/v1alpha1
kind: Volume
metadata:
  name: restored-volume
spec:
  ...
  volumeClassRef:
    name: example-volumeclass
  dataSource:
    name: example-snapshot
    kind: Snapshot
    apiGroup: storage.api.onmetal.de/v1alpha1
```

#### Fields:

`dataSource`: Information regarding the snapshot to be used as the source for the restoration. Contains:
- `name`: Name of the snapshot.
- `kind`: Kind of the source, which should be `Snapshot`.
- `apiGroup`: The group to which the referenced resource belongs, which would be `storage.api.onmetal.de/v1alpha1`.

### Implementation Strategy

Once the proposal is approved, the implementation will follow these stages:

1. **Development**: The design and implementation of the new `Snapshot` and `SnapshotClass` resources, as well as 
updates to the Volume resource and `onmetal-api` for recognition and handling of these resources.
2. **Testing**: Conduct extensive testing, including unit tests, integration tests, and end-to-end tests, and assess 
potential impact on system performance.
3. **Documentation**: Update documentation to reflect these new resources and the restoration path, including their 
specifications, behaviors, and example usages.
4. **Deployment**: Roll out these changes as part of a new release of the onmetal-api, while communicating these 
changes to users.
5 **Feedback and Iteration**: Gather user feedback, monitor usage, and make further improvements or address issues 
that arise.

## Potential Alternatives

Potential alternatives to the proposed changes could include:

1. **Third-Party Snapshot Tools**: One alternative to implementing a `Snapshot` and `SnapshotClass` resource could be 
to recommend users leverage existing third-party tools or cloud-provider services for managing snapshots. However, 
this approach might lack the integration and ease of use that come with an inbuilt solution.
2. **Extend `Volume` Resource**: Rather than introducing new resources, we could extend the current `Volume` resource 
with snapshot-related functionalities. This could include adding fields to the `Volume` spec that would trigger a 
snapshot when certain conditions are met. However, this could potentially complicate the `Volume` resource and would 
limit the flexibility to manage snapshots independently.
3. **Storage Provider Level Snapshots**: Leave the task of snapshot creation and management to individual storage 
providers. Each storage provider has its own way of managing snapshots, and they could provide their own APIs for 
managing them. This might complicate things for users who have to work with different APIs for different 
storage providers.

Please note that these alternatives come with their own sets of advantages and disadvantages, and a thorough evaluation 
should be conducted to select the most appropriate solution.
