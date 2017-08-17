# 404-server (default backend)

404-server is a simple webserver that satisfies the ingress, which means it has to do two things:

 1. Serves a 404 page at `/`
 2. Serves 200 on a `/healthz`

## How to release:

The `404-server` Makefile supports multiple architectures, which means it may cross-compile and build an docker image easily.
If you are releasing a new version, please bump the `TAG` value in the `Makefile` before building the images.

How to build and push all images:
```
# Build for linux/amd64 (default)
$ make push
$ make push ARCH=amd64
# ---> gcr.io/google_containers/defaultbackend-amd64:TAG

$ make push-legacy ARCH=amd64
# ---> gcr.io/google_containers/defaultbackend:TAG (image with backwards compatible naming)

$ make push ARCH=arm
# ---> gcr.io/google_containers/defaultbackend-arm:TAG

$ make push ARCH=arm64
# ---> gcr.io/google_containers/defaultbackend-arm64:TAG

$ make push ARCH=ppc64le
# ---> gcr.io/google_containers/defaultbackend-ppc64le:TAG
```

Of course, if you don't want to push the images, just run `make container`
