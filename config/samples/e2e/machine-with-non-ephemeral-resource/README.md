# Machine Sample deployement

This example deploys a `Machine` with `non-ephemeral` `volume` and `networkinterface`. 
The following artifacts will be deployed in your namespace:   
- IronCore `Network`, `NetworkInterface` and `VirtualIP`
- IronCore `Machine` 
- IronCore `Volume`
- Secret containing the `ignition`

## Prerequisites

- [Butane](https://coreos.github.io/butane/)

## Usage
1. Adapt the `namespace` in `kustomization.yaml`
2. Replace `your-user` [^1], `your-pw-hash`  [^2] and `your-ssh-key` [^3] in the `ignition/ignition.yaml`
3. Run `ignition/regenerate-ignition.sh`
4. Create the below `patch-machineclassref.yaml` in `machine-with-non-ephemeral-resource` folder with the desired `machineClassRef` and `machinePoolRef` as per your environment

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

5. Create the below`patch-volume.yaml`in `machine-with-non-ephemeral-resource` folder with the desired `volumeClassRef`and `volumePoolRef` as per your environment

```
apiVersion: storage.ironcore.dev/v1alpha1
kind: Volume
metadata:
  name: volume-sample
spec:
  volumeClassRef:
    name: new-volumeClass The new name of the volume class reference
  image: new-image:rootfs
  volumePoolRef:
    name: new-volumePool
```
6. Update the `kustomization.yaml` with below content
```
patches:
- path: patch-machineclassref.yaml
- path: patch-volume.yaml
```

7. Run (`kubectl apply -k ./`) 


[^1]: e.g. `max`
[^2]: e.g. `$6$pCNgiQprrT/EmeE5$G7wa6wYm1FyuBHeVsuyH9IXGju07csuFwtrynslvSz6O.wFv4Ub8ADPqlBseewQQZQfp.9LCkWyodvJQjH.fe0`
[^3]: e.g. `ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAklOUpkDHrfHY17SbrmTIpNLTGK9Tjom/BWDSU
GPl+nafzlHDTYW7hdI4yZ5ew18JH4JW9jbhUFrviQzM7xlELEVf4h9lFX5QVkbPppSwg0cda3
Pbv7kOdJ/MTyBlWXFCR+HAo3FXRitBqxiX1nKhXpHAZsMciLq8V6RjsNAQwdsdMFvSlVK/7XA
t3FaoJoAsncM1Q9x5+3V0Ww68/eIFmb1zuUFljQJKprrX88XypNDvjYNby6vw/Pb0rwert/En
mZ+AW4OZPnTPI89ZPmVMLuayrD2cE86Z/il8b+gw3r3+1nKatmIkjn2so1d01QraTlMqVSsbx
NrRFi9wrf+M7Q== max@mylaptop.local`