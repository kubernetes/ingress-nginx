# Changelog

### controller-v1.9.5

Images:

* registry.k8s.io/ingress-nginx/controller:v1.9.5@sha256:b3aba22b1da80e7acfc52b115cae1d4c687172cbf2b742d5b502419c25ff340e
* registry.k8s.io/ingress-nginx/controller-chroot:v1.9.5@sha256:9a8d7b25a846a6461cd044b9aea9cf6cad972bcf2e64d9fd246c0279979aad2d

### All changes:

* update nginx build (#10781)
* update images from golang upgrade (#10762)
* fix: remove tcpproxy copy error handling (#10715)
* Ignore fake certificate for NGINXCertificateExpiry (#10694)
* Comment NGINXCertificateExpiry alert label matcher (#10692)
* chart: allow setting allocateLoadBalancerNodePorts (#10693)
* [release-1.9] feat(helm): add documentation about metric args (#10695)
* chore(dep): change lua-resty-cookie's repo (#10691)
* annotation validation - extended URLWithNginxVariableRegex from alphaNumericChars to extendedAlphaNumeric (#10656)
* fix: adjust unfulfillable validation check for session-cookie-samesite annotation (#10604)
* fix: Validate x-forwarded-prefix annotation with RegexPathWithCapture (#10603)
* Increase HSTS max-age to default to one year (#10580)
* [release-1.9] update nginx base, httpbun, e2e, helm webhook cert gen (#10507)
* [release-1.9] add upstream patch for CVE-2023-44487 (#10499)
* fix brotli build issues (#10468)
* upgrade owasp modsecurity core rule set to v3.3.5 (#10437)
* Accept backend protocol on any case (#10461)
* Chart: Rework network policies. (#10438)
* Rework mage (#10418)

### Dependency updates:

* Bump x/net (#10517)
* Bump google.golang.org/grpc from 1.58.0 to 1.58.1 (#10436)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.9.4...controller-v1.9.5
