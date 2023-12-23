# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.3.0

* Support for Kubernetes v.1.25.0 was added and support for endpoint slices
* Support for Kubernetes v1.20.0 and v1.21.0 was removed
* [8890](https://github.com/kubernetes/ingress-nginx/pull/8890) migrate to endpointslices
* [9059](https://github.com/kubernetes/ingress-nginx/pull/9059) kubewebhookcertgen sha change after go1191
* [9046](https://github.com/kubernetes/ingress-nginx/pull/9046) Parameterize metrics port name
* [9104](https://github.com/kubernetes/ingress-nginx/pull/9104) Fix yaml formatting error with multiple annotations

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.2.1...helm-chart-4.3.0
