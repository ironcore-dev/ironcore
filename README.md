# onmetal-api

![Gardener on Metal Logo](docs/assets/logo.png)

This project contains the types and controllers of the user facing OnMetal API.

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