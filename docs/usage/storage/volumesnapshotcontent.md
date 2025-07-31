# VolumeSnapshotContent
The `Ironcore` `VolumeSnapshotContent` represents the actual storage-provider-specific snapshot content.

## Example VolumeSnapshotContent Resource
An example of how to define a `VolumeSnapshotContent` resource in `Ironcore`

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: VolumeSnapshotContent
metadata:
  name: volumesnapshotcontent-sample
spec:
  source:
    volumeSnapshotHandle: 1334353-234234-45435435
  volumeSnapshotRef:
    name: volumesnapshot-sample
    namespace: namespace-sample
    uid: 12345678-1234-5678-1234-123456789012
```

## Key Fields:

- `source`(`string`): Contains the snapshotHandle, which is a unique identifier for the snapshot in the storage provider.

- `volumeSnapshotRef`(`string`): `volumeSnapshotRef` refers to the VolumeSnapshot that this content belongs to. It includes the name, namespace, and UID of the VolumeSnapshot.


## Reconciliation Process:

Reconciliation of `VolumeSnapshotContent` gets carried out by storage provider.