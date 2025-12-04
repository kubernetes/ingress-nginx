# Changelog

### controller-v1.13.5

Images:

* registry.k8s.io/ingress-nginx/controller:v1.13.5@sha256:5b346855be6752fa2a40f91983fa35a0c004b41493f36be6068a3d4350e69db8
* registry.k8s.io/ingress-nginx/controller-chroot:v1.13.5@sha256:eb6665ca10761ac2b5d1b94959b5cf77e2f6d2bb54178fca16c194933b44c770

### All changes:

* Images: Trigger controller build. (#14245)
* Images: Bump other images. (#14239)
* Images: Update LuaRocks to v3.12.2. (#14236)
* Images: Trigger other builds (2/2). (#14233)
* Images: Trigger other builds (1/2). (#14230)
* CI: Pin Helm version. (#14227)
* Tests: Bump Test Runner to v2.2.5. (#14223)
* Images: Trigger Test Runner build. (#14220)
* Images: Bump NGINX to v2.2.5. (#14217)
* Images: Trigger NGINX build. (#14214)
* Go: Update dependencies. (#14211)
* Docs: Fix typo. (#14187)
* CI: Update Helm to v3.19.2. (#14175)
* Go: Update dependencies. (#14173)
* CI: Update Kubernetes to v1.34.2. (#14171)
* CI: Update Helm to v3.19.1. (#14166)
* Custom Error Pages: Do not write status code too soon. (#14163)
* Images: Bump GCB Docker GCloud to v20251110-7ccd542560. (#14157)
* Go: Update dependencies. (#14160)
* Tests: Bump Ginkgo to v2.27.2. (#14154)
* Go: Bump to v1.25.4. (#14134)
* Controller: Fix host/path overlap detection for multiple rules. (#14131)
* Bye bye, v1.12. (#14127)

### Dependency updates:

* Bump the actions group with 3 updates (#14243)
* Bump google.golang.org/grpc from 1.76.0 to 1.77.0 (#14207)
* Bump google.golang.org/grpc from 1.76.0 to 1.77.0 in /images/go-grpc-greeter-server/rootfs (#14205)
* Bump github.com/prometheus/common from 0.67.2 to 0.67.4 in /images/custom-error-pages/rootfs in the go group across 1 directory (#14203)
* Bump actions/checkout from 5.0.0 to 6.0.0 (#14201)
* Bump the actions group with 3 updates (#14199)
* Bump golang.org/x/crypto from 0.44.0 to 0.45.0 (#14190)
* Bump the actions group with 3 updates (#14184)
* Bump golang.org/x/oauth2 from 0.32.0 to 0.33.0 (#14151)
* Bump golangci/golangci-lint-action from 8.0.0 to 9.0.0 (#14149)
* Bump helm/chart-testing-action from e27de75c91e0f939bbffea4638c3c70430d7b857 to 6ec842c01de15ebb84c8627d2744a0c2f2755c9f (#14147)
* Bump docker/setup-qemu-action from 3.6.0 to 3.7.0 in the actions group (#14145)
* Bump the go group across 1 directory with 4 updates (#14130)
* Bump github/codeql-action from 4.31.0 to 4.31.2 in the actions group (#14126)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.13.4...controller-v1.13.5
