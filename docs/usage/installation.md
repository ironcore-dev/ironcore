# Installing IronCore

## Requirements

* `go` >= 1.20
* `git`, `make` and `kubectl`
* [Kustomize](https://kustomize.io/)
* Access to a Kubernetes cluster ([Minikube](https://minikube.sigs.k8s.io/docs/), [kind](https://kind.sigs.k8s.io/) or a real cluster)

## Clone the Repository

To bring up and install the `ironcore` project, you first need to clone the repository.

```shell
git clone git@github.com:ironcore-dev/ironcore.git
cd ironcore
```

## Install cert-manager

If there is no [cert-manager](https://cert-manager.io/docs/) present in the cluster it needs to be installed.

```shell
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.16.2/cert-manager.yaml
```

## Install APIs into the Cluster

Your Kubernetes API server needs to know about the APIs which come with the `ironcore` project. To install the APIs into your cluster, run

```shell
make install
```

**Note**: This requires the `APISERVER_IMG` (Makefile default set to `apiserver`) to be pullable from your kubernetes
cluster. For local development with `kind`, a make target that builds and loads the api server image and then applies
the manifests is available via

```shell
make kind-install
```

**Note**: In case that there are multiple environments running, ensure that `kind get clusters` is pointing to the
default kind cluster.

## Deploy the Controller Manager

The controller manager can be started via the following command

```shell
make kind-deploy
```
## validate

Make sure you have all the below pods running

```shell
$ kubectl get po -n ironcore-system
NAME                                           READY   STATUS    RESTARTS       AGE
ironcore-apiserver-85995846f9-47247            1/1     Running   0              136m
ironcore-controller-manager-84bf4cc6d5-l224c   2/2     Running   0              136m
ironcore-etcd-0                                1/1     Running   0              143m
```

## Apply Sample Manifests

The `config/samples` folder contains samples for all APIs supported by this project. You can apply any of the samples by
running

```shell
kubectl apply -f config/samples/SOME_RESOURCE.yaml
```

## Cleanup

To remove the APIs from your cluster, simply run

```shell
make uninstall
```
