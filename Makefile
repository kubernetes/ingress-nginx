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

# Add the following 'help' target to your Makefile
# And add help text after each target name starting with '\#\#'

.DEFAULT_GOAL:=help

.EXPORT_ALL_VARIABLES:

ifndef VERBOSE
.SILENT:
endif

# set default shell
SHELL = /bin/bash

# Use the 0.0 tag for testing, it shouldn't clobber any release builds
TAG ?= master
DOCKER ?= docker

# Use docker to run makefile tasks
USE_DOCKER ?= true

# Disable run docker tasks if running in prow.
# only checks the existence of the variable, not the value.
ifdef DIND_TASKS
USE_DOCKER=false
endif

# e2e settings
# Allow limiting the scope of the e2e tests. By default run everything
FOCUS ?= .*
# number of parallel test
E2E_NODES ?= 10
# slow test only if takes > 50s
SLOW_E2E_THRESHOLD ?= 50
# run e2e test suite with tests that check for memory leaks? (default is false)
E2E_CHECK_LEAKS ?=

REPO_INFO ?= $(shell git config --get remote.origin.url)
GIT_COMMIT ?= git-$(shell git rev-parse --short HEAD)

PKG = k8s.io/ingress-nginx

ALL_ARCH = amd64 arm arm64

BUSTED_ARGS =-v --pattern=_test

ARCH ?= $(shell go env GOARCH)

REGISTRY ?= quay.io/kubernetes-ingress-controller
MULTI_ARCH_IMAGE = $(REGISTRY)/nginx-ingress-controller-${ARCH}

GOHOSTOS ?= $(shell go env GOHOSTOS)
GOARCH = ${ARCH}
GOOS = linux
GOBUILD_FLAGS := -v

# fix sed for osx
SED_I ?= sed -i
ifeq ($(GOHOSTOS),darwin)
  SED_I=sed -i ''
endif

# use vendor directory instead of go modules https://github.com/golang/go/wiki/Modules
GO111MODULE=off

# Set default base image dynamically for each arch
BASEIMAGE?=quay.io/kubernetes-ingress-controller/nginx-$(ARCH):26f574dc279aa853736d7f7249965e90e47171d6

TEMP_DIR := $(shell mktemp -d)
DOCKERFILE := $(TEMP_DIR)/rootfs/Dockerfile

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# internal task
.PHONY: sub-container-%
sub-container-%:
	$(MAKE) ARCH=$* build container

# internal task
.PHONY: sub-push-%
sub-push-%: ## Publish image for a particular arch.
	$(MAKE) ARCH=$* push

.PHONY: all-container
all-container: $(addprefix sub-container-,$(ALL_ARCH)) ## Build image for amd64, arm and arm64.

.PHONY: all-push
all-push: $(addprefix sub-push-,$(ALL_ARCH)) ## Publish images for amd64, arm and arm64.

.PHONY: container
container: clean-container .container-$(ARCH) ## Build image for a particular arch.

