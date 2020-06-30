# Developing for NGINX Ingress Controller

This document explains how to get started with developing for NGINX Ingress controller.
It includes how to build, test, and release ingress controllers.

## Quick Start

### Getting the code

The code must be checked out as a subdirectory of k8s.io, and not github.com.

```
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
# Replace "$YOUR_GITHUB_USERNAME" below with your github username
git clone https://github.com/$YOUR_GITHUB_USERNAME/ingress-nginx.git
cd ingress-nginx
```

### Initial developer environment build

```
$ make dev-env
```

### Updating the deployment

The nginx controller container image can be rebuilt using:
```
$ ARCH=amd64 TAG=dev REGISTRY=$USER/ingress-controller make build image
```

The image will only be used by pods created after the rebuild. To delete old pods which will cause new ones to spin up:
```
$ kubectl get pods -n ingress-nginx
$ kubectl delete pod -n ingress-nginx ingress-nginx-controller-<unique-pod-id>
```

## Dependencies

The build uses dependencies in the `vendor` directory, which
must be installed before building a binary/image. Occasionally, you
might need to update the dependencies.

This guide requires you to install go 1.13 or newer.

This will automatically save the dependencies to the `vendor/` directory.

```console
$ go get
$ make dep-ensure
```

## Building

All ingress controllers are built through a Makefile. Depending on your
requirements you can build a raw server binary, a local container image,
or push an image to a remote repository.

In order to use your local Docker, you may need to set the following environment variables:

```console
# "gcloud docker" (default) or "docker"
$ export DOCKER=<docker>

# "quay.io/kubernetes-ingress-controller" (default), "index.docker.io", or your own registry
$ export REGISTRY=<your-docker-registry>
```

To find the registry simply run: `docker system info | grep Registry`

### Building the e2e test image

The e2e test image can also be built through the Makefile.

```console
$ make -C test/e2e-image build
```

Then you can load the docker image using kind:

```console
$ kind load docker-image --name="ingress-nginx-dev" nginx-ingress-controller:e2e
```

### Nginx Controller

Build a raw server binary
```console
$ make build
```

[TODO](https://github.com/kubernetes/ingress-nginx/issues/387): add more specific instructions needed for raw server binary.

Build a local container image

```console
$ TAG=<tag> REGISTRY=$USER/ingress-controller make image
```

## Deploying

There are several ways to deploy the ingress controller onto a cluster.
Please check the [deployment guide](./deploy/index.md)

## Testing

To run unit-tests, just run

```console
$ cd $GOPATH/src/k8s.io/ingress-nginx
$ make test
```

If you have access to a Kubernetes cluster, you can also run e2e tests using ginkgo.

```console
$ cd $GOPATH/src/k8s.io/ingress-nginx
$ KIND_CLUSTER_NAME="ingress-nginx-test" make kind-e2e-test
```
To set focus to a particular set of tests, a FOCUS flag can be set.

```console
KIND_CLUSTER_NAME="ingress-nginx-test" FOCUS="no-auth-locations" make kind-e2e-test
```

NOTE: if your e2e pod keeps hanging in an ImagePullBackoff, make sure you've made your e2e nginx-ingress-controller image available to minikube as explained in the **Building the e2e test image** section

To run unit-tests for lua code locally, run:

```console
$ cd $GOPATH/src/k8s.io/ingress-nginx
$ ./rootfs/etc/nginx/lua/test/up.sh
$ make lua-test
```

Lua tests are located in `$GOPATH/src/k8s.io/ingress-nginx/rootfs/etc/nginx/lua/test`. When creating a new test file it must follow the naming convention `<mytest>_test.lua` or it will be ignored.

## Releasing

All Makefiles will produce a release binary, as shown above. To publish this
to a wider Kubernetes user base, push the image to a container registry, like
[gcr.io](https://cloud.google.com/container-registry/). All release images are hosted under `gcr.io/google_containers` and
tagged according to a [semver](http://semver.org/) scheme.

An example release might look like:
```
$ make release
```

Please follow these guidelines to cut a release:

* Update the [release](https://help.github.com/articles/creating-releases/)
page with a short description of the major changes that correspond to a given
image tag.
* Cut a release branch, if appropriate. Release branches follow the format of
`controller-release-version`. Typically, pre-releases are cut from HEAD.
All major feature work is done in HEAD. Specific bug fixes are
cherry-picked into a release branch.
* If you're not confident about the stability of the code,
[tag](https://help.github.com/articles/working-with-tags/) it as alpha or beta.
Typically, a release branch should have stable code.
