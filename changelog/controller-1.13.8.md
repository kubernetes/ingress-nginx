# Changelog

### controller-v1.13.8

Images:

* registry.k8s.io/ingress-nginx/controller:v1.13.8@sha256:0e7fad5de70f55c7f5fb61858be5ba6794d61091ad0874e963a61851e43edf99
* registry.k8s.io/ingress-nginx/controller-chroot:v1.13.8@sha256:5269537aba95892bad7849ef06f0d0d9883cc586af74a0408fa8173835bd2ea1

### All changes:

* Images: Trigger controller build. (#14672)
* Template: Quote `proxy_pass`. (#14669)
* Annotations: Consider aliases in risk evaluation. (#14666)
* Images: Bump other images. (#14663)
* Images: Trigger other builds (2/2). (#14660)
* Images: Trigger other builds (1/2). (#14657)
* Tests: Bump Test Runner to v2.2.8. (#14642)
* Images: Trigger Test Runner build. (#14639)
* Go: Update dependencies. (#14636)
* Images: Bump NGINX to v2.2.8. (#14633)
* Images: Trigger NGINX build. (#14630)
* Go: Update dependencies. (#14627)
* Go: Bump to v1.26.1. (#14624)
* CI: Update Kubernetes to v1.35.2. (#14607)
* Admission Controller: Remove obsolete error log. (#14602)
* Mage: Rewrite `updateChartValue` to obsolete outdated libraries. (#14601)
* Go: Update dependencies. (#14597)
* Go: Update dependencies. (#14586)
* CI: Update KIND to v1.35.1. (#14566)
* CI: Update Kubernetes to v1.35.1. (#14563)
* Docs: Clarify PROXY protocol is not supported on GKE default load balancer. (#14560)
* Controller: Enable SSL Passthrough when requested on before HTTP-only hosts. (#14557)
* CI: Update Helm to v4.1.1. (#14554)
* Annotations: Use dedicated regular expression for `proxy-cookie-domain`. (#14551)
* Docs: Add retirement notice to website. (#14542)
* Template: Use `RawURLEncoding` instead of `URLEncoding` with padding removal. (#14538)
* Docs: Clarify valid values for `proxy-request-buffering`. (#14533)
* Go: Bump to v1.25.7. (#14527)
* Go: Update dependencies. (#14524)
* Tests: Bump Ginkgo to v2.28.1. (#14521)
* Images: Bump Alpine to v3.23.3. (#14518)
* Lua: Fix type mismatch. (#14515)

### Dependency updates:

* Bump docker/setup-buildx-action from 3.12.0 to 4.0.0 (#14654)
* Bump docker/login-action from 3.7.0 to 4.0.0 (#14651)
* Bump docker/setup-qemu-action from 3.7.0 to 4.0.0 (#14650)
* Bump the actions group with 5 updates (#14648)
* Bump actions/download-artifact from 7.0.0 to 8.0.0 (#14620)
* Bump the go group across 3 directories with 9 updates (#14618)
* Bump actions/upload-artifact from 6.0.0 to 7.0.0 (#14616)
* Bump actions/setup-go from 6.2.0 to 6.3.0 in the actions group (#14613)
* Bump goreleaser/goreleaser-action from 6.4.0 to 7.0.0 (#14594)
* Bump the actions group with 4 updates (#14592)
* Bump google.golang.org/grpc from 1.78.0 to 1.79.1 (#14583)
* Bump the go group across 3 directories with 9 updates (#14581)
* Bump the actions group with 2 updates (#14578)
* Bump golang.org/x/crypto from 0.47.0 to 0.48.0 (#14576)
* Bump google.golang.org/grpc from 1.78.0 to 1.79.1 in /images/go-grpc-greeter-server/rootfs (#14574)
* Bump github/codeql-action from 4.32.0 to 4.32.2 in the actions group (#14549)
* Bump golang.org/x/oauth2 from 0.34.0 to 0.35.0 (#14547)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.13.7...controller-v1.13.8
