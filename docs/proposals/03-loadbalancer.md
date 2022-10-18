---
title: Network Loadbalancer

oep-number: 3

creation-date: 2022-10-18

status: implementable

authors:

- "@gehoern"
  reviewers:
- #"@MalteJ"
- #"@adracus"
- #"@afritzler"

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
Network service Load Balancing is needed in every modern infrastructure (e.g. Kubernetes). Load balancing makes services scalable, fault-tolerant and accessible from external users. The load balancer described here is a L3-Loadbalancer (so it implements an IP level load balancing). The load balancer function is a standard network function that can be to be deployed (as an implementation detail) on a platform/node that is capable of running network functions (e.g. dp-service enabled nodes/nic's) 

## Motivation
The onMetal network is a fully routed network. To run workloads hosted on several nodes (in Kubernetes terms equivalent to cluster IP's) and not only on a single node (node IP's) solutions like ECMP (Equal cost multi pathing) seem to be the natural solution. But ECMP has certain limitations: the amount of nodes it can be used to distribute to and a hash based algorithm to pin the network flows, that changes - based on the amount of targets - its distributed pinning. 
To avoid those limitations and to have better control over the load balancer behavior an onMetal load balancer service needs to be provided.

### Goals
- define a user facing API (onMetal api) for the load balancer
- enable the user to request an L3 Loadbalancer (only IP addresses) with load balancer address (outside facing) and 
- with multiple target's based on the onMetal ([OEP-1](01-networking-integration.md)) network interface architecture
- the targets of the load balancer must be dynamic changeable
- the load balancer needs to avoid forwarding unintended traffic.
  - based on ports/protocols it can be filtered what needs to be load balanced
  - icmp requests are not part of forwarded traffic and will be filtered by the load balancer anyhow
- Load balancer (ip) and targets can be parts of different networks (VNI's) but
- all targets must be of the same network (VNI)

### Non-Goals
- no address or port translation / rewriting (no SNAT / DNAT) (L4 Loadbalancer)
- no injection of additional information (e.g. x-forwarded-for) (L7 Loadbalancer)
- no protocol offloading like ssl (L7 Loadbalancer)
- if multiple load balancer IP's are needed, request multiple load balancers (a load balancer services one load balancer IP)

### Details
- load balancing is used to deliver a packet addressed to the load balancer to one of its targets via the onMetal network routing
- the load balancing will be for the target and the source transparent
- the target needs to know the load balancer's IP and needs to answer with it
- answers to request will be direct delivered since all details are known to the target
- payload packages are not changed, to not lose any information


## Proposal
A network loadbalancer CRD allows the user to define the network function. all entries except the `targetNetworkInterface` must be immutable. 

```yaml
apiVersion: networking.onmetal.de
kind: NetworkLoadBalancer
metadata: 
  name: myLoadBalancer
  namespace: customer-1
spec:
  # same behavior as VirtualIP from OEP-1
  # should also provide non public IP
  loadBalancerIP:
    ephemeral:
      virtualIPTemplate:
        spec:
          type: Public
          ipFamily: IPv4
  ports:
    - name: webserver
      protocol: tcp
      port: 80
    - name: db
      protocol: udp
      port: 1024
      portEnd: 2048
  targetNetworkInterface:
    - name: myDbMachineInterface
    - labelSelector:
        key: db          
status:
  ip: 45.86.152.88
  phase: Bound
  targetNetworkInterfaces:
  - name: myDbMachineInterface
  - name: autoSelectedMachineInterface
  - name: autoSelectedMachineInterface-2
```

