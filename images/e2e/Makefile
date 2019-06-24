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
DOCKER ?= docker

IMAGE = $(REGISTRY)/e2e

all: docker-build docker-push

docker-build:
	$(DOCKER) build \
		--pull \
		--build-arg K8S_RELEASE=v1.15.0 \
		--build-arg ETCD_VERSION=v3.3.12 \
		-t $(IMAGE):$(TAG) .

docker-push:
	$(DOCKER) push $(IMAGE):$(TAG)
	$(DOCKER) tag $(IMAGE):$(TAG) $(IMAGE):latest
	$(DOCKER) push $(IMAGE):latest
