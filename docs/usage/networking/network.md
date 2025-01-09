# Network

A `Network` resource in `ironcore` refers to a logically isolated network. 
This further allows you to fully control your networking environment, including resource placement, connectivity, peering and security. 
The `NetworkController` reconciler leverages this information to create a Network in ironcore infrastructure.
`Machine` type is provided with provision to integrate with the Network via `NetworkInterface`.

## Example Network Resource
An example of how to define a `Network` resource in `ironcore`
```
apiVersion: networking.ironcore.dev/v1alpha1
kind: Network
metadata:
  name: network-sample
spec:
  peerings:
  - name: peering1
    networkRef:
      name: network-sample2
```

# Key Fields:
- `providerID`(`string`): ProviderID is the provider-internal ID of the network.
- `peerings`(`list`): Peerings are the list of network peerings with this network(Optional).
- `incomingPeerings`(`list`): incomingPeerings is a list of PeeringClaimRefs which is nothing but peering claim references of other networks.

# Reconciliation Process:

- **Network creation**: `ironcore-net` which is the network provider for ironcore realizes the `Network` resource and updates 
providerID in the spec. Once resource is in available state status is marked to `Available`.

- **Network peering process**: Network peering is a technique used to interleave two isolated networks, allowing members of both networks to communicate with each 
other as if they were in the same networking domain,  `NetworkPeeringController` facilitates this process.
  - Information related to the referenced `Network` to be paired with is retrieved from the `peering` part of the spec.
  - Validation is done to see if both Networks have specified a matching peering item (i.e. reference each other via networkRef) to mutually accept the peering.
  - The (binding) phase of a spec.peerings item is reflected in a corresponding `status.peerings` item with the same name. 
    The phase can either be `Pending`, meaning there is no active peering or `Bound` meaning the peering as described in the `spec.peerings` item is in place. 

- **Network Release Controller**: `NetworkReleaseController` continuously checks if claiming Networks in other Network's peerings section still exist if not present it will be removed from the `incomingPeerings` list.