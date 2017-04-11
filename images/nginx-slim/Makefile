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

# 0.0.0 shouldn't clobber any released builds
TAG = 0.16
REGISTRY = gcr.io/google_containers
ARCH ?= $(shell go env GOARCH)

IMGNAME = nginx-slim
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)

# Set default base image dynamically for each arch
BASEIMAGE?=gcr.io/google_containers/ubuntu-slim-$(ARCH):0.9

TEMP_DIR := $(shell mktemp -d)

container:
	cp ./* $(TEMP_DIR)
	cd $(TEMP_DIR) && sed -i.bak 's|BASEIMAGE|$(BASEIMAGE)|g' Dockerfile
	docker build --pull -t $(MULTI_ARCH_IMG):$(TAG) $(TEMP_DIR)
ifeq ($(ARCH), amd64)
	docker tag $(MULTI_ARCH_IMG):$(TAG) $(IMAGE):$(TAG)
endif

push: container
	gcloud docker -- push $(MULTI_ARCH_IMG):$(TAG)
ifeq ($(ARCH), amd64)
	gcloud docker -- push $(IMAGE):$(TAG)
endif

clean:
	docker rmi -f $(PREFIX):$(TAG) || true
