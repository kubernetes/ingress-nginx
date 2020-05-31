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

PLATFORM ?= amd64

.PHONY: image
image:
	docker buildx build \
		--load \
		--progress plain \
		--platform $(PLATFORM) \
		--build-arg K8S_RELEASE=v1.15.7 \
		--build-arg ETCD_VERSION=v3.3.18 \
		--build-arg RESTY_CLI_VERSION=0.25rc2 \
		--build-arg RESTY_CLI_SHA=a38d850441384fa037a5922ca012dcce8708d0e4abe34ad2fe4164a01b28bdfb \
		--build-arg CHART_TESTING_VERSION=3.0.0-beta.1 \
		--build-arg YAML_LINT_VERSION=1.13.0 \
		--build-arg YAMALE_VERSION=1.8.0 \
		--build-arg LUA_CHECK_VERSION=0.23.0 \
		--build-arg LUA_CHECK_SHA=b4edf3a7702519502696d4ac7372ed1bd6a82ded63bf81f2b1d7e9b37711be2b \
		--build-arg BUSTED_VERSION=v2.0.rc13-0 \
		--build-arg BUSTED_SHA=74ef88ae04545d78b4922e459c1cd459fb5a0356b73a061549d2e4601b5de254 \
		--build-arg HELM_VERSION=v3.2.0 \
		-t $(IMAGE):$(TAG) .
