# NetworkPolicy
In `Ironcore`, NetworkPolicies are implemented based on the standard Kubernetes `NetworkPolicy` approach, which is enforced by the underlying `Ironcore's` network plugin <a href="https://github.com/ironcore-dev/ironcore-net/blob/main/apinetlet/controllers/networkpolicy_controller.go"> ironcore-net </a> and other components. These policies use label selectors to define the source and destination of allowed traffic within the same network and can specify rules for both ingress (incoming) and egress (outgoing) traffic. 

In the `Ironcore` ecosystem, the `NetworkPolicy` has the following characteristics:

- NetworkPolicy is applied exclusively to NetworkInterfaces selected using label selectors.

- These NetworkInterfaces must belong to the same network.

- The policy governs traffic to and from other `NetworkInterfaces`, `LoadBalancers`, etc., based on the rules defined in the NetworkPolicy.

# Example NetworkPolicy Resource
An example of how to define a `NetworkPolicy` resource in `Ironcore`

```
apiVersion: networking.ironcore.dev/v1alpha1
kind: NetworkPolicy
metadata:
  namespace: default
  name: my-network-policy
spec:
  networkRef:
    name: my-network
  networkInterfaceSelector:
    matchLabels:
      app: db
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - ipBlock:
        cidr: 172.17.0.0/16
    - objectSelector:
        kind: NetworkInterface
        matchLabels:
          app: web
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
https://github.com/ironcore-dev/ironcore/tree/main/config/samples/e2e/network-policy

(`Note`: Refer to <a href="https://github.com/ironcore-dev/ironcore/tree/main/config/samples/e2e/network-policy">E2E Examples</a> for more detailed example on networkpolicy to understant e2e flow)

# Key Fields

- `networkRef`(`string`): NetworkRef is the Network to regulate using this NetworkPolicy.

- `networkInterfaceSelector`(`labelSelector`): NetworkInterfaceSelector defines the target `NetworkInterfaces` for which this `NetworkPolicy` should be applied.

- `policyTypes`(`list`): There are two supported policyTypes `Ingress` and `Egress`.

- `ingress`(`list`): ingress defines the list of `NetworkPolicyIngressRules`. Each NetworkPolicy may include a list of allowed `ingress` rules. Each rule allows traffic that matches both the `from` and `ports` sections. The example policy contains a single rule, which matches traffic on a single port, from one of three sources, the first specified via an ipBlock, the second and third via different objectSelector.

- `egress`(`list`): egress defines the list of `NetworkPolicyEgressRules`. Each NetworkPolicy may include a list of allowed egress rules. Each rule allows traffic that matches both `to` and `ports` sections. The example policy contains a single rule, which matches traffic on a single port to any destination in 10.0.0.0/24.

# Reconciliation Process:
The `NetworkPolicyReconciler` in the Ironcore project is responsible for managing the lifecycle of `NetworkPolicy` resources. Its primary function is to ensure that the rules specified by the user in the NetworkPolicy resource are enforced and applied on the target `NetworkInterface`.

The <a href="https://github.com/ironcore-dev/ironcore-net/blob/main/apinetlet/controllers/networkpolicy_controller.go"> apinetlet </a> component in `ironcore-net` plugin is responsible for translating the policy rules into another APInet type resource `NetworkPolicyRule`. Finally, the <a href="https://github.com/ironcore-dev/ironcore-net/blob/main/metalnetlet/controllers/networkinterface_controller.go"> metalnetlet </a> component in `ironcore-net` and other components propagates these rules for enforcement at `dpservice` level in the Ironcore infrastucture.

The reconciliation process involves several key steps:

- **Fetching the NetworkPolicy Resource**: The reconciler retrieves the NetworkPolicy resource specified in the reconciliation request. If the resource is not found, it may have been deleted, and the reconciler will handle this scenario appropriately.

- **Validating the NetworkPolicy**: The retrieved NetworkPolicy is validated to ensure it confirms the expected specifications. This includes checking fields such as NetworkRef, NetworkInterfaceSelector, Ingress, Egress, and PolicyTypes to ensure they are correctly defined.

- **Fetching Associated Network Interfaces**: Using the NetworkInterfaceSelector, the reconciler identifies the network interfaces that are subject to the policy.

- **Applying Policy Rules**: The reconciler translates the ingress and egress rules defined in the NetworkPolicy into configurations that can be enforced by the underlying network infrastructure. This involves interacting with other components responsible for NetworkPolicy or Firewall rule enforcement.

- **Handling Errors and Reconciliation Loops**: If errors occur during any of the above steps, the reconciler will log the issues and may retry the reconciliation. 

