# Prefix
A `Prefix` resource provides a fully integrated IP address management(IPAM) solution for `Ironcore`. It serves as a means to define IP prefixes along with prefix length to a reserved range of IP addresses. It is also possible to define child subnets with the specified prefix length referring to the parent prefix.

# Example Volume Resource
An example of how to define a root `Prefix` resource in `Ironcore`

```
apiVersion: ipam.ironcore.dev/v1alpha1
kind: Prefix
metadata:
  name: root
  labels:
    subnet-type: public
spec:
  prefix: 10.0.0.0/24

```
An example of how to define a child `Prefix` resource in `Ironcore`

```
apiVersion: ipam.ironcore.dev/v1alpha1
kind: Prefix
metadata:
  name: customer-subnet-1
spec:
  ipFamily: IPv4
  prefixLength: 9
  parentSelector:
    matchLabels:
      subnet-type: public

```
(`Note`: Refer to <a href="https://github.com/ironcore-dev/ironcore/tree/main/config/samples/e2e/">E2E Examples</a> for more detailed example on IPAM to understant e2e flow)

# Key Fields:

- `ipFamily`(`string`): `ipFamily` is the IPFamily of the prefix. If unset but `prefix` is set, this can be inferred.

- `prefix` (`string`): 	`prefix` is the IP prefix to allocate for this Prefix.

- `prefixLength` (`int`): `prefixLength` is the length of prefix to allocate for this Prefix.

- `parentRef` (`string`): `parentRef` references the parent to allocate the Prefix from. If `parentRef` and `parentSelector` is empty, the Prefix is considered a root prefix and thus allocated by itself.

- `parentSelector` (`LabelSelector`): `parentSelector` is the LabelSelector to use for determining the parent for this Prefix.


# Reconciliation Process:

- **Allocate root prefix**: If `parentRef` and `parentSelector` is empty, the PrefixController reconciler considers it as a root prefix and allocates by itself and the status is updated as `Allocated`.

- **Allocating sub-prefix**: If `parentRef` or `parentSelector` is set PrefixController lists all the previously allocated prefix allocations by parent prefix. Finds all the active allocations and prunes outdated ones. If no existing PrefixAllocation object is found new `PrefixAllocation` object is created for the new prefix to allocate. If prefix allocation is successful status is updated to `Allocated` otherwise to `Failed`.

- **Prefix allocation scheduler**: `PrefixAllocationScheduler` continuously watches for Prefix resource and tries to schedule all PrefixAllocation objects for which prefix is not yet allocated. PrefixAllocationScheduler determines suitable prefix for allocation by listing available prefixes based on label filter, namespace and desired IP family. Once a suitable prefix is found PrefixAllocation spec.parentRef is updated with the selected prefix reference.

- **Status update**: Once prefix allocation is successful status is updated to `Allocated`. In the case of sub-prefixes once the prefix is allocated `PrefixController` updates the parent Prefix's status with the used prefix IPs list.

Below is the sample `Prefix.status` :

```
status:
  lastPhaseTransitionTime: "2024-10-21T20:56:24Z"
  phase: Allocated
  used:
  - 10.0.0.1/32
  - 10.0.0.2/32
```