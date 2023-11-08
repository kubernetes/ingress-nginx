# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.1.0

* [8481](https://github.com/kubernetes/ingress-nginx/pull/8481) Fix log creation in chroot script
* [8479](https://github.com/kubernetes/ingress-nginx/pull/8479) changed nginx base img tag to img built with alpine3.14.6
* [8478](https://github.com/kubernetes/ingress-nginx/pull/8478) update base images and protobuf gomod
* [8468](https://github.com/kubernetes/ingress-nginx/pull/8468) Fallback to ngx.var.scheme for redirectScheme with use-forward-headers when X-Forwarded-Proto is empty
* [8456](https://github.com/kubernetes/ingress-nginx/pull/8456) Implement object deep inspector
* [8455](https://github.com/kubernetes/ingress-nginx/pull/8455) Update dependencies
* [8454](https://github.com/kubernetes/ingress-nginx/pull/8454) Update index.md
* [8447](https://github.com/kubernetes/ingress-nginx/pull/8447) typo fixing
* [8446](https://github.com/kubernetes/ingress-nginx/pull/8446) Fix suggested annotation-value-word-blocklist
* [8444](https://github.com/kubernetes/ingress-nginx/pull/8444) replace deprecated topology key in example with current one
* [8443](https://github.com/kubernetes/ingress-nginx/pull/8443) Add dependency review enforcement
* [8434](https://github.com/kubernetes/ingress-nginx/pull/8434) added new auth-tls-match-cn annotation
* [8426](https://github.com/kubernetes/ingress-nginx/pull/8426) Bump github.com/prometheus/common from 0.32.1 to 0.33.0

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.0.18...helm-chart-4.1.0
