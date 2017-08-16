# Getting Started

This document explains how to get started with developing for Kubernetes Ingress.
It includes how to build, test, and release ingress controllers.

## Dependencies

The build uses dependencies in the `ingress/vendor` directory, which
must be installed before building a binary/image. Occasionally, you
might need to update the dependencies. 

This guide requires you to install the [godep](https://github.com/tools/godep) dependency
tool.

Check the version of `godep` you are using and make sure it is up to date.
```console
$ godep version
godep v74 (linux/amd64/go1.6.1)
```

If you have an older version of `godep`, you can update it as follows:
```console
$ cd $GOPATH/src/ingress
$ go get github.com/tools/godep
$ cd $GOPATH/src/github.com/tools/godep
$ go build -o godep *.go
```

This will automatically save the dependencies to the `vendor/` directory.
```console
$ cd $GOPATH/src/ingress
$ godep save ./...
```

In general, you can follow [this guide](https://github.com/kubernetes/kubernetes/blob/release-1.5/docs/devel/godep.md#using-godep-to-manage-dependencies) to update dependencies.
To update a particular dependency, eg: Kubernetes:
```console
$ cd $GOPATH/src/k8s.io/ingress
$ godep restore
$ go get -u k8s.io/kubernetes
$ cd $GOPATH/src/k8s.io/kubernetes
$ godep restore
$ cd $GOPATH/src/k8s.io/kubernetes/ingress
$ rm -rf Godeps
$ godep save ./...
$ git [add/remove] as needed
$ git commit
```

## Building

All ingress controllers are built through a Makefile. Depending on your
requirements you can build a raw server binary, a local container image,
or push an image to a remote repository.

In order to use your local Docker, you may need to set the following environment variables:
```console
# "gcloud docker" (default) or "docker"
$ export DOCKER=<docker>

# "gcr.io/google_containers" (default), "index.docker.io", or your own registry
$ export REGISTRY=<your-docker-registry>
```
To find the registry simply run: `docker system info | grep Registry`

### Nginx Controller

Build a raw server binary
```console
$ make controllers
```

[TODO](https://github.com/kubernetes/ingress/issues/387): add more specific instructions needed for raw server binary.

Build a local container image
```console
$ make docker-build TAG=<tag> PREFIX=$USER/ingress-controller
```

Push the container image to a remote repository
```console
$ make docker-push TAG=<tag> PREFIX=$USER/ingress-controller
```

### GCE Controller

[TODO](https://github.com/kubernetes/ingress/issues/387): add instructions on building gce controller.

## Deploying

There are several ways to deploy the ingress controller onto a cluster. If you don't have a cluster start by
creating one [here](setup-cluster.md).

* [nginx controller](../../examples/deployment/nginx/README.md)
* [gce controller](../../examples/deployment/gce/README.md)

## Testing

To run unit-tests, enter each directory in `controllers/`
```console
$ cd $GOPATH/src/k8s.io/ingress/controllers/<controller>
$ go test ./...
```

If you have access to a Kubernetes cluster, you can also run e2e tests using ginkgo.
```console
$ cd $GOPATH/src/k8s.io/kubernetes
$ ./hack/ginkgo-e2e.sh --ginkgo.focus=Ingress.* --delete-namespace-on-failure=false
```

See also [related FAQs](../faq#how-are-the-ingress-controllers-tested).

[TODO](https://github.com/kubernetes/ingress/issues/5): add instructions on running integration tests, or e2e against
local-up/minikube.

## Releasing

All Makefiles will produce a release binary, as shown above. To publish this
to a wider Kubernetes user base, push the image to a container registry, like
[gcr.io](https://cloud.google.com/container-registry/). All release images are hosted under `gcr.io/google_containers` and
tagged according to a [semver](http://semver.org/) scheme.

An example release might look like:
```
$ make push TAG=0.8.0 PREFIX=gcr.io/google_containers/glbc
```

Please follow these guidelines to cut a release:

* Update the [release](https://help.github.com/articles/creating-releases/getting_started.md)
page with a short description of the major changes that correspond to a given
image tag.
* Cut a release branch, if appropriate. Release branches follow the format of
`controller-release-version`. Typically, pre-releases are cut from HEAD.
All major feature work is done in HEAD. Specific bug fixes are
cherry-picked into a release branch.
* If you're not confident about the stability of the code, tag it as
alpha or beta. Typically, a release branch should have stable code.


