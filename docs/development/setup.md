# Local Development Setup

## Requirements 

* `go` >= 1.12
* `git`, `make` and `kubectl`
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

The controller manager can be started via the following command

```shell
make run
```

## Apply Sample Manifests

The `config/samples` folder contains samples for all CRDs supported by this project. You can apply any of the samples by
running

```shell
kubectl apply -f config/samples/SOME_RESOURCE.yaml
```

## Rebuilding API Type and Manifests

Everytime a change has been done to any of the types definitions of a CRD, the corresponding manifests and code artefacts
have to be rebuild.

```shell
make generate
make manifests
```

!!! note
    Make sure you install all new versions of the CRDs into your cluster by running `make install` after new manifests 
    have been generated.

## Setup formatting tools

Project uses `gofmt` and `goimports` for formatting. `gofmt` is used with default settings. While `goimports` should be used with `--local github.com/onmetal` flag, so that `goimports` would sort `onmetal` pkgs separately.

You can automate running formatting tools in your IDE.

- **VSCode** -- add following to the `settings.json`:
```
    "go.formatTool": "goimports",
    "gopls": {
        "formatting.local": "github.com/onmetal",
    },
```
- **Goland** -- go to `File -> Settings -> Tools -> File Watchers` and replace contents of `Arguments` with `--local github.com/onmetal -w $FilePath$`

## Cleanup

Remove the CRDs from your cluster.

```shell
make uninstall
```

--8<-- "hack/docs/abbreviations.md"