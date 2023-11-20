---
title: NAT Gateway

oep-number: 4

creation-date: 2022-18-10

status: implementable

authors:

- @gehoern
- @adracus

reviewers:

- @MalteJ
- @adracus
- @afritzler
- @guvenc

---

# OEP-4: Cloud Nate Gateway

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
- [Proposal](#proposal)

## Summary

NAT gateways are essential for safe and resource-efficient internet access. Any machine (even those with no
public / virtual IP) using a NAT gateway can access the internet without being directly exposed. IPs belonging
to the NAT gateway are shared between multiple clients. Communication initiated by a member can get
answers from outside (connection tracking) but the member cannot be contacted (no remote initiated traffic) from
outside.

## Motivation

A `NetworkInterface`s may have no dedicated public IP addresses (no `VirtualIP` via `spec.virtualIP`) but still may
need public internet access.
A NAT gateway provides this functionality by defining a default gateway to the network interface
and a NAT for incoming and outgoing traffic (e.g. downloading container images).
The `NetworkInterface` thus can reach the public internet but is not exposed as it would be when using a
`VirtualIP`.

### Goals

- A public NAT gateway should only target a single `Network`.
- Define an API for managing NAT gateways with publicly available addresses.
- Define the maximum ports of a NAT gateway to be used by a target `NetworkInterface`.
- Define the name of the `Network` and the `NetworkInterface` the NAT gateway is operating on.

### Non-Goals

- The NAT gateway is not transparent since it manipulates the source port for outgoing traffic towards the remote
  target.

## Proposal

Introduce a `NATGateway` resource that targets `NetworkInterface`s in a `Network`.
The `Network` is specified via a `networkRef`, the `NetworkInterface`s are targeted via a
`LabelSelector`. During reconciliation, only `NetworkInterface`s that are not yet exposed via `VirtualIP`
are selected and will be NATed and get masqueraded internet access.
To denote a `NATGateway` as publicly facing, `type: Public` must be specified. For now, this is the only supported
type.
A `NATGateway` must specify the IP stack it operates on via `ipFamilies`. This can be `IPv4`, `IPv6` or both (
dual-stack).
The `ips` field names the ips allocated for a `NATGateway`. If `ipFamilies` is dual-stack, both an `IPv4` and `IPv6` ip
address will be allocated for each item in the `ips` field.
The field `portsPerNetworkInterface` defines the maximum number of concurrent connections from a single
`NetworkInterface` to a remote IP.
The current usage of ports is reported in `status.portsUsed`.

[//]: # (@formatter:off)
```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: NATGateway
metadata:
  namespace: default
  name: my-nat
spec:
  # Type denotes the type of nat gateway. For now, only 'Public' is supported.k:w
  type: Public
  # ip families specifies the supported IP stack of a load balancer. May be `IPv4`, `IPv6` or both (dual stack).
  ipFamilies: [ IPv4, IPv6 ]
  # the network the nat gateway targets.
  networkRef:
    name: sample-network
  # ips are the ips to allocate for the nat gateway.
  # If dual-stack is active, at least two ips will be allocated.
  ips:
    - name: ip1
  # defines the concurrent connections per NetworkInterface and target. Must be a power of 2.
  portsPerNetworkInterface: 64
  # networkInterfaceSelector selects the target network interfaces that should be NATed.
  networkInterfaceSelector:
    matchLabels:
      key: db
      foo: bar
status:
  # ips lists the ips allocated for each requested ip.
  ips:
    - name: ip1
      ips:
      - 48.86.152.12
  # portsUsed reports the current port usage of the nat gateway.
  portsUsed: 128 # Equal to portsPerNetworkInterface * entries in routing destinations
```
[//]: # (@formatter:on)
