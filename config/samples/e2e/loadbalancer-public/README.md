# `LoadBalancer` of `type: Public`

This example deploys a `LoadBalancer` of `type: Public` by selecting two target `NetworkInterface`s via a `networkInterfaceSelector` `metav1.LabelSelector`. These two `NetworkInterface`s are referencing to same `Network` and will be attached to two different `Machine`s.

The following artifacts will be deployed in your namespace:   
- IronCore `LoadBalancer`
- IronCore `Network`, `NetworkInterface`s and `VirtualIP`s
- IronCore `Machine`s 
- IronCore `Volume`s
- Secret containing the `ignition`  

## Prerequisites

- [Butane](https://coreos.github.io/butane/)

## Usage
1. Adapt the `namespace` in `kustomization.yaml`
2. Replace `your-user`, `your-pw-hash` and `your-ssh-key` in the `ignition/ignition.yaml`
3. Run `ignition/regenerate-ignition.sh`
4. Create the below `patch-machine.yaml` in `loadbalancer-public` folder with the desired `machineClassRef` and `machinePoolRef` as per your environment

```
apiVersion: compute.ironcore.dev/v1alpha1
kind: Machine
metadata:
  name: machine-sample
spec:
  machineClassRef:
    name: new-machineClass   # The new name of the machine class reference
  machinePoolRef:
    name: new-machinePool
```

5. Create the below`patch-volume.yaml`in `loadbalancer-public` folder with the desired `volumeClassRef`and `volumePoolRef` as per your environment

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: Volume
metadata:
  name: volume-sample
spec:
  volumeClassRef:
    name: new-volumeClass    # The new name of the volume class reference
  image: new-image:rootfs
  volumePoolRef:
    name: new-volumePool
```
6. Update the `kustomization.yaml` with below content
```
patches:
- path: patch-machine.yaml
- path: patch-volume.yaml
```

7. Run (`kubectl apply -k ./`)