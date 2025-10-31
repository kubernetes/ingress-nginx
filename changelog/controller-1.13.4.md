# Changelog

### controller-v1.13.4

Images:

* registry.k8s.io/ingress-nginx/controller:v1.13.4@sha256:4042ae3c512c5d7bcf9682b0fdff96cd7b46a23dcbe15a762349094cd8087be7
* registry.k8s.io/ingress-nginx/controller-chroot:v1.13.4@sha256:49fc51d0767efb4d5c871bcd9bd70684fdcbdd34f9e4164bdf9c9d890db19791

### All changes:

* Images: Trigger controller build. (#14107)
* Annotations: Respect changes to `auth-proxy-set-headers`. (#14104)
* Images: Bump other images. (#14100)
* Images: Trigger other builds (2/2). (#14094)
* Images: Trigger other builds (1/2). (#14093)
* Tests: Bump Test Runner to v2.2.4. (#14075)
* Images: Trigger Test Runner build. (#14073)
* Images: Bump NGINX to v2.2.4. (#14067)
* Images: Trigger NGINX build. (#14064)
* Store: Handle panics in service deletion handler. (#14057)
* Go: Bump to v1.25.3. (#14044)
* Go: Update dependencies. (#14027)
* Images: Bump Alpine to v3.22.2. (#14024)
* Go: Bump to v1.25.2. (#14020)
* Go: Update dependencies. (#14012)
* Controller: Fix `limit_req_zone` sorting. (#14006)
* Annotations: Fix log format. (#14002)

### Dependency updates:

* Bump actions/download-artifact from 5.0.0 to 6.0.0 (#14087)
* Bump github/codeql-action from 4.30.9 to 4.31.0 in the actions group (#14084)
* Bump actions/upload-artifact from 4.6.2 to 5.0.0 (#14082)
* Bump github.com/onsi/ginkgo/v2 from 2.26.0 to 2.27.1 (#14061)
* Bump github/codeql-action from 4.30.8 to 4.30.9 in the actions group (#14053)
* Bump sigs.k8s.io/controller-runtime from 0.22.2 to 0.22.3 in the go group across 1 directory (#14039)
* Bump actions/dependency-review-action from 4.8.0 to 4.8.1 in the actions group (#14036)
* Bump github/codeql-action from 3.30.6 to 4.30.8 (#14034)
* Bump the actions group with 2 updates (#14015)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.13.3...controller-v1.13.4
