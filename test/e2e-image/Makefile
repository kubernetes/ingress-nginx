
DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
E2E_BASE_IMAGE="registry.k8s.io/ingress-nginx/e2e-test-runner:v20230314-helm-chart-4.5.2-32-g520384b11@sha256:754c62f9a5efd1ee515ee908ecc16c0c4d1dda96a8cc8019667182a55f3a9035"

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
