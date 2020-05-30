# Copyright 2017 The Kubernetes Authors. All rights reserved.
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

.DEFAULT_GOAL:=container

# set default shell
SHELL=/bin/bash -o pipefail

# 0.0.0 shouldn't clobber any released builds
TAG ?= 0.103
REGISTRY ?= quay.io/kubernetes-ingress-controller

IMAGE = $(REGISTRY)/nginx

.PHONY: image
image:
	docker buildx build \
		--pull \
		--push \
		--progress plain \
		--platform amd64,arm,arm64,s390x \
		--tag $(IMAGE):$(TAG) rootfs

.PHONY: init-docker-buildx
init-docker-buildx:
ifneq ($(shell docker buildx 2>&1 >/dev/null; echo $?),)
	$(error "buildx not vailable. Docker 19.03 or higher is required")
endif
	docker run --rm --privileged docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64
	docker buildx create --name ingress-nginx --use || true
	docker buildx inspect --bootstrap
