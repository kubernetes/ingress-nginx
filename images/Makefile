# Copyright 2024 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.DEFAULT_GOAL:=build

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

DIR:=$(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))
INIT_BUILDX=$(DIR)/../hack/init-buildx.sh


BASE_IMAGE = $(shell cat $(DIR)/../NGINX_BASE)

# The env below is called GO_VERSION and not GOLANG_VERSION because 
# the gcb image we use to build already defines GOLANG_VERSION and is a 
# really old version
GO_VERSION = $(shell cat $(DIR)/../GOLANG_VERSION)

REGISTRY ?= local
NAME ?=

IMAGE = $(REGISTRY)/$(NAME)
TAG ?= $(shell cat $(NAME)/TAG)

EXTRAARGS ?= $(shell cat $(NAME)/EXTRAARGS)


# required to enable buildx
export DOCKER_CLI_EXPERIMENTAL=enabled

# build with buildx
PLATFORMS?=linux/amd64,linux/arm,linux/arm64
OUTPUT=
PROGRESS=plain


precheck:
ifndef NAME
	$(error NAME variable is required)
endif

build: precheck ensure-buildx
	docker buildx build \
		--label=org.opencontainers.image.source=https://github.com/kubernetes/ingress-nginx \
		--label=org.opencontainers.image.licenses=Apache-2.0 \
		--label=org.opencontainers.image.description="Ingress NGINX $(NAME) image" \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--build-arg GOLANG_VERSION=$(GO_VERSION) \
		--platform=${PLATFORMS} $(OUTPUT) \
		--progress=$(PROGRESS) \
		--pull $(EXTRAARGS) \
		-t $(IMAGE):$(TAG) $(NAME)/rootfs

# push the cross built image
push: OUTPUT=--push
push: build

test: precheck
	cd $(NAME)/rootfs && go test ./...

test-e2e: precheck
	cd $(NAME) && ./hack/e2e.sh

# enable buildx
ensure-buildx:
# this is required for cloudbuild
ifeq ("$(wildcard $(INIT_BUILDX))","")
	@curl -sSL https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/hack/init-buildx.sh | bash
else
	@exec $(INIT_BUILDX)
endif
	@echo "done"

.PHONY: build push ensure-buildx test test-e2e precheck
