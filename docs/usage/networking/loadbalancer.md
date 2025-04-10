# LoadBalancer

A `LoadBalancer` resource is an L3(IP-based) load balancer service implementation provided by Ironcore. It provides an externally accessible IP address that sends traffic to the correct port on your cluster nodes. Ironcore LoadBalancer allows targeting multiple `NetworkInterfaces` and distributes traffic between them. This LoadBalancer supports dual stack IP addresses (IPv4/IPv6). 

## Example Network Resource
An example of how to define a `LoadBalancer` resource in `Ironcore`
```
apiVersion: networking.ironcore.dev/v1alpha1
kind: LoadBalancer
metadata:
  namespace: default
  name: loadbalancer-sample
spec:
  type: Public
  ipFamilies: [IPv4]
  networkRef:
    name: network-sample
  networkInterfaceSelector:
    matchLabels:
      app: web
  ports:
  - port: 80

```
(`Note`: Refer to <a href="https://github.com/ironcore-dev/ironcore/tree/main/config/samples/e2e/loadbalancer-public">E2E Examples</a> for more detailed examples.)

# Key Fields:
- `type`(`string`): The type of `LoadBalancer`. Currently two types of `Loadbalancer` are supported: 
    - `Public`: LoadBalancer that allocates public IP and routes a stable public IP.
    - `Internal`: LoadBalancer that allocates and routes network-internal, stable IPs.
- `ipFamilies`(`list`): ipFamilies are the IP families the LoadBalancer should have(Supported values are `IPv4` and `IPv6`). 
- `ips`(`list`): The ips are the list of IPs to use. This can only be used when the type is LoadBalancerTypeInternal.
- `networkRef`(`string`): networkRef is the Network this LoadBalancer should belong to.
- `networkInterfaceSelector`(`labelSelector`): networkInterfaceSelector defines the NetworkInterfaces for which this LoadBalancer should be applied
- `ports`(`list`): ports are the list of LoadBalancer ports should allow
    - `protocol`(`string`): protocol is the protocol the load balancer should allow. Supported protocols are `UDP`, `TCP`, and `SCTP`, if not specified defaults to TCP.
    - `port`(`int`): port is the port to allow.
    - `endPort`(`int`): endPort marks the end of the port range to allow. If unspecified, only a single port `port` will be allowed.

# Reconciliation Process:

- **NetworkInterfaces selection**: LoadBalancerController continuously watches for `LoadBalancer` resources and reconciles. LoadBalancer resource dynamically selects multiple target `NetworkInterfaces` via a `networkInterfaceSelector` LabelSelector from the spec. Once the referenced Network is in `Available` state, the Loadbalancer destination IP list and referencing `NetworkInterface` is prepared by iterating over selected NetworkIntrefaces status information.

- **Preparing Routing State Object**: Once the destination list is available `LoadBalancerRouting` resource is created. `LoadBalancerRouting` describes `NetworkInterfaces` load balanced traffic is routed to. This object describes the state of the LoadBalancer and results of the LoadBalancer definition specifically `networkInterfaceSelector` and `networkRef`. 
Later this information is used at the Ironcore API level to describe the explicit targets in a pool traffic is routed to.

Sample `LoadBalancerRouting` object(`Note`: it is created by LoadBalancerController)
```
apiVersion: networking.ironcore.dev/v1alpha1
kind: LoadBalancerRouting
metadata:
  namespace: default
  name: loadbalancer-sample # Same name as the load balancer it originates from.
# networkRef references the exact network object the routing belongs to.
networkRef:
  name: network-sample
# destinations list the target network interface instances (including UID) for load balancing.
destinations:
  - name: my-machine-interface-1
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73a
  - name: my-machine-interface-2
    uid: 2020dcf9-e030-427e-b0fc-4fec2016e73d
```
**LoadBalancer status update**: The `LoadBalancerController` in ironcore-net takes care of allocating IPs for defined `ipFamilies` in the spec and updates them in its `status.ips`.