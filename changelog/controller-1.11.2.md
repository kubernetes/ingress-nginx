# Changelog

### controller-v1.11.2

Images:

* registry.k8s.io/ingress-nginx/controller:v1.11.2@sha256:d5f8217feeac4887cb1ed21f27c2674e58be06bd8f5184cacea2a69abaf78dce
* registry.k8s.io/ingress-nginx/controller-chroot:v1.11.2@sha256:21b55a2f0213a18b91612a8c0850167e00a8e34391fd595139a708f9c047e7a8

### All changes:

* Chart: Bump Kube Webhook CertGen & OpenTelemetry. (#11812)
* Images: Trigger controller build. (#11807)
* Tests & Docs: Bump images. (#11805)
* Images: Trigger failed builds. (#11802)
* Images: Trigger other builds. (#11798)
* Controller: Fix panic in alternative backend merging. (#11794)
* Tests: Bump `e2e-test-runner` to v20240812-3f0129aa. (#11792)
* Images: Trigger `test-runner` build. (#11787)
* Images: Bump `NGINX_BASE` to v0.0.12. (#11784)
* Images: Trigger NGINX build. (#11781)
* Cloud Build: Add missing config, remove unused ones. (#11777)
* Generate correct output on NumCPU() when using cgroups2 (#11778)
* Cloud Build: Tweak timeouts. (#11763)
* Cloud Build: Fix substitutions. (#11760)
* Cloud Build: Some chores. (#11757)
* Go: Bump to v1.22.6. (#11749)
* Images: Bump `NGINX_BASE` to v0.0.11. (#11743)
* Images: Trigger NGINX build. (#11737)
* docs: update OpenSSL Roadmap link (#11733)
* Go: Bump to v1.22.5. (#11732)
* Docs: Fix typo in AWS LB Controller reference (#11725)
* Perform some cleaning operations on line breaks. (#11721)
* Missing anchors in regular expression. (#11719)
* Docs: Fix `from-to-www` redirect description. (#11716)
* Chart: Remove `isControllerTagValid`. (#11713)
* Tests: Bump `e2e-test-runner` to v20240729-04899b27. (#11705)
* Docs: Clarify `from-to-www` redirect direction. (#11693)
* added real-client-ip faq (#11664)
* Docs: Format NGINX configuration table. (#11662)
* Docs: Update version in `deploy/index.md`. (#11652)

### Dependency updates:

* Bump github.com/onsi/ginkgo/v2 from 2.19.1 to 2.20.0 (#11773)
* Bump the all group with 2 updates (#11771)
* Bump golang.org/x/crypto from 0.25.0 to 0.26.0 (#11769)
* Bump the all group with 3 updates (#11728)
* Bump github.com/onsi/ginkgo/v2 from 2.19.0 to 2.19.1 in the all group (#11701)
* Bump the all group with 2 updates (#11698)
* Bump the all group with 4 updates (#11677)
* Bump the all group with 2 updates (#11675)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.11.1...controller-v1.11.2
