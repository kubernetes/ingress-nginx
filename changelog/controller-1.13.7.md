# Changelog

### controller-v1.13.7

Images:

* registry.k8s.io/ingress-nginx/controller:v1.13.7@sha256:13db2f8aca4bb71ae7f727288620c4569b01bab4911b01fa3917995ff7755de8
* registry.k8s.io/ingress-nginx/controller-chroot:v1.13.7@sha256:ecc934a4653d3b8c17100882b58cc16c59a697930c3598acda02227edfc41c34

### All changes:

* Images: Trigger controller build. (#14509)
* Annotations: Add `^` and `$` to auth method regex. (#14506)
* Template: Quote all `location` and `server_name` directives, and escape quotes and backslashes. (#14503)
* Controller: Verify UIDs. (#14500)
* Template: Bypass custom error pages when handling auth URL requests. (#14497)
* Admission Controller: Use 9 MB limit. (#14494)
* Images: Bump other images. (#14485)
* Images: Trigger other builds (2/2). (#14481)
* Images: Trigger other builds (1/2). (#14478)
* Tests: Bump Test Runner to v2.2.7. (#14472)
* Images: Trigger Test Runner build. (#14469)
* Images: Bump NGINX to v2.2.7. (#14466)
* Images: Trigger NGINX build. (#14463)
* Go: Update dependencies. (#14460)
* Images: Bump GCB Docker GCloud to v20260127-c1affcc8de. (#14457)
* CI: Update Helm to v4.1.0. (#14454)
* Controller: Fix sync for when host clock jumps to future. (#14450)
* Util: Fix panic for empty `cpu.max` file. (#14449)
* NGINX: Update OWASP CRS to v4.22.0. (#14418)

### Dependency updates:

* Bump the actions group with 2 updates (#14491)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.13.6...controller-v1.13.7
