# MachineClass

A `MachineClass` is an `IronCore` resource used to represent a class/flavor of a Machine. It serves as a means to define the number of resources a `Machine` object can have as capabilities(For eg, CPU, memory) associated with a particular class. The `MachineClassController` reconciler leverages this information to create `MachineClass`.

## Example Machine Resource

An example of how to define a MachineClass resource:

```yaml
apiVersion: compute.ironcore.dev/v1alpha1
kind: MachineClass
metadata:
  name: machineclass-sample
capabilities:
  cpu: 4
  memory: 16Gi
```

**Key Fields**:

- capabilities (`ResourceList`): capabilities are used to define a list of resources a Machine can have along with its capacity.


## Reconciliation Process

- **MachineClass Creation**: The `MachineClassController` uses the `capabilities` field in the MachineClass resource to create a flavor of MachineClass resource.
- **MachineClass Deletion**: Before deleting any MachineClass it's been ensured that it is not in use by any `Machine` and then only deleted.

