# VolumePool
A `VolumePool` is a resource in `Ironcore` that represents a pool of storage volume managed collectively. It defines the infrastructure's storage configuration used to provision and manage volumes, ensuring resource availability and compatibility with associated `VolumeClasses`.

# Example VolumePool Resource
An example of how to define a `VolumePool` resource in `Ironcore`

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: VolumePool
metadata:
  name: volumepool-sample
spec:
  providerID: ironcore://shared
#status:
#  state: Available
#  availableVolumeClasses:
#    ironcore.dev/fast-class: 10Gi
#    ironcore.dev/slow-class: 100Gi
```

# Key Fields:
- `providerID`(`string`): The `providerId` helps the controller identify and communicate with the correct storage system within the specific backened storage porvider.

    for example `ironcore://shared`

# Reconciliation Process:

- **Volume Type Discovery**: It constantly checks what kinds of volumes (volumeClasses) are available in the `Ironcore` Infrastructure.

- **Compatibility Check**: Evaluating whether the volumePool can create and manage each volume type based on its capabilities.

- **Status Update**: Updating the VolumePool's status to indicate the volume types it supports, like a menu of available options.

- **Event Handling**: Watches for changes in VolumeClass resources and ensures the associated VolumePool is reconciled when relevant changes occur.
