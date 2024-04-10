# Copyright 2024 The Kubernetes Authors. All rights reserved.
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
INIT_BUILDX=$(DIR)/../../hack/init-buildx.sh

# 0.0.0 shouldn't clobber any released builds
SHORT_SHA ?=$(shell git rev-parse --short HEAD)
TAG ?=v$(shell date +%Y%m%d)-$(SHORT_SHA)

REGISTRY ?= gcr.io/k8s-staging-ingress-nginx

IMAGE = $(REGISTRY)/opentelemetry

# required to enable buildx
export DOCKER_CLI_EXPERIMENTAL=enabled

# build with buildx
PLATFORMS?=linux/amd64,linux/arm,linux/arm64
OUTPUT=
PROGRESS=plain

precheck:
ifndef NGINX_VERSION
	$(error NGINX_VERSION variable is required)
endif

build: precheck ensure-buildx
	docker buildx build \
		--label=org.opencontainers.image.source=https://github.com/kubernetes/ingress-nginx \
		--label=org.opencontainers.image.licenses=Apache-2.0 \
		--label=org.opencontainers.image.description="Ingress NGINX Opentelemetry image" \
		--platform=${PLATFORMS} $(OUTPUT) \
		--progress=$(PROGRESS) \
		--build-arg=NGINX_VERSION=$(NGINX_VERSION) \
		--pull \
		--tag $(IMAGE)-$(NGINX_VERSION):$(TAG) rootfs

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

.PHONY: build precheck push ensure-buildx
