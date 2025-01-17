# Prefix
A `Prefix` resource provides fully integrated IP address management(IPAM) solution for `Ironcore`. It serves as a means to define IP prefix along with prefix size to reserved range of IP addresses. It is also possible to define child subnets with the specified prefix length referring to the parent prefix.

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
  prefix: 10.0.0.0/8

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

- **Allocate root prefix**: If `parentRef` and `parentSelector` is empty, the PrefixController reconciler consideres it as a root prefix and allocates by itself and the status is updated as `Allocated`.

- **Allocating sub-prefix**: If `parentRef` or `parentSelector` is set PrefixController lists all the previouslly allocated prefix allocations by parent prefix. Finds all the active allocations and prunes outdated ones. If no existing PrefixAllocation object is found new `PrefixAllocation` object is created for the new prefix to allocate. If prefix allocation is successful status is update to `Allocated` otherwise to `Failed` in PrefixAllocation.

- **Prefix allocation scheduler**: PrefixAllocationScheduler continously watches for Prefix resource and tries to schedule all PrefixAllocation objects for which prefix is not yet allocated. PrefixAllocationScheduler determins suitable prefix for allocation by listing available prefixes based on label filter, namespace and desired IP family. Once suitable prefix is found PrefixAllocation spec.parentRef is updated with selected prefix reference.

Once prefix allocation is successfull for sub-prefix, parent Prefix's status is updated with used prefix IP list.