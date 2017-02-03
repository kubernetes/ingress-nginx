RELEASE=0.9.0-beta.1
PREFIX=localhost/haproxy-ingress-controller

PKG=k8s.io/ingress/controllers/haproxy/pkg
REPO_INFO=$(shell git config --get remote.origin.url)
COMMIT=git-$(shell git rev-parse --short HEAD)

GOOS=linux

build:
	CGO_ENABLED=0 GOOS=$(GOOS) go build \
	  -v -installsuffix cgo \
	  -ldflags "-s -w -X $(PKG)/version.RELEASE=$(RELEASE) -X $(PKG)/version.COMMIT=$(COMMIT) -X $(PKG)/version.REPO=$(REPO_INFO)" \
	  -o rootfs/haproxy-ingress-controller $(PKG)/controller
container: build
	docker build -t $(PREFIX):$(RELEASE) rootfs
