# Copyright 2017 The Kubernetes Authors.
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

.PHONY: all
all: all-container

# Use the 0.0 tag for testing, it shouldn't clobber any release builds
TAG ?= 0.26.1
REGISTRY ?= quay.io/kubernetes-ingress-controller
DOCKER ?= docker
SED_I ?= sed -i
GOHOSTOS ?= $(shell go env GOHOSTOS)

# e2e settings
# Allow limiting the scope of the e2e tests. By default run everything
FOCUS ?= .*
# number of parallel test
E2E_NODES ?= 10
# slow test only if takes > 50s
SLOW_E2E_THRESHOLD ?= 50
K8S_VERSION ?= v1.14.1

E2E_CHECK_LEAKS ?=

ifeq ($(GOHOSTOS),darwin)
  SED_I=sed -i ''
endif

REPO_INFO ?= $(shell git config --get remote.origin.url)
GIT_COMMIT ?= git-$(shell git rev-parse --short HEAD)

PKG = k8s.io/ingress-nginx

ARCH ?= $(shell go env GOARCH)
GOARCH = ${ARCH}


GOBUILD_FLAGS := -v

ALL_ARCH = amd64 arm arm64

QEMUVERSION = v4.1.0-1

BUSTED_ARGS =-v --pattern=_test

GOOS = linux

MULTI_ARCH_IMAGE = $(REGISTRY)/nginx-ingress-controller-${ARCH}

# use vendor directory instead of go modules https://github.com/golang/go/wiki/Modules
GO111MODULE=off

export ARCH
export TAG
export PKG
export GOARCH
export GOOS
export GO111MODULE
export GIT_COMMIT
export GOBUILD_FLAGS
export REPO_INFO
export BUSTED_ARGS
export IMAGE
export E2E_NODES
export E2E_CHECK_LEAKS
export SLOW_E2E_THRESHOLD

# Set default base image dynamically for each arch
BASEIMAGE?=quay.io/kubernetes-ingress-controller/nginx-$(ARCH):daf8634acf839708722cffc67a62e9316a2771c6

ifeq ($(ARCH),arm)
	QEMUARCH=arm
endif
ifeq ($(ARCH),arm64)
	QEMUARCH=aarch64
endif

TEMP_DIR := $(shell mktemp -d)

DOCKERFILE := $(TEMP_DIR)/rootfs/Dockerfile

.PHONY: sub-container-%
sub-container-%:
	$(MAKE) ARCH=$* build container

.PHONY: sub-push-%
sub-push-%:
	$(MAKE) ARCH=$* push

.PHONY: all-container
all-container: $(addprefix sub-container-,$(ALL_ARCH))

.PHONY: all-push
all-push: $(addprefix sub-push-,$(ALL_ARCH))

.PHONY: container
container: clean-container .container-$(ARCH)

.PHONY: .container-$(ARCH)
.container-$(ARCH):
	mkdir -p $(TEMP_DIR)/rootfs
	cp bin/$(ARCH)/nginx-ingress-controller $(TEMP_DIR)/rootfs/nginx-ingress-controller
	cp bin/$(ARCH)/dbg $(TEMP_DIR)/rootfs/dbg
	cp bin/$(ARCH)/wait-shutdown $(TEMP_DIR)/rootfs/wait-shutdown

	cp -RP ./* $(TEMP_DIR)
	$(SED_I) "s|BASEIMAGE|$(BASEIMAGE)|g" $(DOCKERFILE)
	$(SED_I) "s|QEMUARCH|$(QEMUARCH)|g" $(DOCKERFILE)

ifeq ($(ARCH),amd64)
	# When building "normally" for amd64, remove the whole line, it has no part in the amd64 image
	$(SED_I) "/CROSS_BUILD_/d" $(DOCKERFILE)
else
	# When cross-building, only the placeholder "CROSS_BUILD_" should be removed
	curl -sSL https://github.com/multiarch/qemu-user-static/releases/download/$(QEMUVERSION)/x86_64_qemu-$(QEMUARCH)-static.tar.gz | tar -xz -C $(TEMP_DIR)/rootfs
	$(SED_I) "s/CROSS_BUILD_//g" $(DOCKERFILE)
endif

	echo "Building docker image..."
	$(DOCKER) build --no-cache --pull -t $(MULTI_ARCH_IMAGE):$(TAG) $(TEMP_DIR)/rootfs


.PHONY: clean-container
clean-container:
	@$(DOCKER) rmi -f $(MULTI_ARCH_IMAGE):$(TAG) || true

.PHONY: register-qemu
register-qemu:
	# Register /usr/bin/qemu-ARCH-static as the handler for binaries in multiple platforms
	@$(DOCKER) run --rm --privileged multiarch/qemu-user-static:register --reset >&2

.PHONY: push
push: .push-$(ARCH)

.PHONY: .push-$(ARCH)
.push-$(ARCH):
	$(DOCKER) push $(MULTI_ARCH_IMAGE):$(TAG)

.PHONY: build
build:
	@build/build.sh

.PHONY: build-plugin
build-plugin:
	@build/build-plugin.sh

.PHONY: clean
clean:
	rm -rf bin/ .gocache/

.PHONY: static-check
static-check:
	@build/static-check.sh

.PHONY: test
test:
	@build/test.sh

.PHONY: lua-test
lua-test:
	@build/test-lua.sh

.PHONY: e2e-test
e2e-test:
	@build/run-e2e-suite.sh

.PHONY: e2e-test-image
e2e-test-image: e2e-test-binary
	make -C test/e2e-image

.PHONY: e2e-test-binary
e2e-test-binary:
	@ginkgo build ./test/e2e

.PHONY: cover
cover:
	@build/cover.sh
	echo "Uploading coverage results..."
	@curl -s https://codecov.io/bash | bash

.PHONY: vet
vet:
	@go vet $(shell go list ${PKG}/internal/... | grep -v vendor)

.PHONY: release
release: all-container all-push
	echo "done"

.PHONY: check_dead_links
check_dead_links:
	@docker run -t \
	  -v $$PWD:/tmp aledbf/awesome_bot:0.1 \
	  --allow-dupe \
	  --allow-redirect $(shell find $$PWD -mindepth 1 -name "*.md" -printf '%P\n' | grep -v vendor | grep -v Changelog.md)

.PHONY: dep-ensure
dep-ensure:
	go mod tidy -v
	find vendor -name '*_test.go' -delete
	go mod vendor

.PHONY: dev-env
dev-env:
	@build/dev-env.sh

.PHONY: live-docs
live-docs:
	@docker build --pull -t ingress-nginx/mkdocs images/mkdocs
	@docker run --rm -it -p 3000:3000 -v ${PWD}:/docs ingress-nginx/mkdocs

.PHONY: build-docs
build-docs:
	@docker build --pull -t ingress-nginx/mkdocs images/mkdocs
	@docker run --rm -v ${PWD}:/docs ingress-nginx/mkdocs build

.PHONY: misspell
misspell:
	@go get github.com/client9/misspell/cmd/misspell
	misspell \
		-locale US \
		-error \
		cmd/* internal/* deploy/* docs/* design/* test/* README.md

.PHONY: kind-e2e-test
kind-e2e-test:
	test/e2e/run.sh

.PHONY: run-ingress-controller
run-ingress-controller:
	@build/run-ingress-controller.sh
