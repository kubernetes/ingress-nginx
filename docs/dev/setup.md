# Developer setup

This doc outlines the steps needed to setup a local dev cluster within which you
can deploy/test an ingress controller.

## Deploy a dev cluster

### Single node local cluster

You can run the nginx ingress controller locally on any node with access to the
internet, and the following dependencies: [docker](https://docs.docker.com/engine/getstarted/step_one/), [etcd](https://github.com/coreos/etcd/releases), [golang](https://golang.org/doc/install), [cfssl](https://github.com/cloudflare/cfssl#installation), [openssl](https://www.openssl.org/), [make](https://www.gnu.org/software/make/), [gcc](https://gcc.gnu.org/), [git](https://git-scm.com/download/linux).


Clone the kubernetes repo:
```console
$ cd $GOPATH/src/k8s.io
$ git clone https://github.com/kubernetes/kubernetes.git
```

Add yourself to the docker group, if you haven't done so already (or give
local-up-cluster sudo)
```
$ sudo usermod -aG docker $USER
$ sudo reboot
..
$ docker ps
CONTAINER ID        IMAGE               COMMAND             CREATED             STATUS              PORTS               NAMES
```

**NB: the next step will bring up Kubernetes daemons directly on your dev
machine, no sandbox, iptables rules, routes, loadbalancers, network bridges
etc are created on the host.**

```console
$ cd $GOPATH/src/k8s.io/kubernetes
$ hack/local-up-cluster.sh
```

Check for Ready nodes
```console
$ kubectl get no --context=local
NAME        STATUS    AGE       VERSION
127.0.0.1   Ready     5s        v1.6.0-alpha.0.1914+8ccecf93aa6db5-dirty
```

### Minikube cluster

[Minikube](https://github.com/kubernetes/minikube) is a popular way to bring up
a sandboxed local cluster. You will first need to [install](https://github.com/kubernetes/minikube/releases)
the minikube binary, then bring up a cluster
```console
$ minikube start
```

Check for Ready nodes
```console
$ kubectl get no
NAME       STATUS    AGE       VERSION
minikube   Ready     42m       v1.4.6
```

List the existing addons
```console
$ minikube addons list
- addon-manager: enabled
- dashboard: enabled
- kube-dns: enabled
- heapster: disabled
```

If this list already contains the ingress controller, you don't need to
redeploy it. If the addon controller is disabled, you can enable it with
```console
$ minikube addons enable ingress
```

If the list *does not* contain the ingress controller, you can either update
minikube, or deploy it yourself as shown in the next section.

You may want to consider [using the VM's docker
daemon](https://github.com/kubernetes/minikube/blob/master/README.md#reusing-the-docker-daemon)
when developing.

### CoreOS Kubernetes

[CoreOS Kubernetes](https://github.com/coreos/coreos-kubernetes/) repository has `Vagrantfile`
scripts to easily create a new Kubernetes cluster on VirtualBox, VMware or AWS.

Follow the CoreOS [doc](https://coreos.com/kubernetes/docs/latest/kubernetes-on-vagrant-single.html)
for detailed instructions.

## Deploy the ingress controller

You can deploy an ingress controller on the cluster setup in the previous step
[like this](../../examples/deployment).

## Run against a remote cluster

If the controller you're interested in using supports a "dry-run" flag, you can
run it on any machine that has `kubectl` access to a remote cluster. Eg:
```console
$ cd $GOPATH/k8s.io/ingress/controllers/gce
$ glbc --help
      --running-in-cluster               Optional, if this controller is running in a kubernetes cluster, use the
		 pod secrets for creating a Kubernetes client. (default true)

$ ./glbc --running-in-cluster=false
I1210 17:49:53.202149   27767 main.go:179] Starting GLBC image: glbc:0.9.2, cluster name
```

Note that this is equivalent to running the ingress controller on your local
machine, so if you already have an ingress controller running in the remote
cluster, they will fight for the same ingress.

