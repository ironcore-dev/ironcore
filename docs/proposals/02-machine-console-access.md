---
title: Machine Console Access

iep-number: 2

creation-date: 2022-12-05

status: implementable|implemented

authors:

- "@adracus"

reviewers:

- "@gehoern"
- "@afritzler"
- "@Gchbg"

---

# IEP-02: Machine Console Access

## Table of Contents

- [IEP-02: Machine Console Access](#IEP-02-machine-console-access)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
    - [User-facing API](#user-facing-api)
    - [Server-Side API](#server-side-api)
  - [Alternatives](#alternatives)

## Summary

A user of the ironcoreshould be able to access the serial console of their machine
to access / debug / run imperative commands on it. For this, an endpoint + client-side
tooling has to be created as well as the server-side machinery.

## Motivation

Users should be able to access / debug / run imperative commands on their machines.
This gives immediate feedback on the state of the machine and is a required feature for
our minimum viable product.

### Goals

* Define an endpoint + client side tooling for machine console access
* Define a server-side interface machine pool providers have to implement in order to
  support console access to their machines.

### Non-Goals

* Due to the imperative nature of consoles, no declarative interface to consoles should be defined.
* Have consoles as a building piece of other parts of the ironcore.

## Proposal

### User-facing API

The `compute.ironcore.dev/Machine` resource is extended with an `exec` subresource. When connecting
to that subresource, a websocket connection to the backing machine console should be opened.
Supported HTTP methods for the `exec` call are `POST` and `GET` (in order to be able to do this from a browser
as well).

Example call to the Kubernetes API server hosting the aggregated API:

```http request
GET https://<address>/apis/compute.ironcore.dev/v1alpha1/namespaces/<namespace>/machines/<machine-name>/exec
```

### Server-Side API

Once the server receives such a request, it gets the `Machine` and looks up the `MachinePool` the `Machine`
is running on. If the `Machine` does not exist or is not scheduled onto a `MachinePool`, an error is returned.

After identifying the responsible `MachinePool`, it is retrieved and its `.status.addresses` field is inspected
for an address to call. The `.status.addresses` field does not exist yet and has to be updated in the ironcore.
It is the responsibility of the `MachinePool` implementor to report its endpoints in the `status`.

For reference on the address type, see
[the Kubernetes node address type](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#nodeaddress-v1-core)
which will be used as reference for designing the address type.

Example manifest:

```yaml
apiVersion: compute.ironcore.dev/v1alpha1
kind: MachinePool
metadata:
  name: my-machine-pool
spec:
  providerID: my://machine-pool
status:
  addresses:
    - address: 10.250.0.38
      type: InternalIP
    - address: my-machine-pool-host
      type: ExternalDNS
```

Once an address has been identified, the ironcore API server calls the endpoint of the `MachinePool` provider
with an `exec` request for the `Machine`. The resulting websocket connection is proxied through the ironcore API
server to the user.

```http request
GET https://<machine-pool-adddress>/apis/compute.ironcore.dev/namespaces/<namespace>/machines/<machine>/exec
```

> Caution: This proposal does *not* include anything on authentication mechanisms yet. Implementors can already
> implement the endpoint but authentication will be added in the future.

## Alternatives
