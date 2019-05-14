TAG ?=v$(shell date +%m%d%Y)-$(shell git rev-parse --short HEAD)
REGISTRY ?= quay.io/kubernetes-ingress-controller
DOCKER ?= docker

IMAGE = $(REGISTRY)/e2e-prow

all: docker-build docker-push

docker-build:
	$(DOCKER) build -t $(IMAGE):$(TAG) .

docker-push:
	$(DOCKER) push $(IMAGE):$(TAG)
