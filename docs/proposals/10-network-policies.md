---
title: Network Policies

iep-number: 10

creation-date: 2023-04-13

status: implementable

authors:

- "@adracus"

reviewers:

- "@afritzler"
- "@lukasfrank"

---

# IEP-10: Network Policies

## Table of Contents

- [IEP-10: Network Policies](#IEP-10-network-policies)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
  - [Alternatives](#alternatives)

## Summary

In an unregulated network, it is impossible to properly enforce the rules
of least-privilege. Each member of a network could potentially communicate
to each other, receive traffic from the public internet (if connected) and
communicate with the public internet (if connected). This imposes a big security
risk which has to be properly addressed in each modern infrastructure.
This proposal describes how to introduce network policies as a means to regulate
traffic inside a network building upon the existing concepts that were proposed in the
[Networking Integration OEP](01-networking-integration.md).

## Motivation

Currently, there is no way to describe which members of a network should be able
to communicate with each other. Same applies to traffic coming from the public
internet / going to the public internet. The `ironcore` should be extended
with traffic control mechanisms, allowing to limit / deny traffic on a
per-instance basis. Of course, the mechanisms should align well with existing
proposals / concepts in the Kubernetes world.

### Goals

* Be able to deny ingress and egress traffic between members of a `Network`.
* Be able to deny ingress and egress traffic between members of a `Network` and
  the public internet.

### Non-Goals

* Define policies that apply to multiple `Network`s simultaneously.

## Proposal

Introduce a new type `NetworkPolicy` that regulates traffic within a certain network.
By default, traffic to members in a `Network` is not regulated. However, as soon as a
`NetworkPolicy` selects members of a `Network`, all ingress and egress traffic
concerning the members is denied unless a `NetworkPolicy` explicitly allows it.
This makes it so multiple `NetworkPolicy` instances can never be in conflict and
instead just allow more ingress / egress traffic to be received / sent.

Members are selected using a Kubernetes `metav1.LabelSelector` to allow specifying
multiple target network interfaces.

A `NetworkPolicy` specifies rules to treat `ingress` and `egress` traffic. To be
able to express whether e.g. no `ingress` or `egress` traffic is allowed without specifying
any rule, `NetworkPolicy`s always have to specify the `policyTypes` (either `Ingress` / `Egress`)
they want to enforce.

Example manifest:

```yaml
apiVersion: networking.ironcore.dev/v1alpha1
kind: NetworkPolicy
metadata:
  namespace: default
  name: my-network-policy
spec:
  # This specifies the target network to limit the traffic in.
  networkRef:
    name: my-network
  # Only network interfaces in the specified network will be selected.
  networkInterfaceSelector:
    matchLabels:
      app: db
  # If the policy types are not specified, they are inferred on whether
  # any ingress / egress rule exists. If no ingress / egress rule exists,
  # the network policy is denied on admission.
  policyTypes:
  - Ingress
  - Egress
  # Multiple ingress / egress rules are possible.
  ingress:
  - from:
    # Traffic can be limited from a source IP block.
    - ipBlock:
        cidr: 172.17.0.0/16
    # Traffic can also be limited to objects of the networking api.
    # For instance, to limit traffic from network interfaces, one could
    # specify the following:
    - objectSelector:
        kind: NetworkInterface
        matchLabels:
          app: web
    # Analogous to network interfaces, it is also possible to limit
    # traffic coming from load balancers:
    - objectSelector:
        kind: LoadBalancer
        matchLabels:
          app: web
    # Ports always have to be specified. Only traffic matching the ports
    # will be allowed.
    ports:
    - protocol: TCP
      port: 5432
  egress:
  - to:
    - ipBlock:
        cidr: 10.0.0.0/24
    ports:
    - protocol: TCP
      port: 8080
```

## Alternatives

* Provide an own networking overlay that enforces the rules. However, this
  involves significant effort and maintenance.
* Have hypervisors / management processes monitor traffic sent by application
  processes. However, monitoring of the hypervisors / management processes still
  is not addressed with this.
