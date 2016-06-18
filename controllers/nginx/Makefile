all: push

# 0.0 shouldn't clobber any release builds
TAG = 0.8
PREFIX = gcr.io/google_containers/nginx-ingress-controller

REPO_INFO=$(shell git config --get remote.origin.url)
UPX:=$(shell command -v upx 2> /dev/null)

ifndef VERSION
  VERSION := git-$(shell git rev-parse --short HEAD)
endif

controller: controller.go clean
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags \
		"-s -w -X main.version=${VERSION} -X main.gitRepo=${REPO_INFO}" \
		-o nginx-ingress-controller
ifdef UPX
	# compressing the binary reduces the size at leat a 75%
	upx nginx-ingress-controller
endif

container: controller
	docker build -t $(PREFIX):$(TAG) .

push: container
	gcloud docker push $(PREFIX):$(TAG)

clean:
	rm -f nginx-ingress-controller
