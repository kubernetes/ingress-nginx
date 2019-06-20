all: push

TAG=0.1
PREFIX?=electroma/ingress-demo-
ARCH?=amd64
GOLANG_VERSION=1.9
TEMP_DIR:=$(shell mktemp -d)

build: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o authsvc/authsvc authsvc/authsvc.go
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o echosvc/echosvc echosvc/echosvc.go

container: build
	docker build --pull -t $(PREFIX)authsvc-$(ARCH):$(TAG) authsvc
	docker build --pull -t $(PREFIX)echosvc-$(ARCH):$(TAG) echosvc

push: container
	docker push $(PREFIX)authsvc-$(ARCH):$(TAG)
	docker push $(PREFIX)echosvc-$(ARCH):$(TAG)

clean:
	rm -f authsvc/authsvc echosvc/echosvc

