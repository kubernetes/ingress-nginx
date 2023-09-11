# Changelog

### 1.9.0-beta.0
Images:

 * registry.k8s.io/ingress-nginx/controller:v1.9.0-beta.0@sha256:531377e4cc9dc62af40d742402222603259673f5a755a64d74122f256dfad8f9
 * registry.k8s.io/ingress-nginx/controller-chroot:v1.9.0-beta.0@sha256:60b4c95349ce2a81a3b2a76423ee483b847b89d3fa8cb148468434f606f3fa0c
 
### All Changes:

* Start release of v1.9.0 beta0 (#10407)
* Update k8s versions on CI (#10406)
* Add a flag to enable or disable aio_write (#10394)
* Update external-articles.md - advanced setup with GKE/Cloud Armor/IAP (#10372)
* Fix e2e test suite doc (#10396)
* Disable user snippets per default (#10393)
* Deployment/DaemonSet: Fix templating & value. (#10240)
* Fix deferInLoop error (#10387)
* Remove gofmt (#10385)
* Deployment/DaemonSet: Template `topologySpreadConstraints`. (#10259)

### Dependencies updates: 
* Bump github.com/onsi/ginkgo/v2 from 2.9.5 to 2.12.0 (#10355)
* Bump golang.org/x/crypto from 0.12.0 to 0.13.0 (#10399)
* Bump actions/setup-go from 4.0.1 to 4.1.0 (#10403)
* Bump goreleaser/goreleaser-action from 4.4.0 to 4.6.0 (#10402)
* Bump actions/upload-artifact from 3.1.2 to 3.1.3 (#10404)
* Bump golangci/golangci-lint-action from 3.6.0 to 3.7.0 (#10400)
* Bump google.golang.org/grpc from 1.57.0 to 1.58.0 (#10398)
* Bump actions/dependency-review-action from 3.0.8 to 3.1.0 (#10401)
 
**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-controller-v1.8.2...controller-controller-v1.9.0-beta.0
