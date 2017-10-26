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

TAG ?= 0.14
REGISTRY = quay.io/kubernetes-ingress-controller
ARCH ?= $(shell go env GOARCH)
ALL_ARCH = amd64 arm arm64 ppc64le

QEMUVERSION=v2.9.1

IMGNAME = ubuntu-slim
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)

BUILD_IMAGE ?= ubuntu-build
TAR_FILE ?= rootfs.tar
DOCKER ?= gcloud docker --

# Set default base image dynamically for each arch
ifeq ($(ARCH),amd64)
	BASEIMAGE?=ubuntu:16.04
endif
ifeq ($(ARCH),arm)
	BASEIMAGE?=armhf/ubuntu:16.04
	QEMUARCH=arm
endif
ifeq ($(ARCH),arm64)
	BASEIMAGE?=aarch64/ubuntu:16.04
	QEMUARCH=aarch64
endif
ifeq ($(ARCH),ppc64le)
	BASEIMAGE?=ppc64le/ubuntu:16.04
	QEMUARCH=ppc64le
endif

ifeq ($(ARCH),s390x)
	BASEIMAGE?=s390x/ubuntu:16.04
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
.container-$(ARCH): clean
	cp ./* $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i 's|BASEIMAGE|$(BASEIMAGE)|g' Dockerfile.build
	cd $(TEMP_DIR) && sed -i "s|ARCH|$(QEMUARCH)|g" Dockerfile.build

ifeq ($(ARCH),amd64)
	# When building "normally" for amd64, remove the whole line, it has no part in the amd64 image
	cd $(TEMP_DIR) && sed -i "/CROSS_BUILD_/d" Dockerfile.build
else
	# When cross-building, only the placeholder "CROSS_BUILD_" should be removed
	# Register /usr/bin/qemu-ARCH-static as the handler for ARM binaries in the kernel
	$(DOCKER) run --rm --privileged multiarch/qemu-user-static:register --reset
	curl -sSL https://github.com/multiarch/qemu-user-static/releases/download/$(QEMUVERSION)/x86_64_qemu-$(QEMUARCH)-static.tar.gz | tar -xz -C $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i "s/CROSS_BUILD_//g" Dockerfile.build
endif

	$(DOCKER) build --pull -t $(BUILD_IMAGE) -f $(TEMP_DIR)/Dockerfile.build $(TEMP_DIR)
	$(DOCKER) create --name $(BUILD_IMAGE) $(BUILD_IMAGE)
	$(DOCKER) export $(BUILD_IMAGE) > $(TAR_FILE)
	cp rootfs.tar $(TEMP_DIR)
	$(DOCKER) build --pull -t $(MULTI_ARCH_IMG):$(TAG) $(TEMP_DIR)

ifeq ($(ARCH), amd64)
	# This is for to maintain the backward compatibility
	$(DOCKER) tag $(MULTI_ARCH_IMG):$(TAG) $(IMAGE):$(TAG)
endif

push: .push-$(ARCH)
.push-$(ARCH): .container-$(ARCH)
	$(DOCKER) push $(MULTI_ARCH_IMG):$(TAG)

ifeq ($(ARCH),amd64)
	$(DOCKER) push $(IMAGE):$(TAG)
endif

clean:
	$(DOCKER) rmi -f $(IMAGE):$(TAG) || true
	$(DOCKER) rmi -f $(BUILD_IMAGE)   || true
	$(DOCKER) rm  -f $(BUILD_IMAGE)   || true
	rm -f $(TAR_FILE)

release: all-container all-push
	echo "done"
