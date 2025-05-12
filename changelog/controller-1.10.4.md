# Changelog

### controller-v1.10.4

Images:

* registry.k8s.io/ingress-nginx/controller:v1.10.4@sha256:505b9048c02dde3d6c8667bf0b52aba7b36adf7b03da34c47d5fa312d2d4c6fc
* registry.k8s.io/ingress-nginx/controller-chroot:v1.10.4@sha256:bf71acf6e71830a4470e2183e3bc93c4f006b954f8a05fb434242ef0f8a24858

### All changes:

* Chart: Bump Kube Webhook CertGen & OpenTelemetry. (#11811)
* Images: Trigger controller build. (#11808)
* Tests & Docs: Bump images. (#11804)
* Images: Trigger failed builds. (#11801)
* Images: Trigger other builds. (#11797)
* Controller: Fix panic in alternative backend merging. (#11793)
* Tests: Bump `e2e-test-runner` to v20240812-3f0129aa. (#11791)
* Images: Trigger `test-runner` build. (#11786)
* Images: Bump `NGINX_BASE` to v0.0.12. (#11783)
* Images: Trigger NGINX build. (#11780)
* Cloud Build: Add missing config, remove unused ones. (#11776)
* Generate correct output on NumCPU() when using cgroups2 (#11775)
* Cloud Build: Tweak timeouts. (#11762)
* Cloud Build: Fix substitutions. (#11759)
* Cloud Build: Some chores. (#11756)
* Go: Bump to v1.22.6. (#11748)
* Images: Bump `NGINX_BASE` to v0.0.11. (#11744)
* Images: Trigger NGINX build. (#11736)
* docs: update OpenSSL Roadmap link (#11734)
* Go: Bump to v1.22.5. (#11731)
* Docs: Fix typo in AWS LB Controller reference (#11724)
* Perform some cleaning operations on line breaks. (#11722)
* Missing anchors in regular expression. (#11718)
* Docs: Fix `from-to-www` redirect description. (#11715)
* Chart: Remove `isControllerTagValid`. (#11714)
* Tests: Bump `e2e-test-runner` to v20240729-04899b27. (#11704)
* Docs: Clarify `from-to-www` redirect direction. (#11692)
* added real-client-ip faq (#11665)
* Docs: Format NGINX configuration table. (#11660)

### Dependency updates:

* Bump github.com/onsi/ginkgo/v2 from 2.19.1 to 2.20.0 (#11772)
* Bump the all group with 2 updates (#11770)
* Bump golang.org/x/crypto from 0.25.0 to 0.26.0 (#11768)
* Bump the all group with 3 updates (#11729)
* Bump github.com/onsi/ginkgo/v2 from 2.19.0 to 2.19.1 in the all group (#11700)
* Bump the all group with 2 updates (#11697)
* Bump the all group with 4 updates (#11676)
* Bump the all group with 2 updates (#11674)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.10.3...controller-v1.10.4
