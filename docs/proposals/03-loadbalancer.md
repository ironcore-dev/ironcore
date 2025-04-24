---
title: Network Loadbalancer

iep-number: 3

creation-date: 2022-10-18

status: implementable

authors:

- "@gehoern"
- "@adracus"

reviewers:

- "@MalteJ"
- "@adracus"
- "@afritzler"
- "@guvenc"

---

# IEP-3: Network Loadbalancer

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
    - [Details](#details)
- [Proposal](#proposal)

## Summary

Load Balancing is an essential requirement in any modern network architecture.
It makes backend services scalable, fault-tolerant and provides easy-to-consume access to external consumers.

There are multiple types and strategies for load balancing: IP-based load balancing (L3 in the OSI model),
Port-based load balancing (L4) and application-based load balancing (L7). This proposal focuses on IP-based
load balancers, since they can be used as a foundation for the higher level load balancer types.

## Motivation

A `VirtualIP` ([IEP-1](01-networking-integration.md#the-virtualip-type)) allows to expose a `NetworkInterface`
with a stable public IP. Services running on a `Machine` using that `NetworkInterface` can be consumed this way.
However, if the `Machine` or the service running on that `Machine` crashes, the service will have an outage.
To be more resilient and to scale beyond single `NetworkInterface`s, a `LoadBalancer` allows targeting multiple
`NetworkInterface`s and distributes traffic between them.

### Goals

- Define an API for managing L3 load balancers with publicly available addresses
- Load balancers should allow specifying their IP stack (`IPv4` / `IPv6` / dual stack). Public IP addresses
  should be allocated according to the specified IP stack.
- Load balancers should support multiple target `NetworkInterface`s (
  see ([IEP-1](01-networking-integration.md#the-networkinterface-type))
- The load balancer should dynamically watch for target `NetworkInterface`s.
- All target `NetworkInterface`s must be in the same `Network`.
- The load balancer should be able to filter unwanted traffic. The filtering must not alter the packages.
  The following filters should be implemented:
    - Filter depending on ports & protocols (UDP/TCP/SCTP).
    - ICMP requests should be filtered out by default.
- Load balancing must be transparent for both target and source.

### Non-Goals

- No address or port translation / rewriting (no SNAT / DNAT) (L4 Loadbalancer) support
- No injection of additional information (e.g. x-forwarded-for) (L7 Loadbalancer) support
- No protocol offloading like ssl (L7 Loadbalancer) support
- If more load balancer IPs are required than a single load balancer serves, more load balancers have to be requested.

### Details

- Load balancing is used to deliver a packet addressed to the load balancer to one of its targets via the ironcore
  network routing
- The target needs to be aware of the load balancer's IP and needs to answer with it (and to receive traffic with it)
- Answers to the request will be directly delivered since all details are known by the target

## Proposal

Introduce a `LoadBalancer` resource that dynamically selects multiple target `NetworkInterface`s via
a `networkInterfaceSelector` `metav1.LabelSelector` (as e.g. in `AliasPrefix`es).
The `LoadBalancer` of `type: Public` should allocate public IPs for its `ipFamilies` and announce them in
its `status.ips`.
`ports` defines an allow list of which traffic should be handled by a `LoadBalancer`. A `port` consists of
a `protocol`, `port` and an optional `portEnd` to support port range filtering.
`networkRef` defines the target `Network` a `NetworkInterface` has to be in in order to be an eligible target
for traffic forwarding (see [IEP-1](01-networking-integration.md#the-networkinterface-type)).

[//]: # (@formatter:off)
```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: LoadBalancer
metadata:
  namespace: default
  name: my-load-balancer
spec:
  # type denotes which kind of load balancer to create. For now, only `Public` is supported.
  type: Public
  # ip families specifies the supported IP stack of a load balancer. May be `IPv4`, `IPv6` or both (dual stack).
  ipFamilies: [ IPv4, IPv6 ]
  # ports is an allow list of traffic to load balance via port(range) and protocol.
  ports:
    - # protocols supported UDP, TCP, SCTP
      protocol: tcp
      # single port
      port: 80
    - protocol: udp
      # port range
      port: 1024
      portEnd: 2048
  # networkRef specifies the target network any target network interface should be in.
  networkRef:
    name: my-network
  # network interface selector specifies the network interfaces to select for load balancing.
  networkInterfaceSelector:
    matchLabels:
      key: db
      foo: bar
status:
  # ips are the ips allocated for the load balancer.
  ips:
    - 45.86.152.88
    - 2001::
```
[//]: # (@formatter:on)

### Routing State Object

The load balancer needs details computable at the ironcore API level to describe the explicit targets in a pool traffic
is routed to. `LoadBalancerRouting` describes `NetworkInterface`s load balanced traffic is routed to.
This object describes a state of the `LoadBalancer` and results of the `LoadBalancer` definition
specifically `networkInterfaceSelector` and `networkRef`. `LoadBalancerRouting` is reconciled by the `ironcore` load
balancer controller.

[//]: # (@formatter:off)
```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: LoadBalancerRouting
metadata:
  namespace: default
  name: my-load-balancer # Same name as the load balancer it originates from.
# networkRef references the exact network object the routing belongs to.
networkRef:
  name: my-network
# destinations lists the target network interface instances (including UID) for load balancing.
destinations:
  - name: my-machine-interface-1
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
  - name: my-machine-interface-2
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73d
```
[//]: # (@formatter:on)
