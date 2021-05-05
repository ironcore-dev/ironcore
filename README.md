# onmetal-api

This is a bleeding edge API POC.

## Installation

To install the CRDs into your cluster 

```shell
make install
```

## Install sample Custom Resources

```shell
kubectl apply -f config/samples/
```

## Regenerate Manifests

```shell
make manifests
```

## Cleanup

Remove the CRDs from your cluster

```shell
make uninstall
```