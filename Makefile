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
SHELL=/bin/bash -o pipefail -o errexit

# Use the 0.0 tag for testing, it shouldn't clobber any release builds
TAG ?= $(shell cat TAG)

# e2e settings
# Allow limiting the scope of the e2e tests. By default run everything
FOCUS ?=
# number of parallel test
E2E_NODES ?= 7
# run e2e test suite with tests that check for memory leaks? (default is false)
E2E_CHECK_LEAKS ?=

REPO_INFO ?= $(shell git config --get remote.origin.url)
COMMIT_SHA ?= git-$(shell git rev-parse --short HEAD)
BUILD_ID ?= "UNSET"

PKG = k8s.io/ingress-nginx

HOST_ARCH = $(shell which go >/dev/null 2>&1 && go env GOARCH)
ARCH ?= $(HOST_ARCH)
ifeq ($(ARCH),)
    $(error mandatory variable ARCH is empty, either set it when calling the command or make sure 'go env GOARCH' works)
endif

ifneq ($(PLATFORM),)
	PLATFORM_FLAG="--platform"
endif

REGISTRY ?= gcr.io/k8s-staging-ingress-nginx

BASE_IMAGE ?= $(shell cat NGINX_BASE)

GOARCH=$(ARCH)

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: image
image: clean-image ## Build image for a particular arch.
	echo "Building docker image ($(ARCH))..."
	docker build \
		${PLATFORM_FLAG} ${PLATFORM} \
		--no-cache \
		--pull \
		--build-arg BASE_IMAGE="$(BASE_IMAGE)" \
		--build-arg VERSION="$(TAG)" \
		--build-arg TARGETARCH="$(ARCH)" \
		--build-arg COMMIT_SHA="$(COMMIT_SHA)" \
		--build-arg BUILD_ID="$(BUILD_ID)" \
		-t $(REGISTRY)/controller:$(TAG) rootfs

.PHONY: gosec
gosec:
	docker run --rm -it -w /source/ -v "$(pwd)"/:/source securego/gosec:2.11.0 -exclude=G109,G601,G104,G204,G304,G306,G307 -tests=false -exclude-dir=test -exclude-dir=images/  -exclude-dir=docs/ /source/...

.PHONY: image-chroot
image-chroot: clean-chroot-image ## Build image for a particular arch.
	echo "Building docker image ($(ARCH))..."
	docker build \
		--no-cache \
		--pull \
		--build-arg BASE_IMAGE="$(BASE_IMAGE)" \
		--build-arg VERSION="$(TAG)" \
		--build-arg TARGETARCH="$(ARCH)" \
		--build-arg COMMIT_SHA="$(COMMIT_SHA)" \
		--build-arg BUILD_ID="$(BUILD_ID)" \
		-t $(REGISTRY)/controller-chroot:$(TAG) rootfs -f rootfs/Dockerfile-chroot

.PHONY: clean-image
clean-image: ## Removes local image
	echo "removing old image $(REGISTRY)/controller:$(TAG)"
	@docker rmi -f $(REGISTRY)/controller:$(TAG) || true


.PHONY: clean-chroot-image
clean-chroot-image: ## Removes local image
	echo "removing old image $(REGISTRY)/controller-chroot:$(TAG)"
	@docker rmi -f $(REGISTRY)/controller-chroot:$(TAG) || true


.PHONY: build
build:  ## Build ingress controller, debug tool and pre-stop hook.
	build/run-in-docker.sh \
		MAC_OS=$(MAC_OS) \
		PKG=$(PKG) \
		ARCH=$(ARCH) \
		COMMIT_SHA=$(COMMIT_SHA) \
		REPO_INFO=$(REPO_INFO) \
		TAG=$(TAG) \
		build/build.sh


.PHONY: clean
clean: ## Remove .gocache directory.
	rm -rf bin/ .gocache/ .cache/


.PHONY: static-check
static-check: ## Run verification script for boilerplate, codegen, gofmt, golint, lualint and chart-lint.
	@build/run-in-docker.sh \
	    MAC_OS=$(MAC_OS) \
		hack/verify-all.sh

###############################
# Tests for ingress-nginx
###############################

.PHONY: test
test:  ## Run go unit tests.
	@build/run-in-docker.sh \
		PKG=$(PKG) \
		MAC_OS=$(MAC_OS) \
		ARCH=$(ARCH) \
		COMMIT_SHA=$(COMMIT_SHA) \
		REPO_INFO=$(REPO_INFO) \
		TAG=$(TAG) \
		GOFLAGS="-buildvcs=false" \
		test/test.sh

.PHONY: lua-test
lua-test: ## Run lua unit tests.
	@build/run-in-docker.sh \
		BUSTED_ARGS=$(BUSTED_ARGS) \
		MAC_OS=$(MAC_OS) \
		test/test-lua.sh

