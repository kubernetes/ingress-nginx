# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.4.0

* Adding support for disabling liveness and readiness probes to the Helm chart by @njegosrailic in https://github.com/kubernetes/ingress-nginx/pull/9238
* add:(admission-webhooks) ability to set securityContext by @ybelMekk in https://github.com/kubernetes/ingress-nginx/pull/9186
* #7652 - Updated Helm chart to use the fullname for the electionID if not specified. by @FutureMatt in https://github.com/kubernetes/ingress-nginx/pull/9133
* Rename controller-wehbooks-networkpolicy.yaml. by @Gacko in https://github.com/kubernetes/ingress-nginx/pull/9123

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.3.0...helm-chart-4.4.0
