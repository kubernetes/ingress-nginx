# Changelog

### controller-v1.12.0

Images:

* registry.k8s.io/ingress-nginx/controller:v1.12.0@sha256:e6b8de175acda6ca913891f0f727bca4527e797d52688cbe9fec9040d6f6b6fa
* registry.k8s.io/ingress-nginx/controller-chroot:v1.12.0@sha256:87c88e1c38a6c8d4483c8f70b69e2cca49853bb3ec3124b9b1be648edf139af3

### All changes:

* Images: Trigger controller build. (#12609)
* Chart: Bump Kube Webhook CertGen. (#12606)
* Tests & Docs: Bump images. (#12603)
* Images: Trigger other builds (2/2). (#12599)
* Images: Trigger other builds (1/2). (#12595)
* Tests: Bump `e2e-test-runner` to v20241224-68ed4e7b. (#12590)
* Images: Trigger `test-runner` build. (#12587)
* Images: Bump `NGINX_BASE` to v1.1.0. (#12582)
* Images: Trigger NGINX build. (#12579)
* Go: Clean `go.work.sum`. (#12573)
* Repository: Update owners. (#12568)
* Images: Bump `gcb-docker-gcloud` to v20241217-ff46a068cd. (#12561)
* CI: Update KIND images. (#12560)
* Images: Bump Alpine to v3.21. (#12528)
* Docs: Add guide on how to set a Maintenance Page. (#12525)
* rikatz is stepping down (#12516)
* rikatz is stepping down (#12494)
* Go: Bump to v1.23.4. (#12483)
* Plugin: Bump `goreleaser` to v2. (#12440)
* GitHub: Fix `exec` in issue template. (#12387)
* CI: Update KIND images. (#12367)
* Images: Bump `gcb-docker-gcloud` to v20241110-72bb0b1665. (#12342)
* Go: Bump to v1.23.3. (#12337)
* Auth TLS: Add `_` to redirect RegEx. (#12326)
* Auth TLS: Improve redirect RegEx. (#12323)
* Update custom headers annotation documentation (#12318)
* Tests: Bump `e2e-test-runner` to v20241104-02a3933e. (#12312)
* Docs: Add CPU usage note for `--metrics-per-undefined-host`. (#12310)
* Images: Trigger `test-runner` build. (#12308)
* Config: Fix panic on invalid `lua-shared-dict`. (#12283)
* Docs: fix limit-rate-after references (#12278)
* Chart: Rework ServiceMonitor. (#12269)
* Chart: Add ServiceAccount tests. (#12263)
* CI: Fix chart testing. (#12258)
* [fix] fix nginx temp configs cleanup (#12225)
* Chart: Suggest `matchLabelKeys` in Topology Spread Constraints. (#12202)
* Docs: Add Pod Security Admission. (#12195)
* Docs: Clarify external & service port in TCP/UDP services explanation. (#12192)
* Images: Trigger controller build. (#12154)
* ⚠️ Metrics: Disable by default. (#12153) ⚠️

  This changes the default of the following CLI arguments:

  * `--enable-metrics` gets disabled by default.

* Tests & Docs: Bump `e2e-test-echo` to v1.0.1. (#12147)
* Images: Trigger `e2e-test-echo` build. (#12140)
* ⚠️ Images: Drop `s390x`. (#12137) ⚠️

  Support for the `s390x` architecture has already been removed from the controller image. This also removes it from the NGINX base image and CI relevant images.

* Images: Build `s390x` controller. (#12126)
* Chart: Bump Kube Webhook CertGen. (#12119)
* Tests & Docs: Bump images. (#12118)
* Cloud Build: Bump `gcb-docker-gcloud` to v20240718-5ef92b5c36. (#12113)
* Images: Trigger other builds. (#12110)
* Tests: Bump `e2e-test-runner` to v20241004-114a6abb. (#12103)
* Images: Trigger `test-runner` build. (#12100)
* Docs: Add a multi-tenant warning. (#12091)
* Go: Bump to v1.22.8. (#12069)
* Images: Bump `NGINX_BASE` to v1.0.0. (#12066)
* Images: Trigger NGINX build. (#12063)
* Images: Remove NGINX v1.21. (#12031)
* Chart: Add `controller.metrics.service.enabled`. (#12056)
* GitHub: Improve Dependabot. (#12033)
* Chart: Add `global.image.registry`. (#12028)
* ⚠️ Images: Remove OpenTelemetry. (#12024) ⚠️

  OpenTelemetry is still supported, but since the module is built into the controller image since v1.10, we hereby remove the init container and image which were used to install it upon controller startup.

* Chart: Improve CI. (#12003)
* Chart: Extend image tests. (#12025)
* Chart: Add `controller.progressDeadlineSeconds`. (#12017)
* Docs: Add health check annotations for AWS. (#12018)
* Docs: Convert `opentelemetry.md` from CRLF to LF. (#12005)
* Chart: Implement `unhealthyPodEvictionPolicy`. (#11992)
* Chart: Add `defaultBackend.maxUnavailable`. (#11995)
* Chart: Test `controller.minAvailable` & `controller.maxUnavailable`. (#12000)
* Chart: Align default backend `PodDisruptionBudget`. (#11993)
* Metrics: Fix namespace in `nginx_ingress_controller_ssl_expire_time_seconds`. (#10274)
* ⚠️ Chart: Remove Pod Security Policy. (#11971) ⚠️

  This removes Pod Security Policies and related resources from the chart.

* Chart: Improve default backend service account. (#11972)
* Go: Bump to v1.22.7. (#11943)
* NGINX: Remove inline Lua from template. (#11806)
* Images: Bump OpenTelemetry C++ Contrib. (#11629)
* Docs: Add note about `--watch-namespace`. (#11947)
* Images: Use latest Alpine 3.20 everywhere. (#11944)
* Fix minor typos (#11935)
* Chart: Implement `controller.admissionWebhooks.service.servicePort`. (#11931)
* Allow any protocol for cors origins (#11153)
* Tests: Bump `e2e-test-runner` to v20240829-2c421762. (#11919)
* Images: Trigger `test-runner` build. (#11916)
* Chart: Add `controller.metrics.prometheusRule.annotations`. (#11849)
* Chart: Add tests for `PrometheusRule` & `ServiceMonitor`. (#11883)
* Annotations: Allow commas in URLs. (#11882)
* CI: Grant checks write permissions to E2E Test Report. (#11862)
* Chart: Use generic values for `ConfigMap` test. (#11877)
* Security: Follow-up on recent changes. (#11874)
* Lua: Remove plugins from `.luacheckrc` & E2E docs. (#11872)
* Dashboard: Remove `ingress_upstream_latency_seconds`. (#11878)
* Metrics: Add `--metrics-per-undefined-host` argument. (#11818)
* Update maxmind post link about geolite2 license changes (#11861)
* ⚠️ Remove global-rate-limit feature (#11851) ⚠️

  This removes the following configuration options:

  * `global-rate-limit-memcached-host`
  * `global-rate-limit-memcached-port`
  * `global-rate-limit-memcached-connect-timeout`
  * `global-rate-limit-memcached-max-idle-timeout`
  * `global-rate-limit-memcached-pool-size`
  * `global-rate-limit-status-code`

  It also removes the following annotations:

  * `global-rate-limit`
  * `global-rate-limit-window`
  * `global-rate-limit-key`
  * `global-rate-limit-ignored-cidrs`

* Revert "docs: Add deployment for AWS NLB Proxy." (#11857)
* Add custom code handling for temporal redirect (#10651)
* Add native histogram support for histogram metrics (#9971)
* Replace deprecated queue method (#11853)
* ⚠️ Enable security features by default (#11819) ⚠️

  This changes the default of the following CLI arguments:

  * `--enable-annotation-validation` gets enabled by default.

  It also changes the default of the following configuration options:

  * `allow-cross-namespace-resources` gets disabled by default.
  * `annotations-risk-level` gets lowered to "High" by default.
  * `strict-validate-path-type` gets enabled by default.

* docs: Add deployment for AWS NLB Proxy. (#9565)
* ⚠️ Remove 3rd party lua plugin support (#11821) ⚠️

  This removes the following configuration options:

  * `plugins`

  It also removes support for user provided Lua plugins in the `/etc/nginx/lua/plugins` directory.

* Auto-generate annotation docs (#11820)
* ⚠️ Metrics: Remove `ingress_upstream_latency_seconds`. (#11795) ⚠️

  This metric has already been deprecated and is now getting removed.

* Release controller v1.11.2/v1.10.4 & chart v4.11.2/v4.10.4. (#11816)
* Chart: Bump Kube Webhook CertGen & OpenTelemetry. (#11809)
* Tests & Docs: Bump images. (#11803)
* Images: Trigger failed builds. (#11800)
* Images: Trigger other builds. (#11796)
* Controller: Fix panic in alternative backend merging. (#11789)
* Tests: Bump `e2e-test-runner` to v20240812-3f0129aa. (#11788)
* Images: Trigger `test-runner` build. (#11785)
* Images: Bump `NGINX_BASE` to v0.0.12. (#11782)
* Images: Trigger NGINX build. (#11779)
* Cloud Build: Add missing config, remove unused ones. (#11774)
* Cloud Build: Tweak timeouts. (#11761)
* Cloud Build: Fix substitutions. (#11758)
* Cloud Build: Some chores. (#11633)
* Go: Bump to v1.22.6. (#11747)
* Images: Bump `NGINX_BASE` to v0.0.11. (#11741)
* Images: Trigger NGINX build. (#11735)
* docs: update OpenSSL Roadmap link (#11730)
* Go: Bump to v1.22.5. (#11634)
* Docs: Fix typo in AWS LB Controller reference (#11723)
* Perform some cleaning operations on line breaks. (#11720)
* Missing anchors in regular expression. (#11717)
* Docs: Fix `from-to-www` redirect description. (#11712)
* Chart: Remove `isControllerTagValid`. (#11710)
* Tests: Bump `e2e-test-runner` to v20240729-04899b27. (#11702)
* Chart: Explicitly set `runAsGroup`. (#11679)
* Docs: Clarify `from-to-www` redirect direction. (#11682)
* added real-client-ip faq (#11663)
* Docs: Format NGINX configuration table. (#11659)
* Release controller v1.11.1/v1.10.3 & chart v4.11.1/v4.10.3. (#11654)
* Tests: Bump `test-runner` to v20240717-1fe74b5f. (#11645)
* Images: Trigger `test-runner` build. (#11636)
* Images: Bump `NGINX_BASE` to v0.0.10. (#11635)
* remove modsecurity coreruleset test files from nginx image (#11617)
* unskip the ocsp tests and update images to fix cfssl bug (#11606)
* Fix indent in YAML for example pod (#11598)
* Images: Bump `test-runner`. (#11600)
* Images: Bump `NGINX_BASE` to v0.0.9. (#11599)
* revert module upgrade (#11594)
* README: Fix support matrix. (#11586)
* Repository: Add changelogs from `release-v1.10`. (#11587)

### Dependency updates:

* Bump k8s.io/apiextensions-apiserver from 0.31.3 to 0.32.0 (#12566)
* Bump github.com/onsi/ginkgo/v2 from 2.22.0 to 2.22.1 (#12555)
* Bump k8s.io/code-generator from 0.31.3 to 0.32.0 (#12550)
* Bump k8s.io/cli-runtime from 0.31.3 to 0.32.0 (#12547)
* Bump k8s.io/apiserver from 0.31.3 to 0.32.0 (#12544)
* Bump the actions group with 2 updates (#12541)
* Bump google.golang.org/grpc from 1.68.1 to 1.69.2 (#12538)
* Bump k8s.io/client-go from 0.31.3 to 0.32.0 (#12512)
* Bump github.com/opencontainers/runc from 1.2.2 to 1.2.3 in the go group across 1 directory (#12509)
* Bump the actions group with 3 updates (#12506)
* Bump k8s.io/kube-aggregator from 0.31.3 to 0.32.0 in /images/kube-webhook-certgen/rootfs (#12505)
* Bump k8s.io/apimachinery from 0.31.3 to 0.32.0 in /images/ext-auth-example-authsvc/rootfs (#12502)
* Bump golang.org/x/crypto from 0.30.0 to 0.31.0 (#12476)
* Bump golang.org/x/crypto from 0.21.0 to 0.31.0 in /magefiles (#12472)
* Bump github.com/prometheus/common from 0.60.1 to 0.61.0 (#12464)
* Bump github/codeql-action from 3.27.5 to 3.27.6 in the actions group (#12461)
* Bump the go group across 1 directory with 2 updates (#12460)
* Bump github.com/onsi/ginkgo/v2 from 2.21.0 to 2.22.0 (#12426)
* Bump github.com/stretchr/testify from 1.9.0 to 1.10.0 (#12418)
* Bump the go group across 3 directories with 10 updates (#12413)
* Bump the actions group with 3 updates (#12412)
* Bump github.com/opencontainers/runc from 1.2.1 to 1.2.2 in the go group across 1 directory (#12380)
* Bump github/codeql-action from 3.27.1 to 3.27.4 in the actions group (#12373)
* Bump golangci-lint on actions and disable deprecated linters (#12361)
* Bump google.golang.org/grpc from 1.67.1 to 1.68.0 (#12357)
* Bump the actions group with 3 updates (#12354)
* Bump golang.org/x/crypto from 0.28.0 to 0.29.0 (#12349)
* Bump github.com/fsnotify/fsnotify from 1.7.0 to 1.8.0 (#12299)
* Bump github.com/opencontainers/runc from 1.2.0 to 1.2.1 in the go group across 1 directory (#12296)
* Bump github.com/onsi/ginkgo/v2 from 2.20.2 to 2.21.0 (#12288)
* Bump actions/dependency-review-action from 4.3.5 to 4.4.0 in the actions group (#12273)
* Bump the go group across 3 directories with 11 updates (#12244)
* Bump github.com/opencontainers/runc from 1.1.15 to 1.2.0 (#12242)
* Bump the actions group with 5 updates (#12236)
* Bump github.com/ncabatoff/process-exporter from 0.8.3 to 0.8.4 in the go group across 1 directory (#12218)
* Bump aquasecurity/trivy-action from 0.27.0 to 0.28.0 in the actions group (#12217)
* Bump github/codeql-action from 3.26.12 to 3.26.13 in the actions group (#12188)
* Bump the go group across 2 directories with 1 update (#12186)
* Bump the actions group with 2 updates (#12180)
* Bump github.com/opencontainers/runc from 1.1.14 to 1.1.15 in the go group across 1 directory (#12178)
* Bump the actions group with 3 updates (#12152)
* Bump golang.org/x/crypto from 0.27.0 to 0.28.0 (#12107)
* Bump the actions group with 3 updates (#12092)
* Bump sigs.k8s.io/mdtoc from 1.1.0 to 1.4.0 (#12062)
* Bump github.com/prometheus/common from 0.59.1 to 0.60.0 (#12060)
* Bump google.golang.org/grpc from 1.67.0 to 1.67.1 in the go group across 1 directory (#12059)
* Bump k8s.io/cli-runtime from 0.30.0 to 0.31.1 (#12061)
* Bump github/codeql-action from 3.26.9 to 3.26.10 in the actions group (#12051)
* Bump the go group across 1 directory with 3 updates (#12050)
* Bump k8s.io/kube-aggregator from 0.29.3 to 0.31.1 in /images/kube-webhook-certgen/rootfs (#12043)
* Bump k8s.io/apimachinery from 0.23.1 to 0.31.1 in /images/ext-auth-example-authsvc/rootfs (#12041)
* Bump github.com/prometheus/client_golang from 1.11.1 to 1.20.4 in /images/custom-error-pages/rootfs (#12040)
* Bump the all group with 2 updates (#12032)
* Bump github/codeql-action from 3.26.7 to 3.26.8 in the all group (#12010)
* Bump google.golang.org/grpc from 1.66.2 to 1.67.0 (#12009)
* Bump github.com/prometheus/client_golang from 1.20.3 to 1.20.4 in the all group (#12008)
* Bump the all group with 2 updates (#11977)
* Bump github/codeql-action from 3.26.6 to 3.26.7 in the all group (#11976)
* Bump github.com/prometheus/common from 0.57.0 to 0.59.1 (#11954)
* Bump golang.org/x/crypto from 0.26.0 to 0.27.0 (#11955)
* Bump github.com/prometheus/client_golang from 1.20.2 to 1.20.3 in the all group (#11953)
* Bump github.com/opencontainers/runc from 1.1.13 to 1.1.14 (#11928)
* Bump the all group with 2 updates (#11922)
* Bump github.com/onsi/ginkgo/v2 from 2.20.1 to 2.20.2 in the all group (#11901)
* Bump google.golang.org/grpc from 1.65.0 to 1.66.0 (#11902)
* Bump github.com/prometheus/common from 0.55.0 to 0.57.0 (#11903)
* Bump github/codeql-action from 3.26.5 to 3.26.6 in the all group (#11904)
* Bump the all group with 2 updates (#11865)
* Bump github/codeql-action from 3.26.2 to 3.26.5 in the all group (#11867)
* Bump github.com/prometheus/client_golang from 1.19.1 to 1.20.1 (#11832)
* Bump sigs.k8s.io/controller-runtime from 0.18.4 to 0.19.0 (#11823)
* Bump dario.cat/mergo from 1.0.0 to 1.0.1 in the all group (#11822)
* Bump k8s.io/component-base from 0.30.3 to 0.31.0 (#11825)
* Bump github/codeql-action from 3.26.0 to 3.26.2 in the all group (#11826)
* Bump github.com/onsi/ginkgo/v2 from 2.19.1 to 2.20.0 (#11766)
* Bump the all group with 2 updates (#11767)
* Bump golang.org/x/crypto from 0.25.0 to 0.26.0 (#11765)
* Bump the all group with 3 updates (#11727)
* Bump github.com/onsi/ginkgo/v2 from 2.19.0 to 2.19.1 in the all group (#11696)
* Bump the all group with 2 updates (#11695)
* Bump the all group with 4 updates (#11673)
* Bump the all group with 2 updates (#11672)
* Bump github.com/prometheus/common from 0.54.0 to 0.55.0 (#11522)
* Bump the all group with 5 updates (#11611)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.11.0...controller-v1.12.0