.PHONY: e2e-test
e2e-test:  ## Run e2e tests (expects access to a working Kubernetes cluster).
	@test/e2e/run-e2e-suite.sh

.PHONY: kind-e2e-test
kind-e2e-test:  ## Run e2e tests using kind.
	@test/e2e/run-kind-e2e.sh

.PHONY: kind-e2e-chart-tests
kind-e2e-chart-tests: ## Run helm chart e2e tests
	@test/e2e/run-chart-test.sh

.PHONY: e2e-test-binary
e2e-test-binary:  ## Build binary for e2e tests.
	@build/run-in-docker.sh \
		MAC_OS=$(MAC_OS) \
		ginkgo build ./test/e2e

.PHONY: print-e2e-suite
print-e2e-suite: e2e-test-binary ## Prints information about the suite of e2e tests.
	@build/run-in-docker.sh \
		MAC_OS=$(MAC_OS) \
		hack/print-e2e-suite.sh

.PHONY: vet
vet:
	@go vet $(shell go list ${PKG}/internal/... | grep -v vendor)

.PHONY: check_dead_links
check_dead_links: ## Check if the documentation contains dead links.
	@docker run ${PLATFORM_FLAG} ${PLATFORM} -t \
	  -w /tmp \
	  -v $$PWD:/tmp dkhamsing/awesome_bot:1.20.0 \
	  --allow-dupe \
	  --allow-redirect $(shell find $$PWD -mindepth 1 -name vendor -prune -o -name .modcache -prune -o -iname Changelog.md -prune -o -name "*.md" | sed -e "s#$$PWD/##")

.PHONY: dev-env
dev-env:  ## Starts a local Kubernetes cluster using kind, building and deploying the ingress controller.
	@build/dev-env.sh

.PHONY: dev-env-stop
dev-env-stop: ## Deletes local Kubernetes cluster created by kind.
	@kind delete cluster --name ingress-nginx-dev



.PHONY: live-docs
live-docs: ## Build and launch a local copy of the documentation website in http://localhost:8000
	@docker build ${PLATFORM_FLAG} ${PLATFORM} \
                  		--no-cache \
                  		 -t ingress-nginx-docs .github/actions/mkdocs
	@docker run ${PLATFORM_FLAG} ${PLATFORM} --rm -it \
		-p 8000:8000 \
		-v ${PWD}:/docs \
		--entrypoint mkdocs \
		ingress-nginx-docs serve --dev-addr=0.0.0.0:8000

.PHONY: misspell
misspell:  ## Check for spelling errors.
	@go install github.com/client9/misspell/cmd/misspell@latest
	misspell \
		-locale US \
		-error \
		cmd/* internal/* deploy/* docs/* design/* test/* README.md

.PHONY: run-ingress-controller
run-ingress-controller: ## Run the ingress controller locally using a kubectl proxy connection.
	@build/run-ingress-controller.sh

.PHONY: ensure-buildx
ensure-buildx:
	./hack/init-buildx.sh

.PHONY: show-version
show-version:
	echo -n $(TAG)

PLATFORMS ?= amd64 arm arm64 s390x
BUILDX_PLATFORMS ?= linux/amd64,linux/arm,linux/arm64,linux/s390x

.PHONY: release # Build a multi-arch docker image
release: ensure-buildx clean
	echo "Building binaries..."
	$(foreach PLATFORM,$(PLATFORMS), echo -n "$(PLATFORM)..."; ARCH=$(PLATFORM) make build;)

	echo "Building and pushing ingress-nginx image...$(BUILDX_PLATFORMS)"

	docker buildx build \
		--no-cache \
		$(MAC_DOCKER_FLAGS) \
		--push \
		--pull \
		--progress plain \
		--platform $(BUILDX_PLATFORMS) \
		--build-arg BASE_IMAGE="$(BASE_IMAGE)" \
		--build-arg VERSION="$(TAG)" \
		--build-arg COMMIT_SHA="$(COMMIT_SHA)" \
		--build-arg BUILD_ID="$(BUILD_ID)" \
		-t $(REGISTRY)/controller:$(TAG) rootfs

	docker buildx build \
		--no-cache \
		$(MAC_DOCKER_FLAGS) \
		--push \
		--pull \
		--progress plain \
		--platform $(BUILDX_PLATFORMS)  \
		--build-arg BASE_IMAGE="$(BASE_IMAGE)" \
		--build-arg VERSION="$(TAG)" \
		--build-arg COMMIT_SHA="$(COMMIT_SHA)" \
		--build-arg BUILD_ID="$(BUILD_ID)" \
		-t $(REGISTRY)/controller-chroot:$(TAG) rootfs -f rootfs/Dockerfile-chroot
