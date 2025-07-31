# VolumeSnapshot
The `Ironcore` `VolumeSnapshot` resource allows users to take point-in-time snapshots of the content of a `Volume` without creating an entirely new volume. This functionality allows users to take backup before performing any modifications on data.

## Example VolumeSnapshot Resource
An example of how to define a `VolumeSnapshot` resource in `Ironcore`

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: VolumeSnapshot
metadata:
  name: volumesnapshot-sample
spec:
  volumeRef:
    name: volume-sample
```

## Key Fields:

- `volumeRef`(`string`): `volumeRef` refers to the name of an Ironcore `volume` to create a volume snapshot.


## Reconciliation Process:

- **Fetch VolumeSnapshot Resource**: Retrieve the `VolumeSnapshot` resource and clean up any orphaned `IRI` volume snapshots if the resource is missing.

- **Add Finalizer**: Ensure a finalizer is added to manage cleanup during deletion.

- **Check IRI VolumeSnapshots**: List and identify `IRI` volume snapshots linked to the `VolumeSnapshot` resource.

- **Create VolumeSnapshot**: Create a new IRI volume snapshot if none exists.

- **Sync Status**: Reflect the IRI volume snapshot's state (e.g., Pending, Ready, Failed) in the `VolumeSnapshot` resource status.

- **Handle Deletion**: Safely delete all associated IRI volume snapshots and remove the finalizer to complete the resource lifecycle.
