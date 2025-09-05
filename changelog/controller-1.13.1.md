# Changelog

### controller-v1.13.1

Images:

* registry.k8s.io/ingress-nginx/controller:v1.13.1@sha256:37e489b22ac77576576e52e474941cd7754238438847c1ee795ad6d59c02b12a
* registry.k8s.io/ingress-nginx/controller-chroot:v1.13.1@sha256:cace9bc8ad1914e817e5b461d691a00caab652347002ba811077189b85009d7f

### All changes:

* Images: Trigger controller build. (#13767)
* Chart: Bump Kube Webhook CertGen. (#13762)
* Tests & Docs: Bump images. (#13761)
* Go: Update dependencies. (#13750)
* Images: Remove redundant ModSecurity-nginx patch. (#13747)
* Tests: Add `ssl-session-*` config values tests. (#13745)
* Docs: Bump mkdocs to v9.6.16, fix links. (#13743)
* Docs: Fix default config values and links. (#13738)
* Images: Trigger other builds (2/2). (#13735)
* Images: Trigger other builds (1/2). (#13731)
* Tests: Bump Test Runner to v2.2.1. (#13727)
* Images: Trigger Test Runner build. (#13722)
* Go: Bump to v1.24.6. (#13719)
* Images: Bump NGINX to v2.2.1. (#13716)
* Images: Trigger NGINX build. (#13713)
* Annotations: Quote auth proxy headers. (#13708)
* Go: Update dependencies. (#13701)
* CI: Fix typo. (#13698)
* Chart: Push to OCI registry. (#13695)
* Docs: Remove `X-XSS-Protection` header from hardening guide. (#13686)
* Controller: Fix nil pointer in path validation. (#13681)
* Go: Update dependencies. (#13676)
* NGINX: Disable mimalloc's architecture specific optimizations. (#13671)
* Controller: Fix SSL session ticket path. (#13667)
* Docs: Use HTTPS for NGINX links. (#13663)
* Docs: Fix links and formatting in user guide. (#13661)
* Make: Add `helm-test` target. (#13659)
* Docs: Update prerequisites in `getting-started.md`. (#13657)
* Hack: Bump `golangci-lint` to v2.3.0. (#13655)
* CI: Update KIND to v1.33.2. (#13647)
* Config/Annotations: Fix `proxy-busy-buffers-size`. (#13638)
* Docs: Improve `opentelemetry-trust-incoming-span`. (#13636)
* Chart: Remove trailing whitespace. (#13634)
* Go: Update dependencies. (#13625)
* CI: Update Kubernetes to v1.33.3. (#13630)
* Go: Bump to v1.24.5. (#13629)
* Bye bye, v1.11. (#13615)

### Dependency updates:

* Bump the actions group with 3 updates (#13758)
* Bump actions/download-artifact from 4.3.0 to 5.0.0 (#13755)
* Bump github/codeql-action from 3.29.3 to 3.29.5 in the actions group (#13706)
* Bump github/codeql-action from 3.29.2 to 3.29.3 in the actions group across 1 directory (#13643)
* Bump the actions group with 3 updates (#13640)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.13.0...controller-v1.13.1
