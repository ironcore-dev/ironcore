# BucketClass
A `BucketClass` is a concept used to define and manage different types of storage buckets, typically based on resource capabilities. It is conceptually similar to Kubernetes `StorageClass`, enabling users to specify the desired properties for an Ironcore `Bucket` resource creation.

# Example BucketClass Resource
An example of how to define a `BucketClass` resource in `Ironcore`

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: BucketClass
metadata:
  name: bucketclass-sample
capabilities:
  tps: 100Mi
  iops: 100
```

# Key Fields:

- `capabilities`: Capabilities has `tps` and `iops` fields which need to be specified, it's a mandatory field,
  - `tps`(`string`): The `tps` represents transactions per second.

  - `iops`(`string`):  `iops` is the number of input/output operations a storage device can complete per second.

# Usage

- **BucketClass Definition**: Create a `BucketClass` to set storage properties based on resource capabilities.

- **Associate with buckets**: Link a `BucketClass` to a `Bucket` using a reference in the Bucket resource.

- **Dynamic configuration**: Update the `BucketClass` to modify storage properties for all its Buckets.

# Reconciliation Process:

- **Fetches & Validates**: Retrieves the `BucketClass` from the cluster and checks if it exists.

- **Synchronizes State**: Keeps the `BucketClass` resource updated with its current state and dependencies.

- **Monitors Dependencies**: Watches for changes in dependent Bucket resources and reacts accordingly.

- **Handles Errors**: Requeues the reconciliation to handle errors and ensure successful completion.

- **Handles Deletion**: Cleans up references, removes the finalizer, and ensures no dependent Buckets exist before deletion.
