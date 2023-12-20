# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.8.0-beta.0

* ci(helm): fix Helm Chart release action 422 error (#10237)
* helm: Use .Release.Namespace as default for ServiceMonitor namespace (#10249)
* [helm] configure allow to configure hostAliases (#10180)
* [helm] pass service annotations through helm tpl engine (#10084)
* Update Ingress-Nginx version controller-v1.9.0-beta.0

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.7.2...helm-chart-4.8.0-beta.0
