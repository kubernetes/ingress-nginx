# Changelog

### controller-v1.12.6

Images:

* registry.k8s.io/ingress-nginx/controller:v1.12.6@sha256:c371fbf42b4f23584ce879d99303463131f4f31612f0875482b983354eeca7e6
* registry.k8s.io/ingress-nginx/controller-chroot:v1.12.6@sha256:7ff9cdb081b18f9431b84d4c3ccd3db9d921ed5f5b7682a45f6a351bfc4ceed4

### All changes:

* Images: Trigger controller build. (#13864)
* Metrics: Fix `nginx_ingress_controller_config_last_reload_successful`. (#13859)
* Chart: Bump Kube Webhook CertGen. (#13858)
* Tests & Docs: Bump images. (#13857)
* Docs: Remove `datadog` ConfigMap options. (#13852)
* Images: Trigger other builds (2/2). (#13849)
* Images: Trigger other builds (1/2). (#13848)
* Tests: Bump Test Runner to v1.4.2. (#13843)
* Images: Trigger Test Runner build. (#13840)
* Images: Bump NGINX to v1.3.2. (#13837)
* Images: Trigger NGINX build. (#13834)
* Go: Update dependencies. (#13829)
* Annotations/AuthTLS: Allow named redirects. (#13820)
* Tests: Bump Ginkgo to v2.25.1. (#13817)
* Docs: Replace no-break spaces (U+A0). (#13814)
* Tests: Bump Ginkgo to v2.25.0. (#13808)
* Tests: Bump Ginkgo to v2.24.0. (#13803)
* Ingresses: Allow `.` in `Exact` and `Prefix` paths. (#13800)
* Tests: Enable default backend access logging tests. (#13789)
* Security: Harden socket creation and validate error code input. (#13786)
* Tests: Enhance SSL Proxy. (#13784)
* Chores: Migrate deprecated `wait.Poll*` to context-aware equivalents. (#13782)
* Go: Update dependencies. (#13779)
* CI: Update Kubernetes to v1.33.4. (#13777)

### Dependency updates:

* Bump the actions group with 3 updates (#13826)
* Bump actions/checkout from 4.3.0 to 5.0.0 (#13797)
* Bump the actions group with 2 updates (#13795)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.12.5...controller-v1.12.6
