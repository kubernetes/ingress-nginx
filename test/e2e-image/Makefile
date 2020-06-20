.PHONY: all
all: image

DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# Use docker to run makefile tasks
USE_DOCKER ?= true

# Disable run docker tasks if running in prow.
# only checks the existence of the variable, not the value.
ifdef DIND_TASKS
USE_DOCKER=false
endif

.PHONY: image
image:
ifeq ($(USE_DOCKER), true)
	@$(DIR)/../../build/run-in-docker.sh make e2e-test-binary
else
	@make -C $(DIR)/../../ e2e-test-binary
endif

	cp $(DIR)/../e2e/e2e.test .
	cp $(DIR)/../e2e/wait-for-nginx.sh .
	cp -R $(DIR)/../../charts .

	# TODO: avoid manual copy
	cp -R $(DIR)/../../test/e2e/settings/ocsp/* .

	docker build \
		--tag nginx-ingress-controller:e2e .

.PHONY: clean
clean:
	rm -rf _cache e2e.test kubectl cluster ginkgo
	docker rmi -f nginx-ingress-controller:e2e || true
