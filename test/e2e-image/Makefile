.PHONY: all
all: container

DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Use docker to run makefile tasks
USE_DOCKER ?= true

# Disable run docker tasks if running in prow.
# only checks the existence of the variable, not the value.
ifdef DIND_TASKS
USE_DOCKER=false
endif

.PHONY: container
container:
ifeq ($(USE_DOCKER), true)
	@${DIR}/../../build/run-in-docker.sh make e2e-test-binary
else
	@make -C ${DIR}/../../ e2e-test-binary
endif

	cp ../e2e/e2e.test .
	cp ../e2e/wait-for-nginx.sh .
	cp -r ../../deploy/cloud-generic .
	cp -r ../../deploy/cluster-wide .

	docker build -t nginx-ingress-controller:e2e .

.PHONY: clean
clean:
	rm -rf _cache e2e.test kubectl cluster ginkgo
	docker rmi -f nginx-ingress-controller:e2e || true
