---
title: Quota

iep-number: 7

creation-date: 2023-01-19

status: implementable

authors:

- "@adracus"

reviewers:

- "@afritzler"
- "@gehoern"
- "@ManuStoessel"

---

# IEP-7: Quota

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Alternatives](#alternatives)

## Summary

Quota is a mechanism to manage and limit the usage of resources across multiple requesting entities.
By introducing quotas, a system can be protected from usage spikes and services can be kept responsive.
Quotas also can ensure that each requesting entity can exercise its right to a fair share of the resources.

## Motivation

[Kubernetes Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/) are a great way to limit
resource consumption for core Kubernetes types (allowing to manage things like overall CPU consumption) and resource
count for all types. However, when it comes to limiting resource usage for custom types (in this special case, the
`ironcore` types), the Kubernetes Quota system falls short of providing means to do so.

For `ironcore` it should be possible to limit the actual requested resources like
the total number of used CPUs, storage and memory as well as limit the count of resources
by a given dimension (e.g. number of `Machine`s for a given `MachineClass`).

### Goals

- Limit resource count in a `Namespace` (by dimension)
- Limit accumulated resource usage in a `Namespace` (by dimension)
- Integrate nicely into the existing Kubernetes `ResourceQuota` concepts

### Non-Goals

- Limit resource count / accumulated resource usage cross-`Namespace`
- Define a system to request quota increases
- Define a user management system
- Couple resource quota to any user system

## Proposal

Introduce a new namespaced type `ResourceQuota` in the new `core` group.
A `ResourceQuota` allows defining hard resource limits that cannot be exceeded.
The limits are defined via `spec.hard` as a `corev1alpha1.ResourceList`.
The currently enforced limits are shown in `status.hard` and the currently used
limits in `status.used`.
Requests to create / update resources that would exceed the quota will fail
with the HTTP status code `403 Forbidden`.

### Compute Resource Quota

For the `ironcore` `compute` group, the following resources can be limited:

| Resource Name   | Description                                                                           |
|-----------------|---------------------------------------------------------------------------------------|
| requests.cpu    | Across all machines in non terminal state, the sum of cpus cannot exceed this value   |
| requests.memory | Across all machines in non terminal state, the sum of memory cannot exceed this value |

### Storage Resource Quota

For the `ironcore` `storage` group, the following

| Resource Name    | Description                                                                           |
|------------------|---------------------------------------------------------------------------------------|
| requests.storage | Across all volumes in non terminal state, the sum of storage cannot exceed this value |

### Object Count Quota

Similar to Kubernetes' object count quota, it is possible to limit the number of resources
per types using the following syntax: `count/<resource>.<group>`. For example,
`count/machines.compute.ironcore.dev` would limit the number of machines from the
`ironcore` `compute.ironcore.dev` group.

### Quota Scopes

To measure / limit usage only for a subset of all resources, a `ResourceQuota` may
specify a `scopeSelector`. A `scopeSelector` may contain multiple expressions and only
matches a resource if it matches the intersection of enumerated scopes.

| Scope        | Description                                               |
|--------------|-----------------------------------------------------------|
| MachineClass | Match machines that reference the specified machine class |
| VolumeClass  | Match volumes that reference the specified volume class   |

By using certain `scopeSelector`s, the quota can only track a specific set of resources.
E.g. for the `MachineClass` `scopeSelector`, only `Machine`s can be tracked.

### Example Manifests

Limit the accumulated amount of `cpu`, `memory` and `storage` across all `Machine`s
and `Volume`s:

[//]: # (@formatter:off)
```yaml
apiVersion: core.ironcore.dev/v1alpha1
kind: ResourceQuota
metadata:
  name: limit-accumulated-usage
spec:
  hard:
    requests.cpu: "1000"
    requests.memory: 200Gi
    requests.storage: 10Ti
```
[//]: # (@formatter:on)

Limit the number of machines for a given machine class:

[//]: # (@formatter:off)
```yaml
apiVersion: core.ironcore.dev/v1alpha1
kind: ResourceQuota
metadata:
  name: limit-large-machines
spec:
  hard:
    count/machines.compute.ironcore.dev: 10
  scopeSelector:
    matchExpressions:
    - scopeName: MachineClass
      operator: In
      values:
        - large
```
[//]: # (@formatter:on)
