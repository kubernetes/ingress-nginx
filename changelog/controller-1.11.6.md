# Changelog

### controller-v1.11.6

Images:

* registry.k8s.io/ingress-nginx/controller:v1.11.6@sha256:4f04fad99f00e604ab488cf0945b4eaa2a93f603f97d2a45fc610ff0f3cad0f9
* registry.k8s.io/ingress-nginx/controller-chroot:v1.11.6@sha256:596f8b9ae4773d3b00dfd65855561419c7e70ecb23569a7a2998b47e160b4f85

### All changes:

* Images: Trigger controller build. (#13314)
* Chart: Bump Kube Webhook CertGen. (#13312)
* Tests & Docs: Bump images. (#13309)
* Images: Trigger other builds (2/2). (#13294)
* Images: Trigger other builds (1/2). (#13291)
* Tests: Bump Test Runner to v1.3.3. (#13288)
* Go: Update dependencies. (#13284)
* Images: Trigger Test Runner build. (#13270)
* Images: Bump NGINX to v0.3.3. (#13267)
* Images: Trigger NGINX build. (#13263)
* Go: Update dependencies. (#13260)
* CI: Update Kubernetes to v1.32.4. (#13256)
* Docs: How to modify NLB TCP timeout. (#13250)
* Go: Update dependencies. (#13247)
* Docs: Improve formatting in `monitoring.md`. (#13240)
* Docs: Enable metrics in manifest-based deployments. (#13236)
* Tests: Bump Test Runner to v1.3.2. (#13234)
* Images: Trigger Test Runner build. (#13228)
* Images: Bump `NGINX_BASE` to v0.3.2. (#13223)
* Images: Trigger NGINX build. (#13220)
* Go: Update dependencies. (#13211)
* Docs: Fix link in installation instructions. (#13193)
* Go: Update dependencies. (#13151)
* Go: Bump to v1.24.2. (#13150)
* Annotations: Allow ciphers with underscores. (#13141)
* CI: Do not fail fast. (#13131)
* Images: Fix FromAsCasing. (#13126)
* Images: Extract modules. (#13124)
* Plugin: Improve error handling. (#13113)
* Docs: Fix OpenTelemetry listing. (#13108)
* Tests: Fallback to `yq`. (#13091)
* Go: Fix Mage. (#13080)

### Dependency updates:

* Bump actions/download-artifact from 4.2.1 to 4.3.0 in the actions group (#13305)
* Bump the actions group with 2 updates (#13281)
* Bump github.com/onsi/ginkgo/v2 from 2.23.3 to 2.23.4 (#13214)
* Bump the go group across 2 directories with 1 update (#13206)
* Bump github.com/prometheus/client_golang from 1.21.1 to 1.22.0 (#13205)
* Bump github/codeql-action from 3.28.14 to 3.28.15 in the actions group (#13202)
* Bump github.com/prometheus/client_golang from 1.21.1 to 1.22.0 in /images/custom-error-pages/rootfs (#13201)
* Bump golang.org/x/oauth2 from 0.28.0 to 0.29.0 (#13182)
* Bump the go group across 2 directories with 1 update (#13180)
* Bump github.com/fsnotify/fsnotify from 1.8.0 to 1.9.0 (#13178)
* Bump golang.org/x/crypto from 0.36.0 to 0.37.0 (#13176)
* Bump the actions group with 2 updates (#13174)
* Bump goreleaser/goreleaser-action from 6.2.1 to 6.3.0 in the actions group (#13134)
* Bump golangci/golangci-lint-action from 6.5.2 to 7.0.0 (#13122)
* Bump the actions group with 2 updates (#13119)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.11.5...controller-v1.11.6
