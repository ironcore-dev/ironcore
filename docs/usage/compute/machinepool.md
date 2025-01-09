# MachinePool

A `MachinePool` is a resource in `Ironcore` that represents a pool of compute resources managed collectively. It defines the infrastructure's compute configuration used to provision and manage `Machines`, ensuring resource availability and compatibility with associated `MachineClasses`.

## Example MachinePool Resource

An example of how to define a MachinePool resource:

```yaml
apiVersion: compute.ironcore.dev/v1alpha1
kind: MachinePool
metadata:
  name: machinepool-sample
  labels:
    ironcore.dev/az: az1
spec:
  providerID: ironcore://shared
```

**Key Fields**:

- `ProviderID`(`string`):  The `providerId` helps the controller identify and communicate with the correct compute system within the specific backend compute provider.
For example `ironcore://shared`

## Reconciliation Process

- **Machine Type Discovery**: It constantly checks what kinds of `MachineClasses` are available in the `Ironcore` Infrastructure
- **Compatibility Check**: Evaluating whether the `MachinePool` can manage available machine classes based on its capabilities. 
- **Status Update**: Updating the MachinePool's status to indicate the supported `MachineClasses` with available capacity and allocatable.
- **Event Handling**: Watches for changes in MachineClass resources and ensures the associated MachinePool is reconciled when relevant changes occur.



