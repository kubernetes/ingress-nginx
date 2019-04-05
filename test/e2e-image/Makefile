IMAGE=nginx-ingress-controller:e2e

.PHONY: all container getbins clean

all: container

container:
	make -C ../../ e2e-test-binary

	cp ../e2e/e2e.test .
	cp ../e2e/wait-for-nginx.sh .

	docker build -t $(IMAGE) .

clean:
	rm -rf _cache e2e.test kubectl cluster ginkgo
	docker rmi -f $(IMAGE) || true
