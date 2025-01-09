# BucketPool
A `BucketPool` is a resource in `Ironcore` that represents a pool of storage buckets managed collectively. It defines the infrastructure's storage configuration used to provision and manage buckets, ensuring resource availability and compatibility with associated BucketClasses.

# Example BucketPool Resource
An example of how to define a `BucketPool` resource in `Ironcore`

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: BucketPool
metadata:
  name: bucketpool-sample
spec:
  providerID: ironcore://shared
#status:
#  state: Available
#  availableBucketClasses:
#    ironcore.dev/fast-class: 10Gi
#    ironcore.dev/slow-class: 100Gi
```

# Key Fields:
- `ProviderID`(`string`):  The `providerId` helps the controller identify and communicate with the correct storage system within the specific backened storage porvider.

    for example `ironcore://shared`

# Reconciliation Process:

- **Bucket Type Discovery**: It constantly checks what kinds of buckets (BucketClasses) are available in the `Ironcore` Infrastructure.

- **Compatibility Check**: Evaluating whether the BucketPool can create and manage each bucket type based on its capabilities.

- **Status Update**: Updating the BucketPool's status to indicate the bucket types it supports, like a menu of available options.

- **Event Handling**: Watches for changes in BucketClass resources and ensures the associated BucketPool is reconciled when relevant changes occur.