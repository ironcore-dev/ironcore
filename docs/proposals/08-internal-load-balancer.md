---
title: IEP Title

iep-number: 8

creation-date: 2023-03-16

status: implementable

authors:

- "@afritzler"
- "@adracus"

---

# IEP-8: Internal Load Balancers

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
- [Alternatives](#alternatives)

## Summary

When developing services in the cloud, not all services should be available to the public internet.
Nevertheless they need to be highly available within a certain network boundary.

To solve this issue, other public cloud vendors allow defining internal load balancers.

Internal load balancers behave much like their external / public counterparts with the important
difference of not exposing the selected service to the public internet.

## Motivation

In `ironcore`, we need to be able to make services highly available internally. Currently,
we can only allocate IP addresses for `NetworkInterface`s and target them, however, as soon as the
backing `Machine` fails, the service would become unavailable.

To prevent this, we have to extend our current `LoadBalancer` type to also function internally.
This should be done with a similar API as for the public use case but allow for the same
flexibility with internal IPs as we have already with the `NetworkInterface` type.

### Goals

* Make a service running on multiple `Machine`s / `NetworkInterface`s in a single `Namespace`
  available behind a load-balanced IP without exposing it outside their `Network`

* Manage the used internal IPs either via literals or by using `ipam.Prefix`es.

* Extend the current `LoadBalancer` type with `type: Internal` indicating its use as an internal
  load balancer.

### Non-Goals

* Cross-Namespace consumption of the `LoadBalancer` - An internal `LoadBalancer` is only
  available within one `Network`

* Cross-Namespace IP allocation - IPs and prefixes are created and deleted in a single namespace.

* Use `VirtualIP`s in a `LoadBalancer` of `type: Internal`:
  `VirtualIP`s are always public (compare to AWS' `ElasticIP`) and their IP allocation differs
  from allocating internal IPs (that don't have to be publicly routable / announced via ASN).
  If a user always wants the same internal IP, this use case is already covered by specifying a
  literal IP value or by referencing an `ipam.Prefix`.

## Proposal

Example manifest:

```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: LoadBalancer
metadata:
  name: my-loadbalancer
spec:
  type: Internal
  networkRef:
    name: my-network
  ipFamilies: [IPv4, IPv6]
  ips:
  - ip: 10.0.0.1 # It is possible to specify a literal IP
  - ephemeral: # Or to allocate using an existing ipam.Prefix
      prefixTemplate:
        spec:
          prefixRef: # The prefix length will always = IPFamily.Bits
            name: my-lb-prefix-v6
  networkInterfaceSelector:
    matchLabels:
      app: web
  ports: # The port filtering is the same as for public load balancers
  - protocol: TCP
    port: 8080
status:
  ips:
  - ip: "10.0.0.1"
  - ip: "2607:f0d0:1002:51::4"
```

## Alternatives

* Target multiple `NetworkInterface`s + IPs and track (e.g. via discovery / concensus protocol)
  which of these are available to solve the high-availability aspect. This has the drawback of
  high implementation effort + having to choose from multiple IPs / putting the burden of choosing
  the correct IP on a potential consumer.

* Create `Machine`(s) that e.g. run `HAProxy` to target your services with. This comes with the
  drawback of having to manage the `Machine`s / having multiple IPs for the load balancing machines
  to choose from (see alternative 1).
