all: push

BUILDTAGS=

# Use the 0.0 tag for testing, it shouldn't clobber any release builds
TAG?=0.9.0-beta.11
REGISTRY?=gcr.io/google_containers
GOOS?=linux
DOCKER?=gcloud docker --
SED_I?=sed -i
GOHOSTOS ?= $(shell go env GOHOSTOS)

ifeq ($(GOHOSTOS),darwin)
  SED_I=sed -i ''
endif

REPO_INFO=$(shell git config --get remote.origin.url)

ifndef COMMIT
  COMMIT := git-$(shell git rev-parse --short HEAD)
endif

PKG=k8s.io/ingress/controllers/nginx

ARCH ?= $(shell go env GOARCH)
GOARCH = ${ARCH}
DUMB_ARCH = ${ARCH}

ALL_ARCH = amd64 arm ppc64le

QEMUVERSION=v2.7.0

IMGNAME = nginx-ingress-controller
IMAGE = $(REGISTRY)/$(IMGNAME)
MULTI_ARCH_IMG = $(IMAGE)-$(ARCH)

# Set default base image dynamically for each arch
BASEIMAGE?=gcr.io/google_containers/nginx-slim-$(ARCH):0.21

ifeq ($(ARCH),arm)
    QEMUARCH=arm
	GOARCH=arm
	DUMB_ARCH=armhf
endif
#ifeq ($(ARCH),arm64)
#        QEMUARCH=aarch64
#endif
ifeq ($(ARCH),ppc64le)
	QEMUARCH=ppc64le
	GOARCH=ppc64le
	DUMB_ARCH=ppc64el
endif
#ifeq ($(ARCH),s390x)
#        QEMUARCH=s390x
#endif

export TEMP_DIR ?= $(shell mktemp -d)

DOCKERFILE := $(TEMP_DIR)/rootfs/Dockerfile

all: all-container

sub-container-%:
	$(MAKE) ARCH=$* build container

sub-push-%:
	$(MAKE) ARCH=$* push

all-container: $(addprefix sub-container-,$(ALL_ARCH))

all-push: $(addprefix sub-push-,$(ALL_ARCH))

container: .container-$(ARCH)
.container-$(ARCH):
	cp -r ./* $(TEMP_DIR)
	$(SED_I) 's|BASEIMAGE|$(BASEIMAGE)|g' $(DOCKERFILE)
	$(SED_I) "s|QEMUARCH|$(QEMUARCH)|g" $(DOCKERFILE)
	$(SED_I) "s|DUMB_ARCH|$(DUMB_ARCH)|g" $(DOCKERFILE)

ifeq ($(ARCH),amd64)
	# When building "normally" for amd64, remove the whole line, it has no part in the amd64 image
	$(SED_I) "/CROSS_BUILD_/d" $(DOCKERFILE)
else
	# When cross-building, only the placeholder "CROSS_BUILD_" should be removed
	# Register /usr/bin/qemu-ARCH-static as the handler for ARM binaries in the kernel
	$(DOCKER) run --rm --privileged multiarch/qemu-user-static:register --reset
	curl -sSL https://github.com/multiarch/qemu-user-static/releases/download/$(QEMUVERSION)/x86_64_qemu-$(QEMUARCH)-static.tar.gz | tar -xz -C $(TEMP_DIR)/rootfs
	$(SED_I) "s/CROSS_BUILD_//g" $(DOCKERFILE)
endif

	$(DOCKER) build -t $(MULTI_ARCH_IMG):$(TAG) $(TEMP_DIR)/rootfs

ifeq ($(ARCH), amd64)
	# This is for to maintain the backward compatibility
	$(DOCKER) tag $(MULTI_ARCH_IMG):$(TAG) $(IMAGE):$(TAG)
endif

push: .push-$(ARCH)
.push-$(ARCH):
	$(DOCKER) push $(MULTI_ARCH_IMG):$(TAG)
ifeq ($(ARCH), amd64)
	$(DOCKER) push $(IMAGE):$(TAG)
endif

clean:
	$(DOCKER) rmi -f $(MULTI_ARCH_IMG):$(TAG) || true

build: clean
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -a -installsuffix cgo \
		-ldflags "-s -w -X ${PKG}/pkg/version.RELEASE=${TAG} -X ${PKG}/pkg/version.COMMIT=${COMMIT} -X ${PKG}/pkg/version.REPO=${REPO_INFO}" \
		-o ${TEMP_DIR}/rootfs/nginx-ingress-controller ${PKG}/pkg/cmd/controller

fmt:
	@echo "+ $@"
	@go list -f '{{if len .TestGoFiles}}"gofmt -s -l {{.Dir}}"{{end}}' $(shell go list ${PKG}/... | grep -v vendor) | xargs -L 1 sh -c

lint:
	@echo "+ $@"
	@go list -f '{{if len .TestGoFiles}}"golint {{.Dir}}/..."{{end}}' $(shell go list ${PKG}/... | grep -v vendor) | xargs -L 1 sh -c

test: fmt lint vet
	@echo "+ $@"
	@go test -v -race -tags "$(BUILDTAGS) cgo" $(shell go list ${PKG}/... | grep -v vendor)

cover:
	@echo "+ $@"
	@go list -f '{{if len .TestGoFiles}}"go test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' $(shell go list ${PKG}/... | grep -v vendor) | xargs -L 1 sh -c
	gover
	goveralls -coverprofile=gover.coverprofile -service travis-ci -repotoken ${COVERALLS_TOKEN}

vet:
	@echo "+ $@"
	@go vet $(shell go list ${PKG}/... | grep -v vendor)

release: all-container all-push
	echo "done"
