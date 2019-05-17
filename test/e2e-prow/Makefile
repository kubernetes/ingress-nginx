TAG ?=v$(shell date +%m%d%Y)-$(shell git rev-parse --short HEAD)
REGISTRY ?= quay.io/kubernetes-ingress-controller
DOCKER ?= docker

IMAGE = $(REGISTRY)/e2e-prow

all: docker-build docker-push

docker-build:
	$(DOCKER) build \
		--pull \
		--build-arg K8S_RELEASE=v1.14.1 \
		--build-arg ETCD_VERSION=v3.3.12 \
		-t $(IMAGE):$(TAG) .

docker-push:
	$(DOCKER) push $(IMAGE):$(TAG)
