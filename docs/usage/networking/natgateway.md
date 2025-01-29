# NATGateway
In the `Ironcore` project, a `NATGateway` (Network Address Translation Gateway) facilitates outbound internet connectivity in private subnets, ensuring that instances in private subnets can access external services without exposing them to unauthorized inbound traffic.

It is a critical network service that provides secure and controlled internet access for private resources in the `Ironcore` infrastructure. It is enforced by the underlying `Ironcore's` network plugin called <a href="https://github.com/ironcore-dev/ironcore-net/blob/main/apinetlet/controllers/natgateway_controller.go"> ironcore-net </a>

# Example NATGateway Resource
An example of how to define a `NATGateway` resource in `Ironcore`

```
apiVersion: networking.ironcore.dev/v1alpha1
kind: NATGateway
metadata:
  namespace: default
  name: natgateway-sample
spec:
  type: Public
  ipFamily: IPv4
  portsPerNetworkInterface: 64
  networkRef:
    name: network-sample
```

# Key Fields 
- `type`(`string`): This represents a NATGateway type that allocates and routes a stable public IP. The supported value for type is `public`

- `ipFamily`(`string`): `IPFamily` is the IP family of the `NATGateway`. Supported values for IPFamily are `IPv4` and `IPv6`.

- `portsPerNetworkInterface`(`int32`): This Specifies the number of ports allocated per network interface and controls how many simultaneous connections can be handled per interface. 

    If empty, 2048 (DefaultPortsPerNetworkInterface) is the default.

- `networkRef`(`string`): It represents which network this `NATGateway` serves.

# Example Use Case:
Imagine you have a private server in a private subnet without a public IP. It needs to download software updates from the internet. Instead of giving it direct internet access (which compromises security), the server sends its requests through the NAT Gateway. The NAT Gateway fetches the updates and returns them to the server while keeping the server's private IP hidden from the external world.

# Reconciliation Process:

- **Fetch NATGateway Resource**: It fetches the current state of `NATGateways`, Based on user specifications the desired state of `NATGateway` is determined. This includes the number of NAT Gateways, their types, associated subnets, and routing configurations.

- **Compare and Reconcile**: The reconciler keeps monitoring the state of NAT Gateways to detect any changes or drifts from the desired state, triggering the reconciliation process as needed.
    - Creation: If a NAT Gateway specified in the desired state does not exist in the current state, it is created. For instance, creating a public NAT Gateway in a public subnet to provide internet access to instances in private subnets.

    - Update: If a NAT Gateway exists but its configuration differs from the desired state, it is updated accordingly. This may involve modifying routing tables or changing associated Elastic IPs.

    - Deletion: If a NAT Gateway exists in the current state but is not present in the desired state, it is deleted to prevent unnecessary resource utilization.

- **Error Handling and Logging**: Throughout the reconciliation process, any errors encountered are logged, schedule retries as necessary to ensure eventual consistency.

- **Update Status**:
After reconciling all `NATGateways`, log the successful reconciliation and update the `NATGateways` status with the corresponding values for `ips`as shown below.

```
status:
  ips:
  - name: ip1
    ip: 10.0.0.1
```
