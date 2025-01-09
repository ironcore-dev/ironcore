# Bucket
A `Bucket` in `Ironcore` refers to a storage resource that organizes and manages data, similar to the concept of buckets in cloud storage services like Amazon S3. Buckets are containers for storing objects, such as files or data blobs, and are crucial for managing storage workloads.

# Example Bucket Resource
An example of how to define a `Bucket` resource in `Ironcore`

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: Bucket
metadata:
  name: bucket-sample
spec:
  bucketClassRef:
    name: bucketclass-sample
#  bucketPoolRef:
#    name: bucketpool-sample
```

# Key Fields:
- `bucketClassRef`(`string`): 
  - Mandatory field
  - `BucketClassRef` is the BucketClass of a bucket

- `bucketPoolRef`(`string`):
  - Optional field
  -  `bucketPoolRef` indicates which BucketPool to use for the bucket, if not specified the controller itself picks the available bucketPool


# Usage
- **Data Storage**: Use `Buckets` to store and organize data blobs, files, or any object-based data.

- **Multi-Tenant Workloads**: Leverage buckets for isolated and secure data storage in multi-tenant environments by using separate BucketClass or BucketPool references.

- **Secure Access**: Buckets store a reference to the `Secret` securely in their status, and the `Secret` has the access credentials, which applications can retrieve access details from the `Secret`.

# Reconciliation Process:
- The controller detects changes and fetches bucket details.

- Creation/Update ensures the backend bucket exists, metadata is synced, and credentials are updated.

- The bucket will automatically sync with the backend storage system, and update the Bucket's state (e.g., `Available`, `Pending`, or `Error`) in the bucket's status.

- Access details and credentials will be managed securely using Kubernetes `Secret` and the bucket status will track a reference to the `Secret`.

- During deletion, resources will be cleaned up gracefully without manual intervention.

- If the bucket is not ready (e.g., backend issues), reconciliation will retry
