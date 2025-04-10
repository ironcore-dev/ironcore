# Volume
The `Ironcore` `Volume` is a storage abstraction provided by the `Ironcore Runtime Interface` `(IRI)` service, designed to integrate with external storage backend for managing persistent storage. It acts as a managed storage unit, ensuring consistency, scalability, and compatibility with Kubernetes workloads.
By integrating Ironcore Volumes with Kubernetes, users benefit from seamless storage management, automation, and advanced features such as encryption and scalability, making it suitable for modern cloud-native and hybrid applications.

# Example Volume Resource
An example of how to define a `Volume` resource in `Ironcore`

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: Volume
metadata:
  name: volume-sample
spec:
  volumeClassRef:
    name: volumeclass-sample
  # volumePoolRef:
  #   name: volumepool-sample
  resources:
    storage: 100Gi
```

# Key Fields:

- `volumeClassRef`(`string`): `volumeClassRef` refers to the name of an Ironcore `volumeClass`( for eg: `slow`, `fast`, `super-fast` etc.) to create a volume,

- `volumePoolRef` (`string`): 	`VolumePoolRef` indicates which VolumePool to use for a volume. If unset, the scheduler will figure out a suitable `VolumePoolRef`.

- `resources`: `Resources` is a description of the volume's resources and capacity.

# Reconciliation Process:

- **Fetch Volume Resource**: Retrieve the `Volume` resource and clean up any orphaned `IRI` volumes if the resource is missing.

- **Add Finalizer**: Ensure a finalizer is added to manage cleanup during deletion.

- **Check IRI Volumes**: List and identify `IRI` volumes linked to the `Volume` resource.

- **Create or Update Volume**:
  - Create a new IRI volume if none exists.
  - Update existing IRI volumes if attributes like size or encryption need adjustments.

- **Sync Status**: Reflect the IRI volume's state (e.g., Pending, Available) in the Kubernetes Volume resource's status.

- **Handle Deletion**: Safely delete all associated IRI volumes and remove the finalizer to complete the resource lifecycle.
