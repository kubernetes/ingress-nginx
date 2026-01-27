
DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
E2E_BASE_IMAGE ?= "registry.k8s.io/ingress-nginx/e2e-test-runner:v2.2.7@sha256:25758395ae93b5ba12c3c9f496f9d592e854286ae1ca275aaa9897f0e88157eb"

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
