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
TAG ?= 0.43
REGISTRY ?= quay.io/kubernetes-ingress-controller
ARCH ?= $(shell go env GOARCH)
DOCKER ?= gcloud docker --

ALL_ARCH = amd64 arm arm64 ppc64le s390x
SED_I?=sed -i
GOHOSTOS ?= $(shell go env GOHOSTOS)

ifeq ($(GOHOSTOS),darwin)
  SED_I=sed -i ''
endif

QEMUVERSION=v2.9.1-1

IMGNAME = nginx
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)

# Set default base image dynamically for each arch
BASEIMAGE?=gcr.io/google-containers/debian-base-$(ARCH):0.3

ifeq ($(ARCH),arm)
	QEMUARCH=arm
endif
ifeq ($(ARCH),arm64)
	QEMUARCH=aarch64
endif
ifeq ($(ARCH),ppc64le)
	QEMUARCH=ppc64le
endif
ifeq ($(ARCH),s390x)
	QEMUARCH=s390x
endif

TEMP_DIR := $(shell mktemp -d)

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
	cp ./* $(TEMP_DIR)
	cd $(TEMP_DIR) && $(SED_I) 's|BASEIMAGE|$(BASEIMAGE)|g' Dockerfile
	cd $(TEMP_DIR) && $(SED_I) "s|ARCH|$(QEMUARCH)|g" Dockerfile

ifeq ($(ARCH),amd64)
	# When building "normally" for amd64, remove the whole line, it has no part in the amd64 image
	cd $(TEMP_DIR) && $(SED_I) "/CROSS_BUILD_/d" Dockerfile
else
	# When cross-building, only the placeholder "CROSS_BUILD_" should be removed
	# Register /usr/bin/qemu-ARCH-static as the handler for ARM binaries in the kernel
	$(DOCKER) run --rm --privileged multiarch/qemu-user-static:register --reset
	curl -sSL https://github.com/multiarch/qemu-user-static/releases/download/$(QEMUVERSION)/x86_64_qemu-$(QEMUARCH)-static.tar.gz | tar -xz -C $(TEMP_DIR)
	cd $(TEMP_DIR) && $(SED_I) "s/CROSS_BUILD_//g" Dockerfile
endif

	$(DOCKER) build -t $(MULTI_ARCH_IMG):$(TAG) $(TEMP_DIR)

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
