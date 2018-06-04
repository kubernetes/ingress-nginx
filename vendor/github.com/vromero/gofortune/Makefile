# A Self-Documenting Makefile: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

.PHONY: all cover test clean format vet build tools help
.DEFAULT_GOAL := help

SOURCE_FILES?=$$(go list ./... | grep -v /vendor/)
TEST_PATTERN?=.
TEST_OPTIONS?=

GOTOOLS := \
	golang.org/x/tools/cmd/cover \
	golang.org/x/tools/cmd/goimports \
	github.com/pierrre/gotestcover \
	github.com/golang/dep/cmd/dep \
	github.com/alecthomas/gometalinter \
	github.com/goreleaser/goreleaser \
	github.com/spf13/cobra \
	github.com/inconshreveable/mousetrap \
	github.com/spf13/viper \
	github.com/mitchellh/go-homedir

GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

all: clean build ## Perform the typical build lifecycle without releasing

clean: ## Clean
	rm -f gofortune
	rm -Rf dist
	rm -f coverage.txt

test: ## Run tests
	gotestcover $(TEST_OPTIONS) -covermode=atomic -coverprofile=coverage.txt $(SOURCE_FILES) -run $(TEST_PATTERN) -timeout=2m

cover: ## Generate test coverage report
	go tool cover -html=coverage.txt

fmt: ## gofmt and goimports all go files
	find . -name '*.go' -not -wholename './vendor/*' | while read -r file; do gofmt -w -s "$$file"; goimports -w "$$file"; done

build: test
	go build

install: ## Install binaries
	go install

setup: ## Setup the required go tools
	go get -u -v $(GOTOOLS)

help: ## Shows this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


