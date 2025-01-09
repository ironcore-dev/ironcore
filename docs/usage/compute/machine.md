# Machine

A `Machine` resource in `IronCore` is used to represent a compute resource or a virtual machine. 
It serves as a means to configure network, storage, type of machine and other information needed to create a VM. The `MachineController` reconciler leverages this information to determine where the machine needs to be created and the type of machine that needs to be created along with the required `Network` and `Storage` configuration which will be further passed to respective `NetworkController` and `StorageController`.

## Example Machine Resource

An example of how to define a Machine resource:

```yaml
apiVersion: compute.ironcore.dev/v1alpha1
kind: Machine
metadata:
  name: machine-sample
spec:
  machineClassRef:
    name: machineclass-sample
  #  machinePoolRef:
  #    name: default
  image: my-image
  volumes:
    - name: rootdisk # first disk is the root disk
      # priority: 0 # optional
      volumeRef:
        name: my-volume
  networkInterfaces:
    - name: primary
      networkInterfaceRef:
        name: networkinterface-sample
  ignitionRef:
    name: my-ignition-secret
```
(`Note`: Refer to <a href="https://github.com/ironcore-dev/ironcore/tree/main/config/samples/e2e">E2E Examples</a> for more detailed examples.)

**Key Fields**:

- machineClassRef (`string`): MachineClassRef is a reference to the machine class/flavor of the machine.
- machinePoolRef (`string`): MachinePoolRef defines the machine pool to run the machine in. If empty, a scheduler will figure out an appropriate pool to run the machine in.
- image (`string`): Image is the optional URL providing the operating system image of the machine.
- volumes (`list`): Volumes are list volumes(Storage) attached to this machine.
- networkInterfaces (`list`): NetworkInterfaces define a list of network interfaces present on the machine
- ignitionRef (`string`): IgnitionRef is a reference to a secret containing the ignition YAML for the machine to boot up. If a key is empty, DefaultIgnitionKey will be used as a fallback.


## Reconciliation Process

1. **Machine Scheduling**: 
The MachineScheduler controller continuously watches for machines without an assigned MachinePool and tries to schedule it on available and matching MachinePool.
  - **Monitor Unassigned Machines**: The scheduler continuously watches for machines without an assigned `machinePoolRef`.
  - **Retrieve Available Machine Pools**: The scheduler fetches the list of available machine pools from the cache.
  - **Make Scheduling Decisions**: The scheduler selects the most suitable machine pool based on resource availability and other policies.
  - **Update Cache**: The scheduler updates the cache with recalculated allocatable `machineClass` quantities.
  - **Assign MachinePoolRef**: The scheduler assigns the selected `machinePoolRef` to the machine object.

2. **IRI Machine creation**: Once the Machine is allocated to a particular pool, the `MachineController` processes the `Machine` resource and it extracts the `IgnitionData`, `NetworkInterface` and `Volume` information from the `spec` and prepares IRI machine resource.

3. **Machine Brokering**: Once IRIMachine object is prepared create/update the machine request is sent to a broker via the IRI interface(via GRPC call). An actual VM is created when the request reaches the compute provider. Once the response is received from IRI call Machine status is updated with the status recieved.

4. **Network Interface handling**: MachineControllerNetworkinterface takes care of attaching/detaching Network interfaces defined for the machine. Once the attachment is successful status is updated from `Pending` to `Attached`.

5. **Volume handling**: MachineControllerVolume takes care of attach/detach of Volumes(Storage) defined for machine. Once the attachment is successful status is updated from `Pending` to `Attached`.

6. **Ephemeral resource handling**: If `NetworkIntreface` or `Volume` is defined as ephemeral(i.e. coupled to the lifetime of the machine object) in the machine spec, MachineEphemeralControllers takes care of creating and destroying respective objects on creation/deletion of the machine. 

## Lifecycle and States

A Machine can be in the following states:
1. **Pending**:  A Machine is in a `Pending`  state when the Machine has been accepted by the system, but not yet completely started. This includes time before being bound to a MachinePool, as well as time spent setting up the Machine on that MachinePool. 
2. **Running**: A Machine in Running state when the machine is running on a MachinePool.
2. **Shutdown**: A Machine is in a Shutdown state.
3. **Terminating**: A Machine is Terminating.
2. **Terminated**: A Machine is in the Terminated state when the machine has been permanently stopped and cannot be started.