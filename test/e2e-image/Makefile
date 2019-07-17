.PHONY: all
all: container

.PHONY: container
container:
	make -C ../../ e2e-test-binary

	cp ../e2e/e2e.test .
	cp ../e2e/wait-for-nginx.sh .
	cp -r ../../deploy/cloud-generic .
	cp -r ../../deploy/cluster-wide .

	docker build -t nginx-ingress-controller:e2e .

.PHONY: clean
clean:
	rm -rf _cache e2e.test kubectl cluster ginkgo
	docker rmi -f nginx-ingress-controller:e2e || true
