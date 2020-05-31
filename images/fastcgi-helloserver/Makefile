# Copyright 2020 The Kubernetes Authors.
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

# Docker image for e2e testing.

# Use the 0.0 tag for testing, it shouldn't clobber any release builds
TAG ?= 0.0

HOSTARCH := $(shell uname -m | sed -e s/x86_64/amd64/ \
	-e s/s390x/s390x/ \
	-e s/armv7l/arm/ \
	-e s/aarch64.*/arm64/)

ifndef ARCH
ARCH := $(HOSTARCH)
endif
ifeq ($(ARCH),)
    $(error mandatory variable ARCH is empty)
endif

REGISTRY ?= ingress-controller

IMAGE = $(REGISTRY)/fastcgi-helloserver

PKG=k8s.io/ingress-nginx/images/fastcgi-helloserver

.PHONY: image
image: build
	docker buildx build \
		--load \
		--platform $(ARCH) \
		-t $(IMAGE):$(TAG) rootfs

.PHONY: build
build:
	GOARCH=$(ARCH) CGO_ENABLED=0 go build -a -installsuffix cgo \
		-ldflags "-s -w" \
		-o rootfs/fastcgi-helloserver ${PKG}/...
