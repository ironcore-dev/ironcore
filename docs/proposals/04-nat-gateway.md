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
NAT gateways are essential for safer and resource saving internet access. Any machine (even those with no public / virtual IP) can access the internet without being directly exposed. One IP is shared between multiple clients. Communication initiated by the machine is able to get answers from outside (connection tracking) but the machine cannot be contacted (no remote initiated traffic) from outside. A default gateway is provided via the definition of a NAT Gateway to anyone having no access to the internet. Besides this standard behavior the Gateway is designed as a Cloud-NAT. Not a single Gateway is used but multiple small gateways directly in front of the machine.

## Motivation
Some machines of the network have no public IP addresses (`virtualIP` of `NetworkInterface` is not defined). But those machines may also need public internet access. A NAT gateway provides this functionality by defining a default gateway to the private machines and an NAT for incoming and outgoing traffic (e.g. downloading container images). A standard setup without a `NATGateway` is to use a `VirtualIp` ([OEP-1](01-networking-integration.md#the-virtualip-type)) but this exposes the machine to public internet.

### Goals
- Define an API for managing NAT gateways with publicly available addresses
- Defining the maximum outgoing connections to a single remote target of any `NetworkInterface` in the network
- Define the name of the `Network` and the `NetworkInterface` the NAT gateway is operating on 

### Non-Goals
- The NAT gateway is not transparent since it manipulates the source port for outgoing traffic towards the remote target
- The NAT gateway is not a single entity creating a bottleneck

## Proposal
If a `NATGateway` is defined for a network defined by a `networkRef` all `NetworkInterfaces` selected by `networkInterfaceSelector` will get NATed or masqueraded internet access. `NetworkInterfaces` that already have a `VirtualIP` will be ignores (because they have their own public IP that is used to access the public internet) ([OEP-1](01-networking-integration.md#the-networkinterface-type))
The `NATGateway` of `type: Public` request as many ips for the defined `ipFamily` as listed out of the public ip address pool and persists its status in `status.ips`. It is possible to assign over the `VirtualIP` object ([OEP-1](01-networking-integration.md#the-virtualip-type)) explicit and persistent ips to the `NATGateway`, so an explicit IP fort the `NATGateway` can be guaranteed. The amount of ips listed for the `ipFamily` should be identical for both families or ignore one family completely.
All fields are immutable exept the `natIPs`, but the `networkInterfaceSelector` allows for a dynamic selection of NAT'ed machines, based e.g. on labels (default Kubernetes selector arithmetics).
The field `portsPerNetworkInterface` defines the maximum of concurrent connections from one `NetworkInterface` to one remote IP. This is needed in the concept of a Cloud-NAT to trace the traffic back. The `portsPerNetworkInterface` have an effect on how `portsUsed`. If a lot of machines are NAT'ed with a high amount of `portsPerNetworkInterface` a high amount of `natIPs` is used.

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
  # public IP addresses the NATGateway has. Every IP has a total of 64512 (655356-1024) ports being available to machines in the NAT domain To have more ports available add more IP addresses 
  type: Public
  natIPs:
    - name: ip1
      ipFamily: IPv4
  # defines the concurrent connections per NetworkInterface and target. 64 is the default (if omitted), must be a power of 2
  portsPerNetworkInterface: 64                    # 64, 128, 256, 512 or 1024
  # a selector for the NetworkInterfaces to be part of the NATGateway. That way it is possible to define interfaces that have explicit Internet access and interfaces that do not have. All interfaces are part of NetworkRef, matching the label by k8s label selector rules and have no VirtualIP 
  networkInterfaceSelector:
    matchLabels:
      key: db
      foo: bar
status:
  ips:
    - 48.86.152.12    
  # information of Network interfaces without a virtualIP is needed
  portsUsed: {{ PortsPerMachine * entries in NATGatewayRouting.destinations }}
```

### routing state object
The `NATGateway` needs a persistent list of `NetworkInterfaces` that are part of the `NATGateway` (Interfaces of `networkRef` without `virtualIP`) and the assigned port ranges (calculated from `portsPerNetworkInterface` resulting in a range from `port:` to `portEnd:`). If e.g. a `NetworkInterface` would get a `virtualIP` it would be removed off the routing list but the assigned Ports to other `NetworkInterface`s do not change.
This object describes a state of the `NATGateway` and results of the `NATGateway` definition specifically `networkInterfaceSelector` and `networkRef`. The object is not directly changeable.

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NATGatewayRouting
metadata:
  #name is identical to the NATGateway.networking.api.onmetal.de object name since it will be mapped on it
  name: my-nat-4711ab
  namespace: customer-1
#networkRef of the NATGateway.networking.api.onmetal.de object. All destination interfaces will be part of this network
networkRef:
  name: my-network
#destination objects of the NetworkInterfaces. Explicit connections containing the k8s object uuids. And it contains the ports the object will be using for outgoing/incomming traffic on the corresponding ip
destinations:
  - name: my-machine-interface-1
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
    natIP: 45.86.152.12
    port: 1024
    portEnd: 1087
  - name: my-machine-interface-2
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73d
    natIP: 45.86.152.12
    port: 1088
    portEnd: 1151
```
