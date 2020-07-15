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

# required to enable buildx
export DOCKER_CLI_EXPERIMENTAL=enabled

# build with buildx
PLATFORMS?=linux/amd64
OUTPUT=
PROGRESS=plain

build: ensure-buildx
	docker buildx build \
		--platform=${PLATFORMS} $(OUTPUT) \
		--progress=$(PROGRESS) \
		--pull \
		--build-arg BASE_IMAGE=us.gcr.io/k8s-artifacts-prod/ingress-nginx/nginx:v20200708-g00f4a215d@sha256:b0727cebafc35f5ab50f3fa0be5fda7f79937f1683f525f40982e75b882195d8 \
		--build-arg GOLANG_VERSION=1.14.5 \
		--build-arg ETCD_VERSION=3.4.3-0 \
		--build-arg K8S_RELEASE=v1.15.7 \
		--build-arg RESTY_CLI_VERSION=0.25 \
		--build-arg RESTY_CLI_SHA=7c3d1317d7ddab2030b49610c89e301995d4ae0f815f16466c48144d05ec7bcb \
		--build-arg LUAROCKS_VERSION=3.3.1 \
		--build-arg LUAROCKS_SHA=837481e408f7c06b59befe7ec194537c657687d624894bca7f79034302141a34 \
		--build-arg CHART_TESTING_VERSION=3.0.0-beta.1 \
		--build-arg YAML_LINT_VERSION=1.13.0 \
		--build-arg YAMALE_VERSION=1.8.0 \
		--build-arg HELM_VERSION=v3.2.1 \
		-t $(IMAGE):$(TAG) rootfs

# push the cross built image
push: OUTPUT=--push
push: build

# enable buildx
ensure-buildx:
# this is required for cloudbuild
ifeq ("$(wildcard $(INIT_BUILDX))","")
	@curl -sSL https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/hack/init-buildx.sh | bash
else
	@exec $(INIT_BUILDX)
endif
	@echo "done"

.PHONY: build push ensure-buildx
