IMAGE=nginx-ingress-controller:e2e
KUBE_VERSION ?= 1.13.3

.PHONY: all container getbins clean

all: container

container:
	./kubectl > /dev/null 2>&1 || curl -Lo ./kubectl \
		https://storage.googleapis.com/kubernetes-release/release/v$(KUBE_VERSION)/bin/linux/amd64/kubectl \
  		&& chmod +x ./kubectl

	$(GOPATH)/bin/ginkgo > /dev/null 2>&1 || go get github.com/onsi/ginkgo/ginkgo
	cp $(GOPATH)/bin/ginkgo .

	cp ../e2e/e2e.test .
	cp ../e2e/wait-for-nginx.sh .

	docker build -t $(IMAGE) .

clean:
	rm -rf _cache e2e.test kubectl cluster ginkgo
	docker rmi -f $(IMAGE) || true
