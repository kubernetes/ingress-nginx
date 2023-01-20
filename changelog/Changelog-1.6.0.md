# Changelog

### 1.6.0

* Upgrade golang to 1.19.4
* Upgrade to Alpine 3.17
* A breaking change on how ingress deals with pathType and paths on ingress objects. Implement pathType validation (#9511)


This release adds a breaking change on how ingress deals with pathType and paths on ingress objects.

Previously, one could add an ingress like:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: grafana
  namespace: default
  annotations:
    nginx.kubernetes.io/use-regex: true
spec:
  rules:
    - host: grafana.bla.com
      http:
        paths:
          - backend:
              service:
                name: grafana
                port:
                  number: 80
            path: /grafana[0-9]{6}
            pathType: Prefix
 ```

But this path is not conformant with RFC or even Kubernetes specifications that states that path field should only 
contain a valid path, and specific cases should have a [pathType: ImplementationSpecific](https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types)

Now, to use any regex on path, **EVEN IF REGEX ANNOTATION OR REWRITE ANNOTATION ARE PRESENT** the pathType should be of type ImplementationSpecific.

- Accepted characters now are A-Za-z0-9-._~/ 
- The validation above is more wide if pathType is ImplementationSpecific, allowing also common regex characters like ^%$[](){}*+? 
- The set of characters is configurable in ConfigMap option `path-additional-allowed-chars` 
- A user can disable this behavior setting the ConfigMap entry `disable-pathtype-validation` to true

Disabling pathType validation will disable the uncommon characters on pathType Exact and Prefix, 
but the additional Allowed Characters will still be respected. If user wants to add more characters they should change the default

Images:

 * registry.k8s.io/controller:controller-v1.6.0@sha256:75d3b12eeb889aba3ae5c9505da5625fb9000a5bdec4d7dabe4c94d2d42fdcc4
 * registry.k8s.io/controller-chroot:controller-v1.6.0@sha256:d8d76a9f28af2675f555c83b97b8429216a440809fd9c57313e02d6846949c9f
 
### All Changes:

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
 
**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-controller-v1.5.2...controller-controller-v1.6.0
