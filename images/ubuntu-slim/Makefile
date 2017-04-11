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

all: push

TAG ?= 0.9
REGISTRY = gcr.io/google_containers
ARCH ?= $(shell go env GOARCH)

IMGNAME = ubuntu-slim
IMAGE = $(REGISTRY)/${IMGNAME}-$(ARCH)

BUILD_IMAGE ?= ubuntu-build
TAR_FILE ?= rootfs.tar
PUSH_TOOL ?= gcloud

# Set default base image dynamically for each arch
ifeq ($(ARCH),amd64)
	BASEIMAGE?=ubuntu:16.04
endif
ifeq ($(ARCH),ppc64le)
	BASEIMAGE?=ppc64le/ubuntu:16.04
endif

TEMP_DIR := $(shell mktemp -d)

container: clean
	cp ./* $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i.bak 's|BASEIMAGE|$(BASEIMAGE)|g' Dockerfile.build
	docker build --pull -t $(BUILD_IMAGE) -f $(TEMP_DIR)/Dockerfile.build $(TEMP_DIR)
	docker create --name $(BUILD_IMAGE) $(BUILD_IMAGE)
	docker export $(BUILD_IMAGE) > $(TAR_FILE)
	cp rootfs.tar $(TEMP_DIR)
	docker build --pull -t $(IMAGE):$(TAG) $(TEMP_DIR)

push: container
	$(PUSH_TOOL) docker -- push $(PREFIX):$(TAG)

clean:
	docker rmi -f $(IMAGE):$(TAG) || true
	docker rmi -f $(BUILD_IMAGE)   || true
	docker rm  -f $(BUILD_IMAGE)   || true
	rm -f $(TAR_FILE)
