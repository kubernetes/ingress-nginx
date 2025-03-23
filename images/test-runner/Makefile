# Copyright 2025 The Kubernetes Authors. All rights reserved.
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

BUILDER ?= ingress-nginx
PLATFORMS ?= linux/amd64,linux/arm64
REGISTRY ?= us-central1-docker.pkg.dev/k8s-staging-images/ingress-nginx
IMAGE ?= $(REGISTRY)/e2e-test-runner
TAG ?= $(shell cat TAG)

DIR = $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
BASE_IMAGE ?= $(shell cat $(DIR)/../../NGINX_BASE)
GOLANG_VERSION ?= $(shell cat $(DIR)/../../GOLANG_VERSION)

.PHONY: builder
builder:
	docker buildx create --name $(BUILDER) --bootstrap || :
	docker buildx inspect $(BUILDER)

.PHONY: build
build: builder
	docker buildx build \
		--builder $(BUILDER) \
		--platform $(PLATFORMS) \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) \
		--build-arg GOLANG_VERSION=$(GOLANG_VERSION) \
		--build-arg ETCD_VERSION=3.5.13-0 \
		--build-arg K8S_RELEASE=v1.32.3 \
		--build-arg RESTY_CLI_VERSION=0.27 \
		--build-arg RESTY_CLI_SHA=e5f4f3128af49ba5c4d039d0554e5ae91bbe05866f60eccfa96d3653274bff90 \
		--build-arg LUAROCKS_VERSION=3.8.0 \
		--build-arg LUAROCKS_SHA=ab6612ca9ab87c6984871d2712d05525775e8b50172701a0a1cabddf76de2be7 \
		--build-arg CHART_TESTING_VERSION=3.8.0 \
		--build-arg YAML_LINT_VERSION=1.33.0 \
		--build-arg YAMALE_VERSION=4.0.4 \
		--build-arg HELM_VERSION=3.14.4 \
		--build-arg GINKGO_VERSION=2.23.3 \
		--build-arg GOLINT_VERSION=latest \
		rootfs \
		--tag $(IMAGE):$(TAG) \
		$(OUTPUT)

.PHONY: push
push: OUTPUT = --push
push: build

.PHONY: clean
clean:
	docker buildx rm $(BUILDER) || :
