
DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
E2E_BASE_IMAGE ?= "registry.k8s.io/ingress-nginx/e2e-test-runner:v20240717-1fe74b5f@sha256:fce2ca5e683708cf76c6ba759f351fd7a592c5a56841491cf2d08784154ebf75"

image:
	echo "..entered Makefile in /test/e2e-image"
	echo "..calling Make target <<e2e-test-binary>> in /Makefile from inside /test/e2e-image/Makefile"
	make -C $(DIR)/../../ e2e-test-binary
	echo "..done building e2e-test-binary from /test/e2e-image/Makefile"

	cp $(DIR)/../e2e/e2e.test .
	cp $(DIR)/../e2e/wait-for-nginx.sh .
	cp -R $(DIR)/../../charts .

	# TODO: avoid manual copy
	cp -R $(DIR)/../../test/e2e/settings/ocsp/* .

	docker build \
		--build-arg E2E_BASE_IMAGE=$(E2E_BASE_IMAGE) \
		--tag nginx-ingress-controller:e2e .

clean:
	rm -rf _cache e2e.test kubectl cluster ginkgo
	docker rmi -f nginx-ingress-controller:e2e || true


.PHONY: image clean
