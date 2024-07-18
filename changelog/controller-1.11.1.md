# Changelog

### controller-v1.11.1

Images:

* registry.k8s.io/ingress-nginx/controller:v1.11.1@sha256:e6439a12b52076965928e83b7b56aae6731231677b01e81818bce7fa5c60161a
* registry.k8s.io/ingress-nginx/controller-chroot:v1.11.1@sha256:7cabe4bd7558bfdf5b707976d7be56fd15ffece735d7c90fc238b6eda290fd8d

### All changes:

* Tests: Bump `test-runner` to v20240717-1fe74b5f. (#11647)
* Images: Re-run `test-runner` build. (#11644)
* Images: Trigger `test-runner` build. (#11640)
* Images: Bump `NGINX_BASE` to v0.0.10. (#11638)
* Images: Trigger NGINX build. (#11632)
* bump testing runner (#11627)
* remove modsecurity coreruleset test files from nginx image (#11620)
* unskip the ocsp tests and update images to fix cfssl bug (#11616)
* Fix indent in YAML for example pod (#11610)
* Images: Bump `test-runner`. (#11605)
* Images: Bump `NGINX_BASE` to v0.0.9. (#11602)
* revert module upgrade (#11597)
* Release: Apply changes from `main`. (#11589)
* Mage: Stop mutating release notes. (#11581)
* Images: Bump `kube-webhook-certgen`. (#11584)
* update test runner to latest build (#11558)
* add k8s 1.30 to ci build (#11554)
* update test runner go base to 3.20 (#11552)
* tag new test runner image with new nginx base 0.0.8 (#11551)
* bump NGINX_BASE to v0.0.8 (#11544)
* add ssl patches to nginx-1.25 image for coroutines to work in lua client hello and cert ssl blocks (#11535)
* trigger build for NGINX-1.25 v0.0.8 (#11539)
* bump alpine version to 3.20 to custom-error-pages (#11538)
* fix: Ensure changes in MatchCN annotation are detected (#11529)

### Dependency updates:

* Bump github.com/prometheus/common from 0.54.0 to 0.55.0 (#11621)
* Bump the all group with 5 updates (#11614)
* Bump golang.org/x/crypto from 0.24.0 to 0.25.0 (#11580)
* Bump google.golang.org/grpc from 1.64.0 to 1.65.0 (#11576)
* Bump the all group with 4 updates (#11575)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.11.0...controller-v1.11.1
