---
title: NAT Gateway

oep-number: 4

creation-date: 2022-18-10

status: implementable

authors:

- "@gehoern"
  reviewers:
- "@MalteJ"
- "@adracus"
- "@afritzler"
- "@guvenc"

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
to the NAT gateway are shared between multiple clients. Communication initiated by the a member can get
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
- Define an API for managing NAT gateways with publicly available addresses
- Define the maximum ports of a NAT gateway to be used by a target `NetworkInterface`.
- Define the name of the `Network` and the `NetworkInterface` the NAT gateway is operating on 

### Non-Goals
- The NAT gateway is not transparent since it manipulates the source port for outgoing traffic towards the remote target
- The NAT gateway is not a single entity creating a bottleneck

## Proposal
Introduce a `NATGateway` resource that targets `NetworkInterface`s in a `Network`.
The `Network` is specified via a `networkRef`, the `NetworkInterface`s are targeted via a 
`corev1.LabelSelector`. During reconciliation, only `NetworkInterface`s that are not yet exposed via `VirtualIP`
are selected and will be NATed and get masqueraded internet access.
To denote a `NATGateway` as publicly facing, `type: Public` must be specified. For now, this is the only supported
type.
A `NATGateway` must specify the IP stack it operates on via `ipFamilies`. This can be `IPv4`, `IPv6` or both (dual-stack).
The `ips` field names the ips allocated for a `NATGateway`. If `ipFamilies` is dual-stack, both an `IPv4` and `IPv6` ip address will be allocated for each item in the `ips` field.
The field `portsPerNetworkInterface` defines the maximum number of concurrent connections from a single
`NetworkInterface` to a remote IP.
The current usage of ports is reported in `status.portsUsed`.

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NATGateway
metadata:
  name: my-nat-4711ab
  namespace: customer-1
spec:
  # the network the NATGateway operates on
  networkRef: 
    name: sample-network
  # type denotes the type of the nat gateway. For now, only 'Public' is supported.
  ips:
    - name: ip1
      ipFamily: IPv4
  # defines the concurrent connections per NetworkInterface and target. 64 is the default (if omitted), must be a power of 2
  portsPerNetworkInterface: 64                    # 64, 128, 256, 512 or 1024
  # networkInterfaceSelector selects the target network interfaces that should be NATed.
  networkInterfaceSelector:
    matchLabels:
      key: db
      foo: bar
status:
  ips:
    - 48.86.152.12    
  # portsUsed reports the current port usage of the nat gateway.
  portsUsed: {{ PortsPerMachine * entries in NATGatewayRouting.destinations }}
```

### Routing State Object
The actual routing of a `NATGateway` at a certain point in time is reflected via `NATGatewayRouting`
(similar to `LoadBalancerRouting` / `AliasPrefixRouting`). It denotes the target `Network` instance (including
the instance's `uid`) and the target `NetworkInterface`s alongside the used IPs and ports.

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NATGatewayRouting
metadata:
  # The name of a `NATGatewayRouting` entity is the same as the name of the `NATGateway` object
    it's created from.
  name: my-nat-4711ab
  namespace: customer-1
# networkRef is a reference to the network instance all network interfaces are part of.
networkRef:
  name: my-network
# destination lists the target network interface instances alongside the ip  and port range used for them.
destinations:
  - name: my-machine-interface-1
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
    ips:
    - ip: 45.86.152.12
      port: 1024
      portEnd: 1087
    port: 1024
    portEnd: 1087
  - name: my-machine-interface-2
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73d
    natIP: 45.86.152.12
    port: 1088
    portEnd: 1151
```
