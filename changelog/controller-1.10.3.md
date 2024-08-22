# Changelog

### controller-v1.10.3

Images:

* registry.k8s.io/ingress-nginx/controller:v1.10.3@sha256:b5a5082f8e508cc1aac1c0ef101dc2f87b63d51598a5747d81d6cf6e7ba058fd
* registry.k8s.io/ingress-nginx/controller-chroot:v1.10.3@sha256:9033e04bd3cd01f92414f8d5999c5095734d4caceb4923942298152a38373d4b

### All changes:

* Images: Trigger `controller` v1.10.3 build. (#11648)
* Tests: Bump `test-runner` to v20240717-1fe74b5f. (#11646)
* Images: Re-run `test-runner` build. (#11643)
* Images: Trigger `test-runner` build. (#11639)
* Images: Bump `NGINX_BASE` to v0.0.10. (#11637)
* Images: Trigger NGINX build. (#11631)
* bump testing runner (#11626)
* remove modsecurity coreruleset test files from nginx image (#11619)
* unskip the ocsp tests and update images to fix cfssl bug (#11615)
* Fix indent in YAML for example pod (#11609)
* Images: Bump `test-runner`. (#11604)
* Images: Bump `NGINX_BASE` to v0.0.9. (#11601)
* revert module upgrade (#11595)
* README: Fix support matrix. (#11593)
* Mage: Stop mutating release notes. (#11582)
* Images: Bump `kube-webhook-certgen`. (#11583)

### Dependency updates:

* Bump github.com/prometheus/common from 0.54.0 to 0.55.0 (#11622)
* Bump the all group with 5 updates (#11613)
* Bump golang.org/x/crypto from 0.24.0 to 0.25.0 (#11579)
* Bump google.golang.org/grpc from 1.64.0 to 1.65.0 (#11577)
* Bump the all group with 4 updates (#11574)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.10.2...controller-v1.10.3
