# Network Peering Sample deployement

This example deploys two `Network`s peered with each other and these `Network`s are referenced in `Machine`s.
The following artifacts will be deployed in your namespace:   
- IronCore `Network`, `NetworkInterface` and `VirtualIP`
- IronCore `Machine` 
- IronCore `Volume`
- Secret containing the `ignition`

## Prerequisites

- [Butane](https://coreos.github.io/butane/)

## Usage
1. Adapt the `namespace` in `kustomization.yaml`
2. Replace `your-user`, `your-pw-hash` and `your-ssh-key`s in the `ignition/ignition.yaml`
3. Run `ignition/regenerate-ignition.sh`
4. Create the below `patch-machine.yaml` in `network-peering` folder with the desired `machineClassRef`, `machinePoolRef`, `volumeClassRef`, `volumePoolRef`, `image` etc. as per your environment

```
apiVersion: compute.ironcore.dev/v1alpha1
kind: Machine
metadata:
  name: machine-sample
spec:
  machineClassRef:
    name: new-machineClass
  machinePoolRef:
    name: new-machinePool
  volumes:
  - name: root-disk-1
    ephemeral:
      volumeTemplate:
        spec:
          volumeClassRef:
            name: new-volumeClass
          volumePoolRef:
            name: new-volumePool
          image: gardenlinux:rootfs-dev-20231025
          resources:
            storage: 15Gi
```

5. Update the `kustomization.yaml` with below content
```
patches:
- path: patch-machine.yaml
```

6. Run (`kubectl apply -k ./`) 