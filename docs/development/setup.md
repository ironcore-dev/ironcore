# Local Development Setup

## Requirements

* `go` >= 1.17
* `git`, `make` and `kubectl`
* [Kustomize](https://kustomize.io/)
* Access to a Kubernetes cluster ([Minikube](https://minikube.sigs.k8s.io/docs/), [kind](https://kind.sigs.k8s.io/) or a
  real cluster)

## Clone the Repository

To bring up and start locally the `onmetal-api` project for development purposes you first need to clone the repository.

```shell
git clone git@github.com:onmetal/onmetal-api.git
cd onmetal-api
```

## Install APIs into the Cluster

Your Kubernetes API server needs to know about the APIs which come with the `onmetal-api` project. To install the APIs
your cluster, run

```shell
make install
```

**Note**: This requires the `APISERVER_IMG` (Makefile default set to `apiserver`) to be pullable from your kubernetes
cluster. For local development with `kind`, a make target that builds and loads the api server image and then applies
the manifests is available via

```shell
make kind-install
```

## Start the Controller Manager

The controller manager can be started via the following command

```shell
make run
```

## Apply Sample Manifests

The `config/samples` folder contains samples for all APIs supported by this project. You can apply any of the samples by
running

```shell
kubectl apply -f config/samples/SOME_RESOURCE.yaml
```

## Rebuilding API Type and Manifests

Everytime a change has been done to any of the types definitions, the corresponding manifests and generated code pieces
have to be rebuilt.

```shell
make generate
make manifests
```

**Note**: Make sure your APIs are up-to-date by running `make install` / `make kind-install` after your code / manifests
have been regenerated.

## Setup formatting tools

The project uses `gofmt` and `goimports` for formatting. `gofmt` is used with default settings. While `goimports` should
be used with `--local github.com/onmetal` flag, so that `goimports` would sort `onmetal` pkgs separately.

You can automate running formatting tools in your IDE.

- **VSCode** -- add following to the `settings.json`:

```
    "go.formatTool": "goimports",
    "gopls": {
        "formatting.local": "github.com/onmetal",
    },
```

- **Goland** -- go to `File -> Settings -> Tools -> File Watchers` and replace contents of `Arguments`
  with `--local github.com/onmetal -w $FilePath$`

## Cleanup

To remove the APIs from your cluster, simply run

```shell
make uninstall
```

--8<-- "hack/docs/abbreviations.md"