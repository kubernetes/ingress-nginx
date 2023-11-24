#!/bin/bash
set -e

export REGISTRY="jfrog.joom.it/docker-registry/joom-ingress-nginx"

export BASE_TAG
BASE_TAG=$(cat TAG)
export TAG="${BASE_TAG}-batching-patch"

export ARCH=amd64

make build ARCH=$ARCH
make image PLATFORM=linux/$ARCH TAG=$TAG REGISTRY=$REGISTRY

docker push "${REGISTRY}/controller:${TAG}"
