# Changelog

### controller-v1.12.8

Images:

* registry.k8s.io/ingress-nginx/controller:v1.12.8@sha256:8f8343060688fb2a85752b7345a988d0d3c890d774e18e80b9e8730756e5b530
* registry.k8s.io/ingress-nginx/controller-chroot:v1.12.8@sha256:07c743429b823dfba7c2e5d399351ef0e43816abab48343ca7c01d00fd6517e3

### All changes:

* Images: Trigger controller build. (#14108)
* Annotations: Respect changes to `auth-proxy-set-headers`. (#14105)
* Images: Bump other images. (#14101)
* Images: Trigger other builds (2/2). (#14096)
* Images: Trigger other builds (1/2). (#14095)
* Tests: Bump Test Runner to v1.4.4. (#14076)
* Images: Trigger Test Runner build. (#14072)
* Images: Bump NGINX to v1.3.4. (#14068)
* Images: Trigger NGINX build. (#14065)
* Store: Handle panics in service deletion handler. (#14058)
* Go: Bump to v1.25.3. (#14045)
* Go: Update dependencies. (#14028)
* Images: Bump Alpine to v3.22.2. (#14025)
* Go: Bump to v1.25.2. (#14021)
* Go: Update dependencies. (#14013)
* Controller: Fix `limit_req_zone` sorting. (#14007)
* Annotations: Fix log format. (#14003)

### Dependency updates:

* Bump actions/download-artifact from 5.0.0 to 6.0.0 (#14086)
* Bump github/codeql-action from 4.30.9 to 4.31.0 in the actions group (#14085)
* Bump actions/upload-artifact from 4.6.2 to 5.0.0 (#14083)
* Bump github.com/onsi/ginkgo/v2 from 2.26.0 to 2.27.1 (#14062)
* Bump github/codeql-action from 4.30.8 to 4.30.9 in the actions group (#14054)
* Bump sigs.k8s.io/controller-runtime from 0.22.2 to 0.22.3 in the go group across 1 directory (#14040)
* Bump actions/dependency-review-action from 4.8.0 to 4.8.1 in the actions group (#14037)
* Bump github/codeql-action from 3.30.6 to 4.30.8 (#14035)
* Bump the actions group with 2 updates (#14016)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.12.7...controller-v1.12.8
