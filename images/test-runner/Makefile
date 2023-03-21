# Copyright 2018 The Kubernetes Authors. All rights reserved.
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

# set default shell
SHELL=/bin/bash -o pipefail -o errexit

DIR:=$(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))
INIT_BUILDX=$(DIR)/../../hack/init-buildx.sh

TAG ?=v$(shell date +%m%d%Y)-$(shell git rev-parse --short HEAD)
REGISTRY ?= local

IMAGE = $(REGISTRY)/e2e-test-runner

NGINX_BASE_IMAGE ?= $(shell cat $(DIR)/../../NGINX_BASE)

# required to enable buildx
export DOCKER_CLI_EXPERIMENTAL=enabled

# build with buildx
PLATFORMS?=linux/amd64,linux/arm64
OUTPUT=
PROGRESS=plain

build: ensure-buildx
	docker buildx build \
		--platform=${PLATFORMS} $(OUTPUT) \
		--progress=$(PROGRESS) \
		--pull \
		--build-arg BASE_IMAGE=$(NGINX_BASE_IMAGE) \
		--build-arg GOLANG_VERSION=1.20.1 \
		--build-arg ETCD_VERSION=3.4.3-0 \
		--build-arg K8S_RELEASE=v1.24.2 \
		--build-arg RESTY_CLI_VERSION=0.27 \
		--build-arg RESTY_CLI_SHA=e5f4f3128af49ba5c4d039d0554e5ae91bbe05866f60eccfa96d3653274bff90 \
		--build-arg LUAROCKS_VERSION=3.8.0 \
		--build-arg LUAROCKS_SHA=ab6612ca9ab87c6984871d2712d05525775e8b50172701a0a1cabddf76de2be7 \
		--build-arg CHART_TESTING_VERSION=3.7.0 \
		--build-arg YAML_LINT_VERSION=1.27.1 \
		--build-arg YAMALE_VERSION=4.0.4 \
		--build-arg HELM_VERSION=v3.9.0 \
		-t $(IMAGE):$(TAG) rootfs

# push the cross built image
push: OUTPUT=--push
push: build

# enable buildx
ensure-buildx:
# this is required for cloudbuild
ifeq ("$(wildcard $(INIT_BUILDX))","")
	@curl -sSL https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/hack/init-buildx.sh | bash
else
	@exec $(INIT_BUILDX)
endif
	@echo "done"

.PHONY: build push ensure-buildx
