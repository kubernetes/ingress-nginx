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
TAG ?= 0.99
REGISTRY ?= quay.io/kubernetes-ingress-controller

IMGNAME = nginx
IMAGE = $(REGISTRY)/$(IMGNAME)

PLATFORMS = amd64 arm arm64

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
COMMA := ,

.PHONY: container
container:
	DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build \
		--progress plain \
		--platform $(subst $(SPACE),$(COMMA),$(PLATFORMS)) \
		--tag $(IMAGE):$(TAG) rootfs

	# https://github.com/docker/buildx/issues/59
	$(foreach PLATFORM,$(PLATFORMS), \
		DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build \
		--load \
		--progress plain \
		--platform $(PLATFORM) \
		--tag $(IMAGE)-$(PLATFORM):$(TAG) rootfs;)

.PHONY: push
push: container
	$(foreach PLATFORM,$(PLATFORMS), \
		docker push $(IMAGE)-$(PLATFORM):$(TAG);)

.PHONY: release
release: push
	echo "done"

.PHONY: init-docker-buildx
init-docker-buildx:
ifneq ($(shell docker buildx 2>&1 >/dev/null; echo $?),)
	$(error "buildx not vailable. Docker 19.03 or higher is required")
endif
	docker run --rm --privileged docker/binfmt:66f9012c56a8316f9244ffd7622d7c21c1f6f28d
	docker buildx create --name ingress-nginx --use || true
	docker buildx inspect --bootstrap
