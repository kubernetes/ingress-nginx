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

# 0.0.0 shouldn't clobber any released builds
TAG ?= 0.96
REGISTRY ?= quay.io/kubernetes-ingress-controller

IMGNAME = nginx
IMAGE = $(REGISTRY)/$(IMGNAME)

PLATFORMS = amd64 arm arm64

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
COMMA := ,

all: container

container:
	DOCKER_CLI_EXPERIMENTAL=enabled docker buildx build \
		--no-cache \
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

ifeq ($(ARCH), amd64)
	# This is for to maintain the backward compatibility
	docker tag $(MULTI_ARCH_IMG):$(TAG) $(IMAGE):$(TAG)
endif

push: container
	$(foreach PLATFORM,$(PLATFORMS), \
		docker push $(IMAGE)-$(PLATFORM):$(TAG);)

ifeq ($(ARCH), amd64)
	docker push $(IMAGE):$(TAG)
endif

release: push
	echo "done"
