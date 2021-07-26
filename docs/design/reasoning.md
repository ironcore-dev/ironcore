# Why certain design decisions have been made

## Namespace replication

Namespaces should be replicated from the user facing (global API) to the 
corresponding region specific API machinery based on the regional notion of the
namespace. This is determined by the `region` field in the scope definition

```yaml
apiVersion: core.onmetal.de/v1alpha1
kind: Scope
metadata:
  name: scope-sample
spec:
  region: frankfurt1
  description: "my scope"
```

In case the `region` field is empty, as it is an optional field, the namespace will
be synced down to all infrastructure clusters.

The idea of keeping track of how many regional objects are in a particular namespace
has been dismissed, as it is hard to track how many of those objects are inside a namespace, 
especially since this has to be done in the `parant` hierarchy as well.

As a result of this descission the `region` field on all given objects must be immutable.