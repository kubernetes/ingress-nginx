all: fmt lint vet

BUILDTAGS=

# building inside travis generates a custom version of the
# backends in order to run e2e tests agains the build.
ifdef TRAVIS_BUILD_ID
  RELEASE := ci-build-${TRAVIS_BUILD_ID}
endif

# 0.0 shouldn't clobber any release builds
RELEASE?=0.0

# by default build a linux version
GOOS?=linux

REPO_INFO=$(shell git config --get remote.origin.url)

ifndef COMMIT
  COMMIT := git-$(shell git rev-parse --short HEAD)
endif

# base package. It contains the common and backends code
PKG := "k8s.io/ingress"

GO_LIST_FILES=$(shell go list ${PKG}/... | grep -v vendor | grep -v -e "test/e2e")

.PHONY: fmt
fmt:
	@go list -f '{{if len .TestGoFiles}}"gofmt -s -l {{.Dir}}"{{end}}' ${GO_LIST_FILES} | xargs -L 1 sh -c

.PHONY: lint
lint:
	@go list -f '{{if len .TestGoFiles}}"golint -min_confidence=0.85 {{.Dir}}/..."{{end}}' ${GO_LIST_FILES} | xargs -L 1 sh -c

.PHONY: test
test:
	@go test -v -race -tags "$(BUILDTAGS) cgo" ${GO_LIST_FILES}

.PHONY: test-e2e
test-e2e: ginkgo
	@go run hack/e2e.go -v --up --test --down

.PHONY: cover
cover:
	@go list -f '{{if len .TestGoFiles}}"go test -coverprofile={{.Dir}}/.coverprofile {{.ImportPath}}"{{end}}' ${GO_LIST_FILES} | xargs -L 1 sh -c
	gover
	goveralls -coverprofile=gover.coverprofile -service travis-ci

.PHONY: vet
vet:
	@go vet ${GO_LIST_FILES}

.PHONY: clean
clean:
	make -C controllers/nginx clean

.PHONY: controllers
controllers:
	make -C controllers/nginx build

.PHONY: docker-build
docker-build:
	make -C controllers/nginx all-container

.PHONY: docker-push
docker-push:
	make -C controllers/nginx all-push

.PHONE: release
release:
	make -C controllers/nginx release

.PHONY: ginkgo
ginkgo:
	go get github.com/onsi/ginkgo/ginkgo
