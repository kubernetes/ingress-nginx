# Changelog

### 1.6.1
Images:

 * registry.k8s.io/controller:controller-v1.6.1@sha256:9f873e9c8cb154014dad73681395f14ce6d477d83774283eeadea2d3c484091a
 * registry.k8s.io/controller-chroot:controller-v1.6.1@sha256:17bbbcd716b36e7eeb2d03996d9af8a3989ced7da53a0dc56be4f9cf1f36c579
 
### All Changes:

* Switch logic on path type validation and setting it to false (#9543)
* tcpproxy: increase buffer size to 16K (#9548)
* Move and spell-check Kubernetes 1.22 migration FAQ (#9544)
* Add CORS template check inside location for externalAuth.SignURL (#8814)
* fix(grafana-dashboard): remove hardcoded namespace references (#9523)
* Align default value for keepalive_request with NGINX default (#9518)
* Implement pathType validation (#9511)
* feat(configmap): expose gzip-disable (#9505)
* Values: Add missing `controller.metrics.service.labels`. (#9501)
* Add docs about orphan_ingress metric (#9514)
* Add new prometheus metric for orphaned ingress (#8230)
* Sanitise request metrics in monitoring docs (#9384)
* Change default value of enable-brotli (#9500)
* feat: support topology aware hints (#9165)
* Remove 1.5.2 from readme (#9498)
* Remove nonexistent load flag from docker build commands (#9122)
* added option to disable sync event creation (#8528)
* Add buildResolvers to the stream module (#9184)
* fix: disable auth access logs (#9049)
* Adding ipdenylist annotation (#8795)
* Add update updateStrategy and minReadySeconds for defaultBackend (#8506)
* Fix indentation on serviceAccount annotation (#9129)
* Update monitoring.md (#9269)
* add github actions stale bot (#9439)
* Admission Webhooks/Job: Add `NetworkPolicy`. (#9218)
* update OpenTelemetry image (#9491)
* bump OpenTelemetry (#9489)
* Optional podman support (#9294)
* fix change images (#9463)
* move tests to gh actions (#9461)
* Automated Release Controller 1.5.2 (#9455)
* Add sslpassthrough tests (#9457)
* updated the link in RELEASE.md file (#9456)
* restart 1.5.2 release process (#9450)
* Update command line arguments documentation (#9224)
* start release 1.5.2 (#9445)
* upgrade nginx base image (#9436)
* test the new e2e test images (#9444)
* avoid builds and tests for non-code changes (#9392)
* CI updates (#9440)
* HPA: Add `controller.autoscaling.annotations` to `values.yaml`. (#9253)
* update the nginx run container for alpine:3.17.0 (#9430)
* cleanup: remove ioutil for new go version (#9427)
* start upgrade to golang 1.19.4 and alpine 3.17.0 (#9417)

### Dependencies updates: 
* Bump google.golang.org/grpc from 1.52.0 to 1.52.3 (#9555)
* Bump k8s.io/klog/v2 from 2.80.1 to 2.90.0 (#9553)
* Bump sigs.k8s.io/controller-runtime from 0.13.1 to 0.14.2 (#9552)
* Bump google.golang.org/grpc from 1.51.0 to 1.52.0 (#9512)
* Bump `client-go` to remove dependence on go-autorest dependency (#9488)
* Bump golang.org/x/crypto from 0.4.0 to 0.5.0 (#9494)
* Bump golang.org/x/crypto from 0.3.0 to 0.4.0 (#9397)
* Bump github.com/onsi/ginkgo/v2 from 2.6.0 to 2.6.1 (#9432)
* Bump github.com/onsi/ginkgo/v2 from 2.6.0 to 2.6.1 (#9421)
* Bump github/codeql-action from 2.1.36 to 2.1.37 (#9423)
* Bump actions/checkout from 3.1.0 to 3.2.0 (#9425)
* Bump goreleaser/goreleaser-action from 3.2.0 to 4.1.0 (#9426)
* Bump actions/dependency-review-action from 3.0.1 to 3.0.2 (#9424)
* Bump ossf/scorecard-action from 2.0.6 to 2.1.0 (#9422)
 
**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-controller-v1.5.2...controller-controller-v1.6.1
