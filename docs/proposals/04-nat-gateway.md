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
NAT gateways are essential for safer and resource saving internet access. Any machine (even those with no public / virtual IP) can access the internet without being directly exposed. One IP is shared between multiple clients. Communication initiated by the machine is able to get answers but the machine cannot be contacted (no remote initiated traffic). A default gateway is provided via the definition of a NAT Gateway to anyone having no access to the internet. Besides this standard behavior the Gateway is designed as a cloud NAT. Not a single Gateway is used but multiple small gateways directly in front of the machine.

## Motivation
Some machines of the network have no public IP addresses. But those machines also need public internet access. A NAT gateway provides this functionality by providing a default gateway to the private machines and an address translation for incoming and outgoing traffic. OnMetal needs also to provide access to private machines (e.g. downloading container images). A standard setup is to use a VirtualIp ([OEP-1](01-networking-integration.md)) but this exposes the machines with a public IP address. This service introduces NAT gateway for everyone not having a public IP address.

### Goals
- define a user facing API (onMetal api) for the NAT Gateway
- defining the maximum outgoing connections of any machine in the network
- defining the (public) IP Address of the gateway

### Non-Goals
- gateways are network function 

## Proposal
The NatGateway definition allows to define a gateway for a network. The prefix is implying the Network (VNI) this NAT-Gateway runs on. Better: every machine in that prefix will get a NAT-Gateway. If a machine gets a virtual IP it no longer needs a NAT-Gateway assignment. Anything except the `natIPs` is immutable on the NatGateway. Those are used to scale the size of the Gateway. Changes of the `PortPerMachine` cause the whole NAT-Gateway to be recomputed and can therefore only happen on initializing

### User defined object
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
  #TODO other types of IP addresses (e.g. internal NATGateway) and pinned IP types
  type: Public
  natIPs:
    - ipFamily: IPv4
  # defines the concurrent connections per machine and target. 64 is the default (if omitted), must be a power of 2
  PortsPerMachine: 64                    # 64, 128, 256, 512 or 1024
  # a selector for the NetworkInterfaces to be part of the NATGateway. That way it is possible to define interfaces that have explicit Internet access and interfaces that do not have. All interfaces are part of NetworkRef, mathing the label by k8s label selector rules and have no VirtualIP 
  networkInterfaceSelector:
    matchLabels:
      key: db
      foo: bar
status:
  ips:
    - 48.86.152.12    
  # information of Network interfaces without a virtualIP is needed
  portsUsed: {{ PortsPerMachine * entries in NATGatewayRouting.destinations }}
  portsAvailable: {{ ( entries in natIPs * 64512 ) - PortsUsed}}
```

### internal reconcile object (not customer visible)
The NATGateway needs a persistent list of NetworkInterfaces that are part of the NATGateway (Interfaces of `networkRef` without `virtualIP`) and the assigned ports. If e.g. a NetworkInterface would get a virtualIP it would be off the list but the assigned Ports to other machines do not change.
This object should not be visible to the customer but is available in the customer namespace. It can be computed by reconciling on the objects the NATGateway describes.

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
