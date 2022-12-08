 Developing for NGINX Ingress Controller

This document explains how to get started with developing for NGINX Ingress controller.

For the really new contributors, who want to contribute to the INGRESS-NGINX project, but need help with understanding some basic concepts,
that are needed to work with the Kubernetes ingress resource, here is a link to the [New Contributors Guide](https://github.com/kubernetes/ingress-nginx/blob/main/NEW_CONTRIBUTOR.md).
This guide contains tips on how a http/https request travels, from a browser or a curl command,
to the webserver process running inside a container, in a pod, in a Kubernetes cluster, but enters the cluster via a ingress resource.
For those who are familiar with those basic networking concepts like routing of a packet with regards to a
http request, termination of connection, reverseproxy etc. etc., you can skip this and move on to the sections below.
(or read it anyways just for context and also provide feedbacks if any)

## Prerequisites

Install [Go 1.14](https://golang.org/dl/) or later.

!!! note
    The project uses [Go Modules](https://github.com/golang/go/wiki/Modules)

Install [Docker](https://docs.docker.com/engine/install/) (v19.03.0 or later with experimental feature on)

!!! important
    The majority of make tasks run as docker containers

## Quick Start


1. Fork the repository
2. Clone the repository to any location in your work station
3. Add a `GO111MODULE` environment variable with `export GO111MODULE=on`
4. Run `go mod download` to install dependencies

### Local build

Start a local Kubernetes cluster using [kind](https://kind.sigs.k8s.io/), build and deploy the ingress controller

```console
make dev-env
```
- If you are working on the v1.x.x version of this controller, and you want to create a cluster with kubernetes version 1.22, then please visit the [documentation for kind](https://kind.sigs.k8s.io/docs/user/configuration/#a-note-on-cli-parameters-and-configuration-files), and look for how to set a custom image for the kind node (image: kindest/node...), in the kind config file.

### Testing

**Run go unit tests**

```console
make test
```

**Run unit-tests for lua code**

```console
make lua-test
```

Lua tests are located in the directory `rootfs/etc/nginx/lua/test`

!!! important
    Test files must follow the naming convention `<mytest>_test.lua` or it will be ignored


**Run e2e test suite**

```console
make kind-e2e-test
```

To limit the scope of the tests to execute, we can use the environment variable `FOCUS`

```console
FOCUS="no-auth-locations" make kind-e2e-test
```

!!! note
    The variable `FOCUS` defines Ginkgo [Focused Specs](https://onsi.github.io/ginkgo/#focused-specs)

Valid values are defined in the describe definition of the e2e tests like [Default Backend](https://github.com/kubernetes/ingress-nginx/blob/main/test/e2e/defaultbackend/default_backend.go#L29)

The complete list of tests can be found [here](../e2e-tests.md)

### Custom docker image

In some cases, it can be useful to build a docker image and publish such an image to a private or custom registry location.

This can be done setting two environment variables, `REGISTRY` and `TAG`

```console
export TAG="dev"
export REGISTRY="$USER"

make build image
```

and then publish such version with

```console
docker push $REGISTRY/controller:$TAG
```
