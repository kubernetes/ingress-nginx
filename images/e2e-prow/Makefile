TAG ?=v$(shell date +%m%d%Y)-$(shell git rev-parse --short HEAD)
REGISTRY ?= quay.io/kubernetes-ingress-controller

IMAGE = $(REGISTRY)/e2e-prow

.PHONY: image
image:
	docker buildx build \
		--pull \
		--push \
		--platform amd64 \
		--build-arg K8S_RELEASE=v1.17.0 \
		--build-arg ETCD_VERSION=v3.3.18 \
		--build-arg KIND_VERSION=v0.8.0 \
		--build-arg GO_VERSION=1.14.2 \
		-t $(IMAGE):$(TAG) .
