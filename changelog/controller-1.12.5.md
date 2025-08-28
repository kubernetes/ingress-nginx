# Changelog

### controller-v1.12.5

Images:

* registry.k8s.io/ingress-nginx/controller:v1.12.5@sha256:f4a204a39ce99e7d297c54b02e64e421d872675c5ee29ab1b6edb62d4d69be5c
* registry.k8s.io/ingress-nginx/controller-chroot:v1.12.5@sha256:5bee417e81f5478b166e35b66b62824275fba150cb737adf665ba05c61ff4632

### All changes:

* Images: Trigger controller build. (#13768)
* Chart: Bump Kube Webhook CertGen. (#13764)
* Tests & Docs: Bump images. (#13763)
* Go: Update dependencies. (#13751)
* Images: Remove redundant ModSecurity-nginx patch. (#13748)
* Tests: Add `ssl-session-*` config values tests. (#13746)
* Docs: Bump mkdocs to v9.6.16, fix links. (#13744)
* Docs: Fix default config values and links. (#13739)
* Images: Trigger other builds (2/2). (#13734)
* Images: Trigger other builds (1/2). (#13733)
* Tests: Bump Test Runner to v1.4.1. (#13728)
* Images: Trigger Test Runner build. (#13723)
* Go: Bump to v1.24.6. (#13720)
* Images: Bump NGINX to v1.3.1. (#13717)
* Images: Trigger NGINX build. (#13712)
* Annotations: Quote auth proxy headers. (#13709)
* Go: Update dependencies. (#13702)
* CI: Fix typo. (#13699)
* Chart: Push to OCI registry. (#13696)
* Docs: Remove `X-XSS-Protection` header from hardening guide. (#13687)
* Controller: Fix nil pointer in path validation. (#13682)
* Go: Update dependencies. (#13677)
* NGINX: Disable mimalloc's architecture specific optimizations. (#13670)
* Controller: Fix SSL session ticket path. (#13668)
* Docs: Use HTTPS for NGINX links. (#13664)
* Docs: Fix links and formatting in user guide. (#13662)
* Make: Add `helm-test` target. (#13660)
* Docs: Update prerequisites in `getting-started.md`. (#13658)
* Hack: Bump `golangci-lint` to v2.3.0. (#13656)
* CI: Update KIND to v1.33.2. (#13648)
* Docs: Improve `opentelemetry-trust-incoming-span`. (#13637)
* Go: Update dependencies. (#13626)
* CI: Update Kubernetes to v1.33.3. (#13632)
* Go: Bump to v1.24.5. (#13631)
* Bye bye, v1.11. (#13616)

### Dependency updates:

* Bump the actions group with 3 updates (#13757)
* Bump actions/download-artifact from 4.3.0 to 5.0.0 (#13756)
* Bump github/codeql-action from 3.29.3 to 3.29.5 in the actions group (#13707)
* Bump github/codeql-action from 3.29.2 to 3.29.3 in the actions group across 1 directory (#13644)
* Bump the actions group with 3 updates (#13641)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.12.4...controller-v1.12.5
