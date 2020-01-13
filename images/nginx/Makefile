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
ARCH ?= $(shell go env GOARCH)
DOCKER ?= docker

ALL_ARCH = amd64 arm arm64

IMGNAME = nginx
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)

all: all-container

image-info:
	echo -n '{"image":"$(IMAGE)","tag":"$(TAG)"}'

sub-container-%:
	$(MAKE) ARCH=$* container

sub-push-%:
	$(MAKE) ARCH=$* push

all-container: $(addprefix sub-container-,$(ALL_ARCH))

all-push: $(addprefix sub-push-,$(ALL_ARCH))

container: .container-$(ARCH)
.container-$(ARCH):
	$(DOCKER) buildx build \
		--no-cache \
		--progress plain \
		--platform linux/$(ARCH) \
		-t $(MULTI_ARCH_IMG):$(TAG) rootfs

ifeq ($(ARCH), amd64)
	# This is for to maintain the backward compatibility
	$(DOCKER) tag $(MULTI_ARCH_IMG):$(TAG) $(IMAGE):$(TAG)
endif

push: .push-$(ARCH)
.push-$(ARCH): .container-$(ARCH)
	$(DOCKER) push $(MULTI_ARCH_IMG):$(TAG)
ifeq ($(ARCH), amd64)
	$(DOCKER) push $(IMAGE):$(TAG)
endif

clean: $(addprefix sub-clean-,$(ALL_ARCH))
sub-clean-%:
	$(DOCKER) rmi -f $(IMAGE)-$*:$(TAG) || true

release: all-container all-push
	echo "done"
