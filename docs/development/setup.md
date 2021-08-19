# Local Development Setup

## Prerequisites 

* `go` >= 1.12
* `git` and `make`
* [Kustomize](https://kustomize.io/)
* Access to a Kubernetes cluster ([Minikube](https://minikube.sigs.k8s.io/docs/), [kind](https://kind.sigs.k8s.io/) or a real cluster)

## Clone the Repository

To bring up and start locally the `onmetal-api` project for development purposes you first need to clone the repository.

```shell
git clone git@github.com:onmetal/onmetal-api.git
cd onmetal-api
```

## Install CRDs into Cluster

Your Kubernetes API server needs to know about the CRDs which come with the `onmetal-api` project. 
To install the CRDs into your cluster run

```shell
make install
```

## Start the Controller Manager

```shell
make run
```

## Apply Sample Manifests

## Rebuilding API Type and Manifests

```shell
make generate
make manifests
```

## Cleanup

Remove the CRDs from your cluster.

```shell
make uninstall
```

--8<-- "hack/docs/abbreviations.md"