# NetworkInterface
A `NetworkInterface` resource in `Ironcore` represents a connection point between a virtual machine(VM) and a virtual network. It encapsulates the configuration and life cycle management of the virtual network interface, ensuring seamless connectivity for VMs.

The `MachineEphemeralNetworkInterfaceReconciler` is responsible for managing the lifecycle of ephemeral network interfaces associated with machines. Its primary function is to ensure that the actual state of these network interfaces aligns with the desired state specified in each machine's configuration.

# Example NetworkPolicy Resource
An example of how to define a `NetworkInterface` resource in `Ironcore`

```
apiVersion: networking.ironcore.dev/v1alpha1
kind: NetworkInterface
metadata:
  name: networkinterface-sample
spec:
  networkRef:
    name: network-sample
  ipFamilies:
    - IPv4
  ips:
    - value: 10.0.0.1 # internal IP
  virtualIP:
    virtualIPRef:
      name: virtualip-sample
```
**Note**: For a detailed end-to-end example to understand the ephemeral and non `NetworkInterface`, please refer <a href="https://github.com/ironcore-dev/ironcore/tree/main/config/samples/e2e">E2E Examples</a>

# Key Fields 
- **networkRef**: `NetworkRef` is the Network this NetworkInterface is connected to

- **ipFamilies**: `IPFamilies` defines the list of IPFamilies this `NetworkInterface` supports. For eg: `IPV4` and `IPV6`

- **ips**: `IPs` are the list of provided internal IPs which should be assigned to this NetworkInterface

- **virtualIP**: `VirtualIP` specifies the public ip that should be assigned to this NetworkInterface.

# Reconciliation Process:

- **Fetch Machine Resource**:
Retrieve the specified Machine resource from the reconciliation request.
If the Machine is marked for deletion (indicated by a non-zero DeletionTimestamp), exit the process without further action.

- **Generate Desired Ephemeral Network Interfaces**:
Analyze the Machine's specification to identify the desired ephemeral NetworkInterface resources.
Construct a map detailing these desired NetworkInterfaces, including their configurations and expected states.

- **Fetch Existing Network Interfaces**:
List all existing NetworkInterface resources within the same namespace as the Machine.

- **Compare and Reconcile**:
    - For each existing Network Interface:
Determine if it is managed by the current Machine and whether it matches the desired state.
    - If unmanaged but should be managed, avoid adopting it to prevent conflicts.
    - For each desired Network Interface not present:
Create the missing `NetworkInterface` and establish the Machine as its controller.

- **Handle Errors**:
Collect any errors encountered during the reconciliation of individual NetworkInterfaces.
Log these errors and schedule retries as necessary to ensure eventual consistency.

- **Update Status**:
After reconciling all NetworkInterfaces, log the successful reconciliation and update the `NetworkInterface` status with the corresponding values for `ips`, `state`, and `virtualIP`, as shown below.

```
status:
  ips:
  - 10.0.0.1
  lastStateTransitionTime: "2025-01-13T11:39:17Z"
  state: Available
  virtualIP: 172.89.244.23
```
The `state` is updated as one of the following lifecycle states based on the reconciliation result
- **Pending**
- **Available**
- **Error**