# internal task to build image for a particular arch.
.PHONY: .container-$(ARCH)
.container-$(ARCH):
	mkdir -p $(TEMP_DIR)/rootfs
	cp bin/$(ARCH)/nginx-ingress-controller $(TEMP_DIR)/rootfs/nginx-ingress-controller
	cp bin/$(ARCH)/dbg $(TEMP_DIR)/rootfs/dbg
	cp bin/$(ARCH)/wait-shutdown $(TEMP_DIR)/rootfs/wait-shutdown

	cp -RP ./* $(TEMP_DIR)
	$(SED_I) "s|BASEIMAGE|$(BASEIMAGE)|g" $(DOCKERFILE)
	$(SED_I) "s|VERSION|$(TAG)|g" $(DOCKERFILE)

	echo "Building docker image..."
	$(DOCKER) buildx build \
		--no-cache \
		--load \
		--progress plain \
		--platform linux/$(ARCH) \
		-t $(MULTI_ARCH_IMAGE):$(TAG) $(TEMP_DIR)/rootfs

.PHONY: clean-container
clean-container: ## Removes local image
	@$(DOCKER) rmi -f $(MULTI_ARCH_IMAGE):$(TAG) || true

.PHONY: push
push: .push-$(ARCH) ## Publish image for a particular arch.

# internal task
.PHONY: .push-$(ARCH)
.push-$(ARCH):
	$(DOCKER) push $(MULTI_ARCH_IMAGE):$(TAG)

.PHONY: build
build: check-go-version ## Build ingress controller, debug tool and pre-stop hook.
ifeq ($(USE_DOCKER), true)
	@build/run-in-docker.sh \
		PKG=$(PKG) \
		ARCH=$(ARCH) \
		GIT_COMMIT=$(GIT_COMMIT) \
		REPO_INFO=$(REPO_INFO) \
		TAG=$(TAG) \
		GOBUILD_FLAGS=$(GOBUILD_FLAGS) \
		build/build.sh
else
	@build/build.sh
endif

.PHONY: build-plugin
build-plugin: check-go-version ## Build ingress-nginx krew plugin.
ifeq ($(USE_DOCKER), true)
	@build/run-in-docker.sh \
		PKG=$(PKG) \
		ARCH=$(ARCH) \
		GIT_COMMIT=$(GIT_COMMIT) \
		REPO_INFO=$(REPO_INFO) \
		TAG=$(TAG) \
		GOBUILD_FLAGS=$(GOBUILD_FLAGS) \
		build/build-plugin.sh
else
	@build/build-plugin.sh
endif

.PHONY: clean
clean: ## Remove .gocache directory.
	rm -rf bin/ .gocache/

.PHONY: static-check
static-check: ## Run verification script for boilerplate, codegen, gofmt, golint and lualint.
ifeq ($(USE_DOCKER), true)
	@build/run-in-docker.sh \
		build/static-check.sh
else
	@build/static-check.sh
endif

.PHONY: test
test: check-go-version ## Run go unit tests.
ifeq ($(USE_DOCKER), true)
	@build/run-in-docker.sh \
		PKG=$(PKG) \
		ARCH=$(ARCH) \
		GIT_COMMIT=$(GIT_COMMIT) \
		REPO_INFO=$(REPO_INFO) \
		TAG=$(TAG) \
		GOBUILD_FLAGS=$(GOBUILD_FLAGS) \
		build/test.sh
else
	@build/test.sh
endif

.PHONY: lua-test
lua-test: ## Run lua unit tests.
ifeq ($(USE_DOCKER), true)
	@build/run-in-docker.sh \
		BUSTED_ARGS=$(BUSTED_ARGS) \
		build/test-lua.sh
else
	@build/test-lua.sh
endif

.PHONY: e2e-test
e2e-test: check-go-version ## Run e2e tests (expects access to a working Kubernetes cluster).
	@build/run-e2e-suite.sh

.PHONY: e2e-test-image
e2e-test-image: e2e-test-binary ## Build image for e2e tests.
	@make -C test/e2e-image

.PHONY: e2e-test-binary
e2e-test-binary: check-go-version ## Build ginkgo binary for e2e tests.
ifeq ($(USE_DOCKER), true)
	@build/run-in-docker.sh \
		ginkgo build ./test/e2e
else
	@ginkgo build ./test/e2e
endif

.PHONY: cover
cover: check-go-version ## Run go coverage unit tests.
	@build/cover.sh
	echo "Uploading coverage results..."
	@curl -s https://codecov.io/bash | bash

.PHONY: vet
vet:
	@go vet $(shell go list ${PKG}/internal/... | grep -v vendor)

.PHONY: release
release: all-container all-push ## Build and publish images of the ingress controller.
	echo "done"

.PHONY: check_dead_links
check_dead_links: ## Check if the documentation contains dead links.
	@docker run -t \
	  -v $$PWD:/tmp aledbf/awesome_bot:0.1 \
	  --allow-dupe \
	  --allow-redirect $(shell find $$PWD -mindepth 1 -name "*.md" -printf '%P\n' | grep -v vendor | grep -v Changelog.md)

.PHONY: dep-ensure
dep-ensure: check-go-version ## Update and vendo go dependencies.
	go mod tidy -v
	find vendor -name '*_test.go' -delete
	go mod vendor

.PHONY: dev-env
dev-env: check-go-version ## Starts a local Kubernetes cluster using minikube, building and deploying the ingress controller.
	@DIND_DOCKER=0 USE_DOCKER=false build/dev-env.sh

.PHONY: live-docs
live-docs: ## Build and launch a local copy of the documentation website in http://localhost:3000
	@docker build --pull -t ingress-nginx/mkdocs images/mkdocs
	@docker run --rm -it -p 3000:3000 -v ${PWD}:/docs ingress-nginx/mkdocs

.PHONY: build-docs
build-docs: ## Build documentation (output in ./site directory).
	@docker build --pull -t ingress-nginx/mkdocs images/mkdocs
	@docker run --rm -v ${PWD}:/docs ingress-nginx/mkdocs build

.PHONY: misspell
misspell: check-go-version ## Check for spelling errors.
	@go get github.com/client9/misspell/cmd/misspell
	misspell \
		-locale US \
		-error \
		cmd/* internal/* deploy/* docs/* design/* test/* README.md

.PHONY: kind-e2e-test
kind-e2e-test: check-go-version ## Run e2e tests using kind.
	@test/e2e/run.sh

.PHONY: run-ingress-controller
run-ingress-controller: ## Run the ingress controller locally using a kubectl proxy connection.
	@build/run-ingress-controller.sh

.PHONY: check-go-version
check-go-version:
	@hack/check-go-version.sh
