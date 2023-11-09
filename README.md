# rancher-bind

## What is it?

This repository hosts the backend implementation for the [kube-bind][] project for rancher.

This allows external clusters to consume rancher native services without a need for installation.

- The exposed rancher CRDs gets installed in the consumer cluster, objects are syncronized with the provider side.
- Permissions for the consumer cluster access are added gradually based on the APIServiceBinding spec.
- The service provider does not inject controllers/operators into the service consumer's cluster.

Functionality, such as fine-grained kubeconfig for rancher cluster is exposed via plugin commands.

[kube-bind]: https://github.com/kube-bind/kube-bind

## Prerequisites

- [Kubectl][]
- [Krew][]
- [Rancher][]

[Kubectl]: https://kubernetes.io/docs/tasks/tools/#kubectl
[Rancher]: https://ranchermanager.docs.rancher.com/pages-for-subheaders/install-upgrade-on-a-kubernetes-cluster
[Krew]: https://krew.sigs.k8s.io/docs/user-guide/setup/install/

## Try it out

### Fine-grained kubeconfig for Rancher cluster

```shell
kubectl krew index add bind https://github.com/Danil-Grigorev/rancher-bind.git
kubectl krew install rancher-bind/rancher-bind
kubectl rancher-bind -f ./example-role.yaml > kubeconfig
cat kubeconfig
# Outputs:
# apiVersion: v1
# kind: Config
# clusters:
# - name: "local"
# ...
```
