all: push

# 0.0.0 shouldn't clobber any released builds
TAG = 0.16
REGISTRY = gcr.io/google_containers
ARCH := $(shell go env GOARCH)

IMGNAME = nginx-slim
IMAGE = $(REGISTRY)/${IMGNAME}-$(ARCH)

# Set default base image dynamically for each arch
ifeq ($(ARCH),amd64)
	BASEIMAGE?=gcr.io/google_containers/ubuntu-slim:0.8
endif
ifeq ($(ARCH),ppc64le)
	BASEIMAGE?=ppc64le/ubuntu
endif

TEMP_DIR := $(shell mktemp -d)

container:
	cp ./* $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i.bak 's|BASEIMAGE|$(BASEIMAGE)|g' Dockerfile
	docker build --pull -t $(IMAGE):$(TAG) $(TEMP_DIR)

push: container
	gcloud docker -- push $(IMAGE):$(TAG)

clean:
	docker rmi -f $(PREFIX):$(TAG) || true
