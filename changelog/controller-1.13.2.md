# Changelog

### controller-v1.13.2

Images:

* registry.k8s.io/ingress-nginx/controller:v1.13.2@sha256:1f7eaeb01933e719c8a9f4acd8181e555e582330c7d50f24484fb64d2ba9b2ef
* registry.k8s.io/ingress-nginx/controller-chroot:v1.13.2@sha256:2beb2139c53d6bcb9c8b11d68b412a6a1aa1de3a7e6040695848b0ce997b2be8

### All changes:

* Images: Trigger controller build. (#13863)
* Metrics: Fix `nginx_ingress_controller_config_last_reload_successful`. (#13860)
* Chart: Bump Kube Webhook CertGen. (#13856)
* Tests & Docs: Bump images. (#13855)
* Docs: Remove `datadog` ConfigMap options. (#13851)
* Images: Trigger other builds (2/2). (#13847)
* Images: Trigger other builds (1/2). (#13846)
* Tests: Bump Test Runner to v2.2.2. (#13842)
* Images: Trigger Test Runner build. (#13839)
* Images: Bump NGINX to v2.2.2. (#13836)
* Images: Trigger NGINX build. (#13833)
* Go: Update dependencies. (#13828)
* Annotations/AuthTLS: Allow named redirects. (#13819)
* Tests: Bump Ginkgo to v2.25.1. (#13816)
* Docs: Replace no-break spaces (U+A0). (#13813)
* Tests: Bump Ginkgo to v2.25.0. (#13807)
* Tests: Bump Ginkgo to v2.24.0. (#13802)
* Ingresses: Allow `.` in `Exact` and `Prefix` paths. (#13799)
* Config/Annotations: Remove `proxy-busy-buffers-size` default value. (#13790)
* Tests: Enable default backend access logging tests. (#13788)
* Security: Harden socket creation and validate error code input. (#13785)
* Tests: Enhance SSL Proxy. (#13783)
* Chores: Migrate deprecated `wait.Poll*` to context-aware equivalents. (#13781)
* Go: Update dependencies. (#13778)
* CI: Update Kubernetes to v1.33.4. (#13776)

### Dependency updates:

* Bump the actions group with 3 updates (#13825)
* Bump actions/checkout from 4.3.0 to 5.0.0 (#13796)
* Bump the actions group with 2 updates (#13794)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.13.1...controller-v1.13.2
