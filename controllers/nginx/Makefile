all: push

# 0.0 shouldn't clobber any release builds
TAG = 0.61
PREFIX = gcr.io/google_containers/nginx-ingress-controller

REPO_INFO=$(shell git config --get remote.origin.url)

ifndef VERSION
  VERSION := git-$(shell git rev-parse --short HEAD)
endif

controller: controller.go clean
	CGO_ENABLED=0 GOOS=linux godep go build -a -installsuffix cgo -ldflags \
		"-w -X main.version=${VERSION} -X main.gitRepo=${REPO_INFO}" \
		-o nginx-ingress-controller

container: controller
	docker build -t $(PREFIX):$(TAG) .

push: container
	gcloud docker push $(PREFIX):$(TAG)

clean:
	rm -f nginx-ingress-controller
