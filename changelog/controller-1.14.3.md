# Changelog

### controller-v1.14.3

Images:

* registry.k8s.io/ingress-nginx/controller:v1.14.3@sha256:82917be97c0939f6ada1717bb39aa7e66c229d6cfb10dcfc8f1bd42f9efe0f81
* registry.k8s.io/ingress-nginx/controller-chroot:v1.14.3@sha256:ffdab64d0e0556f810d82d618a0fa97c4fc8dc2bc5717c51bfe83b5d4252c73e

### All changes:

* Images: Trigger controller build. (#14508)
* Annotations: Add `^` and `$` to auth method regex. (#14505)
* Template: Quote all `location` and `server_name` directives, and escape quotes and backslashes. (#14502)
* Controller: Verify UIDs. (#14499)
* Template: Bypass custom error pages when handling auth URL requests. (#14496)
* Admission Controller: Use 9 MB limit. (#14493)
* Images: Bump other images. (#14484)
* Images: Trigger other builds (2/2). (#14480)
* Images: Trigger other builds (1/2). (#14477)
* Tests: Bump Test Runner to v2.2.7. (#14471)
* Images: Trigger Test Runner build. (#14468)
* Images: Bump NGINX to v2.2.7. (#14465)
* Images: Trigger NGINX build. (#14462)
* Go: Update dependencies. (#14459)
* Images: Bump GCB Docker GCloud to v20260127-c1affcc8de. (#14456)
* CI: Update Helm to v4.1.0. (#14453)
* Controller: Fix sync for when host clock jumps to future. (#14451)
* Util: Fix panic for empty `cpu.max` file. (#14448)
* NGINX: Update OWASP CRS to v4.22.0. (#14417)

### Dependency updates:

* Bump the actions group with 2 updates (#14490)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.14.2...controller-v1.14.3
