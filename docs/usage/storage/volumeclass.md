# VolumeClass
The `VolumeClass` in `Ironcore` is a Kubernetes-like abstraction that defines a set of parameters or configurations for provisioning storage resources through the `Ironcore Runtime Interface (IRI)`. It is conceptually similar to Kubernetes `StorageClass`, enabling users to specify the desired properties for an Ironcore `Volume` resource creation.

# Example VolumeClass Resource
An example of how to define a `VolumeClass` resource in `Ironcore`

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: VolumeClass
metadata:
  name: volumeclass-sample
capabilities:
  tps: 100Mi
  iops: 100
```

# Key Fields:
- `capabilities`: Capabilities has tps and iops fields that need to be specified, it's a mandatory field,
  - `tps`(`string`): The `tps` represents transactions per second.

  - `iops`(`string`): `iops` is the number of input/output operations a storage device can complete per second.

# Usage

- **VolumeClass Definition**: Create a `VolumeClass` to set storage properties based on resource capabilities.

- **Associate with Volume**: Link a `VolumeClass` to a `Volume` using a reference in the Volume resource.

- **Dynamic configuration**: Update the `VolumeClass` to modify storage properties for all its Volumes.

# Reconciliation Process:

- **Fetches & Validates**: Retrieves the VolumeClass from the cluster and checks if it exists.

- **Synchronizes State**: Keeps the VolumeClass resource updated with its current state and dependencies.

- **Monitors Dependencies**: Watches for changes in dependent Volume resources and reacts accordingly.

- **Handles Errors**: Requeues the reconciliation to handle errors and ensure successful completion.

- **Handles Deletion**: Cleans up references, removes the finalizer, and ensures no dependent Volumes exist before deletion.
