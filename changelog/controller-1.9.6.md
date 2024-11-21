# Changelog

### controller-v1.9.6

Images:

* registry.k8s.io/ingress-nginx/controller:v1.9.6@sha256:1405cc613bd95b2c6edd8b2a152510ae91c7e62aea4698500d23b2145960ab9c
* registry.k8s.io/ingress-nginx/controller-chroot:v1.9.6@sha256:7eb46ff733429e0e46892903c7394aff149ac6d284d92b3946f3baf7ff26a096

### All changes:

* update web hook cert gen to latest release v20231226-1a7112e06
* annotation validation: validate regex in common name annotation (#10880)
* change MODSECURITY_VERSION_LIB to 3.0.11 (#10879)
* Include SECLEVEL and STRENGTH as part of ssl-cipher list validation (#10871)

### Dependency updates:

* Bump github.com/opencontainers/runc from 1.1.10 to 1.1.11 (#10878)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.9.5...controller-v1.9.6
