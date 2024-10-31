# Kubernetes Local Deploy

Implements a local cluster using kind to validate kubernetes configuration.

#### Requirements

- [Docker](https://www.docker.com/)
- [Kubectl | Kubelet](https://kubernetes.io/docs/tasks/tools/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [Helm](https://helm.sh/docs/intro/install/)

#### How to

```Shell
make start-k8s-cluster
make setup-k8s-namespace
make create-k8s-service-account #generates token for dashboard
make start-k8s-dashboard
make start-k8s-helm-dashboard
```

##### Useful commands

- Kubectl :
Check out this [cheat sheet](https://kubernetes.io/pt-br/docs/reference/kubectl/cheatsheet/) .

```Shell
kubectl cluster-info dump
kubectl 
kubectl describe nodes node-x
kubectl describe pods pod-x
```

- Kind:

```Shell
kind load docker-image profile/image --name cluster
```

- Docker:

``` Shell
docker exec -it dp-icea-brutm-cluster-worker crictl images 
```
