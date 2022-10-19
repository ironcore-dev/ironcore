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
Network service for Load Balancing is needed in every modern infrastructure (e.g. Kubernetes). Load balancing makes services scalable, fault-tolerant and accessible from external users. The load balancer described here is a L3-Loadbalancer (so it implements an IP level load balancing). The load balancer function is a standard network function that can be to be deployed (as an implementation detail) on a platform/node that is capable of running network functions (e.g. dp-service enabled nodes/nic's) 

## Motivation
The onMetal network is a fully routed network. To run workloads hosted on several nodes (in Kubernetes terms equivalent to cluster IP's) and not only on a single node (node IP's) solutions like ECMP (Equal cost multi pathing) seem to be the natural solution. But ECMP has certain limitations: the amount of nodes it can be used to distribute traffic to and a hash based algorithm to pin the network flows, that changes - based on the amount of targets - its distributed pinning. 
To avoid those limitations and to have better control over the load balancer behavior an onMetal load balancer service needs to be provided.

### Goals
- define a user facing API (onMetal api) for the load balancer
- enable the user to request an L3 Loadbalancer (only IP addresses) with load balancer address (outside facing) and 
- with multiple target's based on the onMetal ([OEP-1](01-networking-integration.md)) network interface architecture
- the targets of the load balancer must be dynamic changeable
- the load balancer needs to avoid forwarding unintended traffic.
  - based on ports/protocols (UDP/TCP/SCTP) it must filter what needs to be load balanced
  - ICMP requests are not part of forwarded traffic and will be filtered by the load balancer anyhow (the loa)
- Load balancer (IP) and targets can be parts of different networks but
- all targets (NetworkInterfaces) must be of the same network 

### Non-Goals
- no address or port translation / rewriting (no SNAT / DNAT) (L4 Loadbalancer) support
- no injection of additional information (e.g. x-forwarded-for) (L7 Loadbalancer) support
- no protocol offloading like ssl (L7 Loadbalancer) support
- if multiple load balancer IP's are needed, request multiple load balancers (a load balancer services one load balancer IP)

### Details
- load balancing is used to deliver a packet addressed to the load balancer to one of its targets via the onMetal network routing
- the load balancing will be for the target and the source transparent
- the target needs to know the load balancers IP and needs to answer with it (and to receive traffic with it)
- answers to request will be directly delivered since all details are known to the target
- payload packages are not changed, to not lose any information

## Proposal
A network loadbalancer CRD allows the user to define the network function. all entries except the `targetNetworkInterface` must be immutable. 

### User defined object
```yaml
apiVersion: networking.api.onmetal.de/v1alpha1
kind: LoadBalancer
metadata: 
  name: myLoadBalancer-2abf34
  namespace: customer-1
spec:
  #the load balancer IP. is a generated IP and only of type public (internet IP) 
  #TODO other types of IP addresses (e.g. internal LB) and pinned IP types
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
  #TODO explicit machines ?
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
### internal reconcile object (not customer visible)
The load balancer needs some data computable at the onMetal API level to describe the explicit targets in a pool. Since onMetal is using a complete routing infrastructure this object in fact describes NetworkInterfaces load balanced traffic is routed to.
This object should not be visible to the customer but is available in the customer namespace

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

