# VirtualIP
A `Virtual IP (VIP)` in  the `Ironcore` ecosystem is an abstract network resource representing an IP address that is dynamically associated with an `ironcore` `networkInterface`, which in turn is linked to an `ironcore machine/vm`.

# Examaple VirtualIP Resource

```
apiVersion: networking.ironcore.dev/v1alpha1
kind: VirtualIP
metadata:
  name: virtualip-sample
spec:
  type: Public
  ipFamily: IPv4
#status:
#  ip: 10.0.0.1 # This will be populated by the corresponding controller.
```

# Key Fields
- **type**(`string`):  Currently supported type is `public`, which allocates and routes a stable public IP.

- **ipFamily**(`string`): `IPFamily` is the ip family of the VirtualIP. Supported values for IPFamily are `IPv4` or `IPv6`.


# Reconciliation Process:

- **VirtualIP Creation**: 
A VirtualIP resource is created, specifying attributes like `ipFamily`: IPv4 or IPv6 and `Type`: public 

- **Reconciliation and IP Assignment**: 
The VirtualIP reconciler
Creates or updates a corresponding APINet IP in Ironcore's APINet system.
Ensures the IP is dynamically allocated and made available for use.

- **Error Handling**:
If the creation or update of the APINet IP fails, update the VirtualIP status to indicate it is unallocated.
Requeue the reconciliation to retry the operation.

- **Synchronize Status**:
Update the VirtualIP status to reflect the actual state of the APINet IP.
If successfully allocated, update the status with the assigned IP address.

- **Networking Configuration**: 
    - VM Integration: The allocated VirtualIP is associated with the VM through network configuration mechanisms
    - Load Balancer Integration: If a load balancer is used, the virtualIP is configured as the frontend IP to route requests to the VM.

