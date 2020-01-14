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

all: docker-build docker-push

docker-build:
	docker build \
		--pull \
		--load \
		--build-arg K8S_RELEASE=v1.15.7 \
		--build-arg ETCD_VERSION=v3.3.18 \
		--build-arg GOLANG_VERSION=1.13.5 \
		--build-arg GOLANG_SHA=27d356e2a0b30d9983b60a788cf225da5f914066b37a6b4f69d457ba55a626ff \
		--build-arg RESTY_CLI_VERSION=0.25rc2 \
		--build-arg RESTY_CLI_SHA=a38d850441384fa037a5922ca012dcce8708d0e4abe34ad2fe4164a01b28bdfb \
		-t $(IMAGE):$(TAG) .

docker-push:
	docker push $(IMAGE):$(TAG)
	docker tag $(IMAGE):$(TAG) $(IMAGE):latest
	docker push $(IMAGE):latest
