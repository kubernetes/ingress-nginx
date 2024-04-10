# Changelog

This release is the first using NGINX v1.25.0!

## Breaking changes
* This version does not support chroot image, this will be fixed on a future minor patch release
* This version dropped Opentracing and zipkin modules, just Opentelemetry is supported
* This version dropped support for PodSecurityPolicy
* This version dropped support for GeoIP (legacy). Only GeoIP2 is supported

### controller-v1.10.0

Images:

* registry.k8s.io/ingress-nginx/controller:v1.10.0@sha256:42b3f0e5d0846876b1791cd3afeb5f1cbbe4259d6f35651dcc1b5c980925379c

### All changes:

* Start the release of v1.10.0 (#11038)
* bump nginx and Go, remove tag file and old CI jobs (#11037)
* Fix kubewebhook image tag (#11033)
* add missing backend-protocol annotation option (#9545)
* Update controller-prometheusrules.yaml (#8902)
* Stop reporting interrupted tests (#11027)
* test(gzip): reach ingress (#9541)
* fix datasource, $exported_namespace variable in grafana nginx dashboard (#9092)
* Properly support a TLS-wrapped OCSP responder (#10164)
* Fix print-e2e-suite (#9536)
* chore(deps): upgrade headers-more module to 0.37 (#10991)
* Update ingress-path-matching.md (#11008)
* Update ingress-path-matching.md (#11007)
* E2E Tests: Explicitly enable metrics. (#10962)
* Chart: Set `--enable-metrics` depending on `controller.metrics.enabled`. (#10959)
* Chart: Remove useless `default` from `_params.tpl`. (#10957)
* Fix golang makefile var name (#10932)
* Fixing image push (#10931)
* fix: live-docs script (#10928)
* docs: Add vouch-proxy OAuth example (#10929)
* Add OTEL build test and for NGINX v1.25 (#10889)
* docs: update annotations docs with missing session-cookie section (#10917)
* Release controller 1.9.6 and helm 4.9.1 (#10919)

### Dependency updates:

* Bump kubewebhook certgen (#11034)
* Bump go libraries (#11023)
* Bump modsecurity on nginx 1.25 (#11024)
* Bump grpc and reintroduce OTEL compilation (#11021)
* Bump github/codeql-action from 3.24.0 to 3.24.5 (#11017)
* Bump actions/dependency-review-action from 4.0.0 to 4.1.3 (#11016)
* Bump dorny/paths-filter from 3.0.0 to 3.0.1 (#10994)
* Bump github.com/prometheus/client_model from 0.5.0 to 0.6.0 (#10998)
* Bump actions/upload-artifact from 4.3.0 to 4.3.1 (#10978)
* Bump actions/download-artifact from 4.1.1 to 4.1.2 (#10981)
* Bump aquasecurity/trivy-action from 0.16.1 to 0.17.0 (#10979)
* Bump golangci/golangci-lint-action from 3.7.0 to 4.0.0 (#10980)
* Bump golang.org/x/crypto from 0.18.0 to 0.19.0 (#10976)
* Bump github/codeql-action from 3.23.2 to 3.24.0 (#10971)
* Bump github.com/opencontainers/runc from 1.1.11 to 1.1.12 (#10951)
* Bump google.golang.org/grpc from 1.60.1 to 1.61.0 (#10938)
* Bump actions/upload-artifact from 4.2.0 to 4.3.0 (#10937)
* Bump dorny/test-reporter from 1.7.0 to 1.8.0 (#10936)
* Bump github/codeql-action from 3.23.1 to 3.23.2 (#10935)
* Bump dorny/paths-filter from 2.11.1 to 3.0.0 (#10934)
* Bump alpine to 3.19.1 (#10930)
* Bump go to v1.21.6 and set a single source of truth (#10926)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.9.6...controller-v1.10.0
