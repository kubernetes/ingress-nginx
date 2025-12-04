# Changelog

### controller-v1.14.1

Images:

* registry.k8s.io/ingress-nginx/controller:v1.14.1@sha256:f95a79b85fb93ac3de752c71a5c27d5ceae10a18b61904dec224c1c6a4581e47
* registry.k8s.io/ingress-nginx/controller-chroot:v1.14.1@sha256:29840e06768457b82ef0a9f70bdde03b3b9c42e84a9d78dd6f179326848c1a88

### All changes:

* Images: Trigger controller build. (#14244)
* Images: Bump other images. (#14238)
* Images: Update LuaRocks to v3.12.2. (#14235)
* Images: Trigger other builds (2/2). (#14232)
* Images: Trigger other builds (1/2). (#14229)
* CI: Pin Helm version. (#14226)
* Tests: Bump Test Runner to v2.2.5. (#14222)
* Images: Trigger Test Runner build. (#14219)
* Images: Bump NGINX to v2.2.5. (#14216)
* Images: Trigger NGINX build. (#14213)
* Go: Update dependencies. (#14210)
* Docs: Fix typo. (#14186)
* CI: Update Helm to v3.19.2. (#14174)
* Go: Update dependencies. (#14172)
* CI: Update Kubernetes to v1.34.2. (#14170)
* CI: Update Helm to v3.19.1. (#14165)
* Custom Error Pages: Do not write status code too soon. (#14162)
* Images: Bump GCB Docker GCloud to v20251110-7ccd542560. (#14156)
* Go: Update dependencies. (#14159)
* Tests: Bump Ginkgo to v2.27.2. (#14153)
* Go: Bump to v1.25.4. (#14135)
* Controller: Fix host/path overlap detection for multiple rules. (#14132)
* Bye bye, v1.12. (#14124)

### Dependency updates:

* Bump the actions group with 3 updates (#14242)
* Bump google.golang.org/grpc from 1.76.0 to 1.77.0 (#14208)
* Bump google.golang.org/grpc from 1.76.0 to 1.77.0 in /images/go-grpc-greeter-server/rootfs (#14206)
* Bump github.com/prometheus/common from 0.67.2 to 0.67.4 in /images/custom-error-pages/rootfs in the go group across 1 directory (#14204)
* Bump actions/checkout from 5.0.0 to 6.0.0 (#14202)
* Bump the actions group with 3 updates (#14200)
* Bump golang.org/x/crypto from 0.44.0 to 0.45.0 (#14189)
* Bump the actions group with 3 updates (#14183)
* Bump golang.org/x/oauth2 from 0.32.0 to 0.33.0 (#14150)
* Bump golangci/golangci-lint-action from 8.0.0 to 9.0.0 (#14148)
* Bump helm/chart-testing-action from e27de75c91e0f939bbffea4638c3c70430d7b857 to 6ec842c01de15ebb84c8627d2744a0c2f2755c9f (#14146)
* Bump docker/setup-qemu-action from 3.6.0 to 3.7.0 in the actions group (#14144)
* Bump the go group across 1 directory with 4 updates (#14129)
* Bump github/codeql-action from 4.31.0 to 4.31.2 in the actions group (#14125)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.14.0...controller-v1.14.1
