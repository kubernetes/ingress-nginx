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
TAG ?= 0.22.0
REGISTRY ?= quay.io/kubernetes-ingress-controller
DOCKER ?= docker
SED_I ?= sed -i
GOHOSTOS ?= $(shell go env GOHOSTOS)

# e2e settings
# Allow limiting the scope of the e2e tests. By default run everything
FOCUS ?= .*
# number of parallel test
E2E_NODES ?= 4
# slow test only if takes > 40s
SLOW_E2E_THRESHOLD ?= 40

NODE_IP ?= $(shell minikube ip)

ifeq ($(GOHOSTOS),darwin)
  SED_I=sed -i ''
endif

REPO_INFO=$(shell git config --get remote.origin.url)

ifndef GIT_COMMIT
  GIT_COMMIT := git-$(shell git rev-parse --short HEAD)
endif

PKG = k8s.io/ingress-nginx

ARCH ?= $(shell go env GOARCH)
GOARCH = ${ARCH}
DUMB_ARCH = ${ARCH}

GOBUILD_FLAGS :=

ALL_ARCH = amd64 arm64

QEMUVERSION = v3.0.0

BUSTED_ARGS =-v --pattern=_test

IMGNAME = nginx-ingress-controller
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)

# Set default base image dynamically for each arch
BASEIMAGE?=quay.io/kubernetes-ingress-controller/nginx-$(ARCH):0.75

ifeq ($(ARCH),arm64)
	QEMUARCH=aarch64
endif

TEMP_DIR := $(shell mktemp -d)

DEF_VARS:=ARCH=$(ARCH)           \
	TAG=$(TAG)               \
	PKG=$(PKG)               \
	GOARCH=$(GOARCH)         \
	GIT_COMMIT=$(GIT_COMMIT) \
	REPO_INFO=$(REPO_INFO)   \
	PWD=$(PWD)

DOCKERFILE := $(TEMP_DIR)/rootfs/Dockerfile

.PHONY: image-info
image-info:
	echo -n '{"image":"$(IMAGE)","tag":"$(TAG)"}'

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
	@echo "+ Copying artifact to temporary directory"
	mkdir -p $(TEMP_DIR)/rootfs
	cp bin/$(ARCH)/nginx-ingress-controller $(TEMP_DIR)/rootfs/nginx-ingress-controller
	cp bin/$(ARCH)/dbg $(TEMP_DIR)/rootfs/dbg
	@echo "+ Building container image $(MULTI_ARCH_IMG):$(TAG)"
	cp -RP ./* $(TEMP_DIR)
	$(SED_I) "s|BASEIMAGE|$(BASEIMAGE)|g" $(DOCKERFILE)
	$(SED_I) "s|QEMUARCH|$(QEMUARCH)|g" $(DOCKERFILE)
	$(SED_I) "s|DUMB_ARCH|$(DUMB_ARCH)|g" $(DOCKERFILE)

ifeq ($(ARCH),amd64)
	# When building "normally" for amd64, remove the whole line, it has no part in the amd64 image
	$(SED_I) "/CROSS_BUILD_/d" $(DOCKERFILE)
else
	# When cross-building, only the placeholder "CROSS_BUILD_" should be removed
	curl -sSL https://github.com/multiarch/qemu-user-static/releases/download/$(QEMUVERSION)/x86_64_qemu-$(QEMUARCH)-static.tar.gz | tar -xz -C $(TEMP_DIR)/rootfs
	$(SED_I) "s/CROSS_BUILD_//g" $(DOCKERFILE)
endif

	$(DOCKER) build --no-cache --pull -t $(MULTI_ARCH_IMG):$(TAG) $(TEMP_DIR)/rootfs

ifeq ($(ARCH), amd64)
	# This is for maintaining backward compatibility
	$(DOCKER) tag $(MULTI_ARCH_IMG):$(TAG) $(IMAGE):$(TAG)
endif

.PHONY: clean-container
clean-container:
	@echo "+ Deleting container image $(MULTI_ARCH_IMG):$(TAG)"
	$(DOCKER) rmi -f $(MULTI_ARCH_IMG):$(TAG) || true

.PHONY: register-qemu
register-qemu:
	# Register /usr/bin/qemu-ARCH-static as the handler for binaries in multiple platforms
	$(DOCKER) run --rm --privileged multiarch/qemu-user-static:register --reset

.PHONY: push
push: .push-$(ARCH)

.PHONY: .push-$(ARCH)
.push-$(ARCH):
	$(DOCKER) push $(MULTI_ARCH_IMG):$(TAG)
ifeq ($(ARCH), amd64)
	$(DOCKER) push $(IMAGE):$(TAG)
endif

.PHONY: build
build:
	@echo "+ Building bin/$(ARCH)/nginx-ingress-controller"
	@$(DEF_VARS) \
	GOBUILD_FLAGS="$(GOBUILD_FLAGS)" \
	build/go-in-docker.sh build/build.sh

.PHONY: clean
clean:
	rm -rf bin/ .gocache/ .env

.PHONY: static-check
static-check:
	@$(DEF_VARS) \
	build/go-in-docker.sh build/static-check.sh

.PHONY: test
test:
	@$(DEF_VARS)                 \
	NODE_IP=$(NODE_IP)           \
	DOCKER_OPTS="-i --net=host"  \
	build/go-in-docker.sh build/test.sh

.PHONY: lua-test
lua-test:
	@$(DEF_VARS)                 \
	BUSTED_ARGS="$(BUSTED_ARGS)" \
	build/go-in-docker.sh build/test-lua.sh

.PHONY: e2e-test
e2e-test:
	@$(DEF_VARS)                             \
	FOCUS=$(FOCUS)                           \
	E2E_NODES=$(E2E_NODES)                   \
	DOCKER_OPTS="-i --net=host"              \
	NODE_IP=$(NODE_IP)                       \
	SLOW_E2E_THRESHOLD=$(SLOW_E2E_THRESHOLD) \
	build/go-in-docker.sh build/e2e-tests.sh

.PHONY: cover
cover:
	@$(DEF_VARS)                 \
	DOCKER_OPTS="-i --net=host"  \
	build/go-in-docker.sh build/cover.sh

	echo "Uploading coverage results..."
	@curl -s https://codecov.io/bash | bash

.PHONY: vet
vet:
	@go vet $(shell go list ${PKG}/... | grep -v vendor)

.PHONY: release
release: all-container all-push
	echo "done"

.PHONY: check_dead_links
check_dead_links:
	docker run -t \
	  -v $$PWD:/tmp aledbf/awesome_bot:0.1 \
	  --allow-dupe \
	  --allow-redirect $(shell find $$PWD -mindepth 1 -name "*.md" -printf '%P\n' | grep -v vendor | grep -v Changelog.md)

.PHONY: dep-ensure
dep-ensure:
	dep version || curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
	dep ensure -v
	dep prune -v
	find vendor -name '*_test.go' -delete

.PHONY: dev-env
dev-env:
	@build/dev-env.sh

.PHONY: live-docs
live-docs:
	@docker build --pull -t ingress-nginx/mkdocs build/mkdocs
	@docker run --rm -it -p 3000:3000 -v ${PWD}:/docs ingress-nginx/mkdocs

.PHONY: build-docs
build-docs:
	@docker build --pull -t ingress-nginx/mkdocs build/mkdocs
	@docker run --rm -it -v ${PWD}:/docs ingress-nginx/mkdocs build

.PHONY: misspell
misspell:
	@go get github.com/client9/misspell/cmd/misspell
	misspell \
		-locale US \
		-error \
		cmd/* internal/* deploy/* docs/* design/* test/* README.md
