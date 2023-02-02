# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.5.0

* add lint on chart before release (#9570)
* Values: Add missing `controller.metrics.service.labels`. (#9501)
* HPA: Add `controller.autoscaling.annotations` to `values.yaml`. (#9253)
* ci: remove setup-helm step (#9404)
* feat(helm): Optionally use cert-manager instead admission patch (#9279)
* run helm release on main only and when the chart/value changes only (#9290)
* Update Ingress-Nginx version controller-v1.6.2

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.4.2...helm-chart-4.5.0
