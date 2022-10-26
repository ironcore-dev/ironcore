---
title: Network Loadbalancer

oep-number: 3

creation-date: 2022-10-18

status: implementable

authors:

- "@gehoern"
reviewers:
- "@adracus"
- "@afritzler"
- "@guvenc"
- "@MalteJ"

---

# OEP-3: Network Loadbalancer

## Table of Contents

- [Summary](#summary)
- [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
    - [Details](#details)
- [Proposal](#proposal)

## Summary
Network service for Load Balancing is needed modern network architecture(e.g. requirement for Kubernetes services). Load balancing makes services scalable, fault-tolerant and accessible from external users. The load balancer described here is a L3-LoadBalancer (implements an IP level load balancing, not port and not protocol). The load balancer is a network function that will be to be deployed on any platform/node being capable of running load balancer functions. Load balancers do not require to be co-located to load balancer targets but usually network regions. No cross region load balancer is described here.

## Motivation
The onmetal network is a fully routed network. We announce a VirtualIP ([OEP-1](01-networking-integration.md#the-virtualip-type)) which is a single stable public IP attached to a `NetworkInterface`, to scale beyond single network interfaces, LoadBalancers offer the capability to dynamically route traffic to multiple targets (Kubernetes equivalent is a cluster IP's) and not only on a single node (node IP's). 
Solutions like ECMP (Equal cost multi pathing) seem to be the natural but ECMP has certain limitations: the amount of nodes it can be used to distribute traffic to and a hash based algorithm to pin the network flows. Everytime the number of targets changes the pining hash buckets are newly shuffled. So ECMP is not sufficient.
To avoid those limitations and to have better control over the load balancer behavior an onmetal load balancer service will be provided.

### Goals
- Define an API for managing L3 load balancers with publicly available addresses
- LoadBalancer are IP only and allows to define one IPv4 and one IPv6 addresses (outside facing) and 
- Multiple target's based on the onmetal ([OEP-1](01-networking-integration.md#the-networkinterface-type)) network interface architecture
- The load balancer object allows dynamic changes (adding/removing) of targets during its lifetime.
- Load balancer (`ips`) and targets can be parts of different `Networks` but
- All targets (`NetworkInterfaces`) must be in the same `Network`
- The load balancer needs to avoid forwarding unintended traffic.
  - based on ports/protocols (UDP/TCP/SCTP) it must filter what needs to be load balanced
  - ICMP requests are not part of forwarded traffic and will be filtered by the load balancer
  - filtering L4 based but does not change IP objects, just decides to load balance or ignore packets
- load balancing will be transparent for both target and source 

### Non-Goals
- No address or port translation / rewriting (no SNAT / DNAT) (L4 Loadbalancer) support
- No injection of additional information (e.g. x-forwarded-for) (L7 Loadbalancer) support
- No protocol offloading like ssl (L7 Loadbalancer) support
- If multiple load balancer IP's are needed, request multiple load balancers (a load balancer services one load balancer IP)

### Details
- Load balancing is used to deliver a packet addressed to the load balancer to one of its targets via the onmetal network routing
- The target needs to be aware of the load balancer's IP and needs to answer with it (and to receive traffic with it)
- Answers to request will be directly delivered since all details are known by the target
- Payload packages are not changed (ingress / egress), to not lose any information

## Proposal
A `LoadBalancer` allows the user to define the network function. All fields are immutable, but the `networkInterfaceSelector` allows for a dynamic target selection, based e.g. on labels (default Kubernetes selector arithmetics). 
The `LoadBalancer` of `type: Public` requests ips for the defined `ipFamilies` out of the public ip address pool and persists its status in `status.ips`. It is possible to assign over the `VirtualIP` object ([OEP-1](01-networking-integration.md#the-virtualip-type)) an explicit and persistent ip to the `LoadBalancer`. The `LoadBalancer` is not aware of any explicit services of higher levels in the software stack: e.g. in Kubernetes the NodeIP's can be targets of the `LoadBalancer` but the actual service dispatching is done by the Kubernetes Ingress.
`ports` is the definition of the filtering mechanism for interesting traffic. This is purely filtering but not changing packages and coveres ports (as `port:`) and port ranges (as `port:` and `portEnd:` combinations)
`networkRef` defines explicitly the `networkRef` of a `NetworkInterface` ([OEP-1](01-networking-integration.md#the-networkinterface-type)) in the `networkInterfaceSelector`.

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: LoadBalancer
metadata: 
  name: myLoadBalancer-2abf34
  namespace: customer-1
spec:
  #the load balancer IP. is a generated IP and only of type public (internet IP) 
  type: Public
  #a load balancer can have two IP addresses: IPv4 or IPv6 but not more addresses IPfamily
  ipFamilies: [ IPv4, IPv6 ]

  #filtering the load balanced traffic by protocol and port/port ranges 
  ports:
    - name: webserver  # might be optional
      # protocols supported UDP, TCP, SCTP
      protocol: tcp
      # single port
      port: 80
    - name: db
      protocol: udp
      # portrange
      port: 1024
      portEnd: 2048
  #all networkInterfaces must be connected to this network.
  networkRef:
    name:
  #a selector for the NetworkInterfaces to be load balancer targets, normal kubernetes selector logic
  networkInterfaceSelector:
    matchLabels: 
      key: db
      foo: bar
status:
  #the ips the load balancer is listening to (max 2, an IPv4 and/or IPv6)
  ips: 
    - 45.86.152.88
    - 2001::
```

### routing state object
The load balancer needs details computable at the onmetal API level to describe the explicit targets in a pool traffic is routed to. `LoadBalancerRouting` describes `NetworkInterfaces` load balanced traffic is routed to.
This object describes a state of the `LoadBalancer` and results of the `LoadBalancer` definition specifically `networkInterfaceSelector` and `networkRef`. The object is not directly changeable.

```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: LoadBalancerRouting
metadata:
  #name is identical to the LoadBalancer.networking.api.onmetal.de object name since it will be mapped on it
  name: myLoadBalancer-2abf34
  namespace: customer-1
#networkRef of the LoadBalancer.networking.api.onmetal.de object. All destination interfaces will be part of this network
networkRef:
  name: my-network
#destination objects of the NetworkInterfaces. Explicit connections containing the k8s object uuids.
destinations:
  - name: my-machine-interface-1
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
  - name: my-machine-interface-2
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73d
```

