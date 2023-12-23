# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.2.0

* Support for Kubernetes v1.19.0 was removed
* [8810](https://github.com/kubernetes/ingress-nginx/pull/8810) Prepare for v1.3.0
* [8808](https://github.com/kubernetes/ingress-nginx/pull/8808) revert arch var name
* [8805](https://github.com/kubernetes/ingress-nginx/pull/8805) Bump k8s.io/klog/v2 from 2.60.1 to 2.70.1
* [8803](https://github.com/kubernetes/ingress-nginx/pull/8803) Update to nginx base with alpine v3.16
* [8802](https://github.com/kubernetes/ingress-nginx/pull/8802) chore: start v1.3.0 release process
* [8798](https://github.com/kubernetes/ingress-nginx/pull/8798) Add v1.24.0 to test matrix
* [8796](https://github.com/kubernetes/ingress-nginx/pull/8796) fix: add MAC_OS variable for static-check
* [8793](https://github.com/kubernetes/ingress-nginx/pull/8793) changed to alpine-v3.16
* [8781](https://github.com/kubernetes/ingress-nginx/pull/8781) Bump github.com/stretchr/testify from 1.7.5 to 1.8.0
* [8778](https://github.com/kubernetes/ingress-nginx/pull/8778) chore: remove stable.txt from release process
* [8775](https://github.com/kubernetes/ingress-nginx/pull/8775) Remove stable
* [8773](https://github.com/kubernetes/ingress-nginx/pull/8773) Bump github/codeql-action from 2.1.14 to 2.1.15
* [8772](https://github.com/kubernetes/ingress-nginx/pull/8772) Bump ossf/scorecard-action from 1.1.1 to 1.1.2
* [8771](https://github.com/kubernetes/ingress-nginx/pull/8771) fix bullet md format
* [8770](https://github.com/kubernetes/ingress-nginx/pull/8770) Add condition for monitoring.coreos.com/v1 API
* [8769](https://github.com/kubernetes/ingress-nginx/pull/8769) Fix typos and add links to developer guide
* [8767](https://github.com/kubernetes/ingress-nginx/pull/8767) change v1.2.0 to v1.2.1 in deploy doc URLs
* [8765](https://github.com/kubernetes/ingress-nginx/pull/8765) Bump github/codeql-action from 1.0.26 to 2.1.14
* [8752](https://github.com/kubernetes/ingress-nginx/pull/8752) Bump github.com/spf13/cobra from 1.4.0 to 1.5.0
* [8751](https://github.com/kubernetes/ingress-nginx/pull/8751) Bump github.com/stretchr/testify from 1.7.2 to 1.7.5
* [8750](https://github.com/kubernetes/ingress-nginx/pull/8750) added announcement
* [8740](https://github.com/kubernetes/ingress-nginx/pull/8740) change sha e2etestrunner and echoserver
* [8738](https://github.com/kubernetes/ingress-nginx/pull/8738) Update docs to make it easier for noobs to follow step by step
* [8737](https://github.com/kubernetes/ingress-nginx/pull/8737) updated baseimage sha
* [8736](https://github.com/kubernetes/ingress-nginx/pull/8736) set ld-musl-path
* [8733](https://github.com/kubernetes/ingress-nginx/pull/8733) feat: migrate leaderelection lock to leases
* [8726](https://github.com/kubernetes/ingress-nginx/pull/8726) prometheus metric: upstream_latency_seconds
* [8720](https://github.com/kubernetes/ingress-nginx/pull/8720) Ci pin deps
* [8719](https://github.com/kubernetes/ingress-nginx/pull/8719) Working OpenTelemetry sidecar (base nginx image)
* [8714](https://github.com/kubernetes/ingress-nginx/pull/8714) Create Openssf scorecard
* [8708](https://github.com/kubernetes/ingress-nginx/pull/8708) Bump github.com/prometheus/common from 0.34.0 to 0.35.0
* [8703](https://github.com/kubernetes/ingress-nginx/pull/8703) Bump actions/dependency-review-action from 1 to 2
* [8701](https://github.com/kubernetes/ingress-nginx/pull/8701) Fix several typos
* [8699](https://github.com/kubernetes/ingress-nginx/pull/8699) fix the gosec test and a make target for it
* [8698](https://github.com/kubernetes/ingress-nginx/pull/8698) Bump actions/upload-artifact from 2.3.1 to 3.1.0
* [8697](https://github.com/kubernetes/ingress-nginx/pull/8697) Bump actions/setup-go from 2.2.0 to 3.2.0
* [8695](https://github.com/kubernetes/ingress-nginx/pull/8695) Bump actions/download-artifact from 2 to 3
* [8694](https://github.com/kubernetes/ingress-nginx/pull/8694) Bump crazy-max/ghaction-docker-buildx from 1.6.2 to 3.3.1

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.1.2...helm-chart-4.2.0
