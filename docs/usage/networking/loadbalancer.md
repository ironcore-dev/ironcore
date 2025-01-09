# LoadBalancer

A `LoadBalancer` resource is an L3(IP-based) load balancer service implementation provided by Ironcore. It provides an externally accessible IP address that sends traffic to the correct port on your cluster nodes. Ironcore LoadBalancer allows targeting multiple `NetworkInterfaces` and distributes traffic between them. This Load Balancer supports dual stack IP addresses (IPv4/IPv6). 

## Example Network Resource
An example of how to define a `LoadBalancer` resource in `ironcore`
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
#status:
#  ips:
#  - 10.0.0.1 # The publicly available IP of the load balancer

```
(`Note`: Refer to <a href="https://github.com/ironcore-dev/ironcore/tree/main/config/samples/e2e/loadbalancer-public">E2E Examples</a> for more detailed examples.)

# Key Fields:
- `type`(`string`): Type is the type of LoadBalancer. Currently two types of Loadbalancer are supported: 
    - `Public`: LoadBalancer that allocates public IP and routes a stable public IP.
    - `Internal`: LoadBalancer that allocates and routes network-internal, stable IPs.
- `ipFamilies`(`list`): ipFamiliesare the IP families the load balancer should have(Supported values are `IPv4` and `IPv6`). 
- `ips`(`list`): The IPs are the list of IPs to use. This can only be used when the Type is LoadBalancerTypeInternal.
- `networkRef`(`string`): NetworkRef is the Network this LoadBalancer should belong to.
- `networkInterfaceSelector`(`labelSelector`): NetworkInterfaceSelector defines the NetworkInterfaces for which this LoadBalancer should be applied
- `ports`(`list`): Ports are the list of list of loadbalancer ports should allow
    - `protocol`(`string`): Protocol is the protocol the load balancer should allow. Supported protocols are `UDP`, `TCP`, and `SCTP`, if not specified defaults to TCP.
    - `port`(`int`): Port is the port to allow.
    - `endPort`(`int`): endPort marks the end of the port range to allow. If unspecified, only a single port `port` will be allowed.

# Reconciliation Process:

- **NetworkInterfaces selection**: LoadBalancerController continuously watches for LoadBalancer resources and reconciles. LoadBalancer resource dynamically selects multiple target `NetworkInterfaces` via a networkInterfaceSelector LabelSelector from the spec. Once the referenced Network is in `Available` state, the Loadbalancer destination IP list and referencing NetworkInterface is prepared by iterating over selected NetworkIntrefaces status information.

- **Preparing Routing State Object**: Once the destination list is available `LoadBalancerRouting` resource is created. `LoadBalancerRouting` describes NetworkInterfaces load balanced traffic is routed to. This object describes the state of the LoadBalancer and results of the LoadBalancer definition specifically networkInterfaceSelector and networkRef. 
Later this information is used at the ironcore API level to describe the explicit targets in a pool traffic is routed to.

Sample LoadBalancerRouting object(`Note`: it is created by LoadBalancerController)
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
