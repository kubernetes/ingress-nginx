# Changelog

### controller-v1.14.4

Images:

* registry.k8s.io/ingress-nginx/controller:v1.14.4@sha256:f8c7959ed0cc0c1dd6060f291fc50ccaf27a5497d182bbb6bc4ffed943616f23
* registry.k8s.io/ingress-nginx/controller-chroot:v1.14.4@sha256:d03d78b6b3a1efa21e02d4d0e9771d07c3e8f0e4a97c5156b12f5c7bd1fc5460

### All changes:

* Images: Trigger controller build. (#14671)
* Template: Quote `proxy_pass`. (#14668)
* Annotations: Consider aliases in risk evaluation. (#14665)
* Images: Bump other images. (#14662)
* Images: Trigger other builds (2/2). (#14659)
* Images: Trigger other builds (1/2). (#14656)
* Tests: Bump Test Runner to v2.2.8. (#14641)
* Images: Trigger Test Runner build. (#14638)
* Go: Update dependencies. (#14635)
* Images: Bump NGINX to v2.2.8. (#14632)
* Images: Trigger NGINX build. (#14629)
* Go: Update dependencies. (#14626)
* Go: Bump to v1.26.1. (#14623)
* CI: Update Kubernetes to v1.35.2. (#14606)
* Admission Controller: Remove obsolete error log. (#14603)
* Mage: Rewrite `updateChartValue` to obsolete outdated libraries. (#14600)
* Go: Update dependencies. (#14596)
* Go: Update dependencies. (#14585)
* CI: Update KIND to v1.35.1. (#14565)
* CI: Update Kubernetes to v1.35.1. (#14562)
* Docs: Clarify PROXY protocol is not supported on GKE default load balancer. (#14559)
* Controller: Enable SSL Passthrough when requested on before HTTP-only hosts. (#14556)
* CI: Update Helm to v4.1.1. (#14553)
* Annotations: Use dedicated regular expression for `proxy-cookie-domain`. (#14550)
* Controller: Use 4KiB buffers for PROXY protocol parsing in TLS passthrough. (#14543)
* Docs: Add retirement notice to website. (#14541)
* Template: Use `RawURLEncoding` instead of `URLEncoding` with padding removal. (#14537)
* Docs: Clarify valid values for `proxy-request-buffering`. (#14534)
* Go: Bump to v1.25.7. (#14526)
* Go: Update dependencies. (#14523)
* Tests: Bump Ginkgo to v2.28.1. (#14520)
* Images: Bump Alpine to v3.23.3. (#14517)
* Lua: Fix type mismatch. (#14514)

### Dependency updates:

* Bump docker/setup-buildx-action from 3.12.0 to 4.0.0 (#14653)
* Bump docker/login-action from 3.7.0 to 4.0.0 (#14652)
* Bump docker/setup-qemu-action from 3.7.0 to 4.0.0 (#14649)
* Bump the actions group with 5 updates (#14647)
* Bump actions/download-artifact from 7.0.0 to 8.0.0 (#14619)
* Bump the go group across 3 directories with 9 updates (#14617)
* Bump actions/upload-artifact from 6.0.0 to 7.0.0 (#14615)
* Bump actions/setup-go from 6.2.0 to 6.3.0 in the actions group (#14614)
* Bump goreleaser/goreleaser-action from 6.4.0 to 7.0.0 (#14593)
* Bump the actions group with 4 updates (#14591)
* Bump google.golang.org/grpc from 1.78.0 to 1.79.1 (#14582)
* Bump the go group across 3 directories with 9 updates (#14580)
* Bump github.com/pires/go-proxyproto from 0.10.0 to 0.11.0 (#14579)
* Bump the actions group with 2 updates (#14577)
* Bump golang.org/x/crypto from 0.47.0 to 0.48.0 (#14575)
* Bump google.golang.org/grpc from 1.78.0 to 1.79.1 in /images/go-grpc-greeter-server/rootfs (#14573)
* Bump github/codeql-action from 4.32.0 to 4.32.2 in the actions group (#14548)
* Bump golang.org/x/oauth2 from 0.34.0 to 0.35.0 (#14546)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.14.3...controller-v1.14.4
