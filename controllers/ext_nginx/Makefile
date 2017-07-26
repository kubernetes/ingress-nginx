# Use the 0.0 tag for testing, it shouldn't clobber any release builds
TAG?=0.1
GOOS?=linux
#SED_I?=sed -i
#GOHOSTOS ?= $(shell go env GOHOSTOS)

REPO_INFO=$(shell git config --get remote.origin.url)

ifndef COMMIT
  COMMIT := git-$(shell git rev-parse --short HEAD)
endif

PKG=k8s.io/ingress/controllers/ext_nginx

ARCH ?= $(shell go env GOARCH)
GOARCH = ${ARCH}
DUMB_ARCH = ${ARCH}

ALL_ARCH = amd64 arm ppc64le

TEMP_DIR := $(shell mktemp -d)


clean:
	rm -f rootfs/nginx-ingress-controller
#	$(DOCKER) rmi -f $(MULTI_ARCH_IMG):$(TAG) || true

build: clean
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build -a -installsuffix cgo \
		-ldflags "-s -w -X ${PKG}/pkg/version.RELEASE=${TAG} -X ${PKG}/pkg/version.COMMIT=${COMMIT} -X ${PKG}/pkg/version.REPO=${REPO_INFO}" \
		-o rootfs/nginx-ingress-controller ${PKG}/pkg/cmd/controller

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
