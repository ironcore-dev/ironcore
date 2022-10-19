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
NAT gateways are essential for safer and resource saving internet access. Any machine (even those with no public / virtual IP) can access the internet without being directly exposed. One IP is shared between multiple clients. Communication initiated by the machine is able to get answers but the machine cannot be contacted (no remote initiated traffic). A default gateway is provided via the definition of a NAT Gateway to anyone having no access to the internet. Besides this standard behavior the Gateway is designed as a cloud NAT. Not one single Gateway is used but multiple small gateways share the load.

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

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NATGateway
metadata:
  name: my-nat-4711ab
  namespace: customer-1
spec:
  # the network the Nat gateway operates on
  networkRef: 
    name: sample-network
  # public ip's the NatGateway has. To have more ports available add more IP addresses 
  type: Public
  natIPs:
    - ipFamily: IPv4
  # defines the concurrent connections per machine
  PortsPerMachine: 64                    # 64, 128, 256, 512 or 1024
status:
 # information of Network interfaces without a virtualIP is needed

  portsUsed: {{ PortsPerMachine * sample-network.networkInterfaceWithoutVirtualIP }}
  portsAvailable: {{ ( natIPs * 64512 ) - PortsUsed}}
```

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: NATGatewayRouting
metadata:
  name: my-nat-4711ab
  namespace: customer-1
networkRef:
  name: my-network
destinations:
  - name: my-machine-interface-1
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
    port: 1024
    portEnd: 1087
  - name: my-machine-interface-2
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73d
    port: 1088
    portEnd: 1151
```
