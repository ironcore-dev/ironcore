---
title: Refactor Prefix type for IPv6 usage

iep-number: 11

creation-date: 2024-02-25

status: implementable

authors:

- "@afitzler"

reviewers:

- "@guvenc"
- "@MalteJ"

---

# IEP-11: Refactor Prefix type for IPv6 usage

## Table of Contents

- [IEP-11: Public Prefix](#iep-11-public-prefix)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
  - [Alternatives](#alternatives)

## Summary

In an IPv6 environment, dynamically allocating IP prefixes is crucial for efficient address space management, ensuring 
network security, and maintaining scalability. Allowing users to freely choose IPs and prefixes can lead to conflicts, 
inefficient use of the address space, and security vulnerabilities. Furthermore, the use of local ranges contradicts 
the design principles of IPv6, aimed at eliminating the need for such practices by providing a vast address space. 
Therefore, we propose a new approach to how users can assign prefixes to `NetworkInterfaces`, as the current model, 
introduced by the [Networking Integration IEP](01-networking-integration.md), does not align with these principles.

## Motivation

Currently, the `NetworkInterface` allows defining `IPs` and `Prefixes` for both IP families (IPv4 and IPv6) either by 
providing a discrete value or by using an `EphemeralPrefixSource`. The `EphemeralPrefixSource` can be derived from an 
IPv4 or IPv6 `Prefix`, allowing for the free selection of `NetworkInterface` IPs. In the IPv4 realm, it is common 
practice to use a local, non-publicly routable IP range, with internet connectivity ensured via a `NATGateway` for 
egress and a `LoadBalancer` for ingress traffic. However, this approach is not applicable to IPv6.
Here we need to ensure that either a unique local IPv6 address is being used (`fd00::/8`) or an IP or `Prefix` is derived 
from a parent `Prefix` object.

### Goals

- Control the unique distribution of IPs and `Prefixes`.
- Assign IPs and `Prefixes` to `NetworkInterfaces` derived from a requested `Prefix`.
- Disallow the direct assignment of non-unique local IPv6 (`fd00::/8`) addresses to `NetworkInterfaces` and internal `LoadBalancers`.
- Implement non-disruptive changes for existing IPv4-based setups.

### Non-Goals

- Bring-your-own IPv6 public prefix support is not covered by this proposal. 

## Proposal

Restrict the implementation of the `Prefix` type to only allow `prefixLength` when using the IPv6 IP family. This can 
be achieved by adding validation logic that prohibits the creation of a non-unique local IPv6 `Prefix` in the `spec.Prefix` 
field. The allocation of a public routable IPv6 `Prefix` is ensured by a subsequent component and the effective `Prefix`
is stored in the `Status` of the `Prefix` resource.

The further division of the `Prefix` into sub-allocations via  `spec.parentPrefix` reference allows the user to create 
`Prefixes` derived from the parent.

For IPv4, no change is needed as we will continue with the existing API contract.

Below is an example of a `Prefix` resource requesting a prefix of size 56:

```yaml
apiVersion: ipam.ironcore.dev/v1alpha1
kind: Prefix
metadata:
  namespace: default
  name: my-prefix
spec:
  ipFamily: IPv6
  prefixLength: 56
status:
  phase: Pending # Becomes Allocated once the prefix length has been assigned.
  prefix: ffff::/56 # The effective assigned prefix.
```

Here is an example of how a subsequent Prefix, which derives its address from a parent prefix, is defined:

```yaml
apiVersion: ipam.ironcore.dev/v1alpha1
kind: Prefix
metadata:
  namespace: default
  name: my-sub-prefix
spec:
  ipFamily: IPv6
  parentPrefix:
    name: my-prefix
  prefixLength: 128
status:
  phase: Pending # Becomes Allocated once the prefix length has been assigned.
  prefix: ffff::dddd/128 # The effective assigned prefix.
```

In the same manner a validation on the `NetworkInterface` resource will ensure that `spec.IPs` only contain unique local
IPv6 addresses. Alternatively, an ephemeral `Prefix` with a parent relationship to the initial requested `Prefix` ensures 
a unique public IPv6 address allocation for a `NetworkInterface`.

```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: NetworkInterface
metadata:
  namespace: default
  name: my-nic
spec:
  ips:
    ephemeral:
      prefixTemplate:
        spec:
          parentRef:
            name: my-prefix
```

Here the `prefixTemplate` validation will ensure that no non-unique local IPv6 addresses are assigned.

A similar behaviour is ensured for the internal `LoadBalancer` case. The example below illustrates how an internal
`LoadBalancer` can request a unique IPv6 address from a parent `Prefix`.

```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: LoadBalancer
metadata:
  namespace: default
  name: my-lb
spec:
  type: Internal
  ipFamilies:
    - IPv6
  ips:
    - ephemeral:
        prefixTemplate:
          spec:
            parentPrefix:
              name: my-prefix
```

## Alternatives

An alternative to the proposed solution would be to define a dedicated API type acting as a Prefix request for a
specified IPv6 prefix length.
