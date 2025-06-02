# Changelog

### controller-v1.12.2

Images:

* registry.k8s.io/ingress-nginx/controller:v1.12.2@sha256:03497ee984628e95eca9b2279e3f3a3c1685dd48635479e627d219f00c8eefa9
* registry.k8s.io/ingress-nginx/controller-chroot:v1.12.2@sha256:a697e2bfa419768315250d079ccbbca45f6099c60057769702b912d20897a574

### All changes:

* Images: Trigger controller build. (#13313)
* Chart: Bump Kube Webhook CertGen. (#13311)
* Tests & Docs: Bump images. (#13308)
* Images: Trigger other builds (2/2). (#13293)
* Images: Trigger other builds (1/2). (#13290)
* Tests: Bump Test Runner to v1.3.3. (#13287)
* Go: Update dependencies. (#13283)
* Images: Trigger Test Runner build. (#13269)
* Images: Bump NGINX to v1.2.3. (#13266)
* Images: Trigger NGINX build. (#13262)
* Go: Update dependencies. (#13259)
* CI: Update Kubernetes to v1.32.4. (#13255)
* Docs: How to modify NLB TCP timeout. (#13249)
* Go: Update dependencies. (#13246)
* Docs: Improve formatting in `monitoring.md`. (#13241)
* Docs: Enable metrics in manifest-based deployments. (#13235)
* Tests: Bump Test Runner to v1.3.2. (#13233)
* Images: Trigger Test Runner build. (#13225)
* Images: Bump `NGINX_BASE` to v1.2.2. (#13222)
* Images: Trigger NGINX build. (#13219)
* Go: Update dependencies. (#13210)
* Docs: Fix link in installation instructions. (#13192)
* Go: Update dependencies. (#13149)
* Go: Bump to v1.24.2. (#13148)
* Annotations: Allow ciphers with underscores. (#13140)
* CI: Do not fail fast. (#13130)
* Images: Fix FromAsCasing. (#13125)
* Images: Extract modules. (#13123)
* Plugin: Improve error handling. (#13112)
* Docs: Fix OpenTelemetry listing. (#13107)
* Tests: Fallback to `yq`. (#13090)
* Go: Fix Mage. (#13078)

### Dependency updates:

* Bump actions/download-artifact from 4.2.1 to 4.3.0 in the actions group (#13304)
* Bump the actions group with 2 updates (#13280)
* Bump github.com/onsi/ginkgo/v2 from 2.23.3 to 2.23.4 (#13213)
* Bump the go group across 2 directories with 1 update (#13207)
* Bump github.com/prometheus/client_golang from 1.21.1 to 1.22.0 (#13204)
* Bump github/codeql-action from 3.28.14 to 3.28.15 in the actions group (#13203)
* Bump github.com/prometheus/client_golang from 1.21.1 to 1.22.0 in /images/custom-error-pages/rootfs (#13200)
* Bump golang.org/x/oauth2 from 0.28.0 to 0.29.0 (#13181)
* Bump the go group across 2 directories with 1 update (#13179)
* Bump github.com/fsnotify/fsnotify from 1.8.0 to 1.9.0 (#13177)
* Bump golang.org/x/crypto from 0.36.0 to 0.37.0 (#13175)
* Bump the actions group with 2 updates (#13173)
* Bump goreleaser/goreleaser-action from 6.2.1 to 6.3.0 in the actions group (#13133)
* Bump golangci/golangci-lint-action from 6.5.2 to 7.0.0 (#13121)
* Bump the actions group with 2 updates (#13118)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.12.1...controller-v1.12.2
