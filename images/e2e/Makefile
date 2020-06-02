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

TAG ?=v$(shell date +%m%d%Y)-$(shell git rev-parse --short HEAD)
REGISTRY ?= quay.io/kubernetes-ingress-controller

IMAGE = $(REGISTRY)/e2e

HOST_ARCH = $(shell which go >/dev/null 2>&1 && go env GOARCH)
ARCH ?= $(HOST_ARCH)
ifeq ($(ARCH),)
    $(error mandatory variable ARCH is empty, either set it when calling the command or make sure 'go env GOARCH' works)
endif

.PHONY: image
image:
	docker build \
		--build-arg TARGETARCH="$(ARCH)" \
		--build-arg K8S_RELEASE=v1.15.7 \
		--build-arg RESTY_CLI_VERSION=0.25rc2 \
		--build-arg RESTY_CLI_SHA=a38d850441384fa037a5922ca012dcce8708d0e4abe34ad2fe4164a01b28bdfb \
		--build-arg CHART_TESTING_VERSION=3.0.0-beta.1 \
		--build-arg YAML_LINT_VERSION=1.13.0 \
		--build-arg YAMALE_VERSION=1.8.0 \
		--build-arg HELM_VERSION=v3.2.0 \
		-t $(IMAGE):$(TAG) rootfs

.PHONY: show-image
show-image:
	echo -n $(IMAGE):$(TAG)
