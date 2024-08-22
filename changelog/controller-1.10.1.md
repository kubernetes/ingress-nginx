# Changelog

### controller-v1.10.1

Images:

* registry.k8s.io/ingress-nginx/controller:v1.10.1@sha256:e24f39d3eed6bcc239a56f20098878845f62baa34b9f2be2fd2c38ce9fb0f29e
* registry.k8s.io/ingress-nginx/controller-chroot:v1.10.1@sha256:c155954116b397163c88afcb3252462771bd7867017e8a17623e83601bab7ac7

### All changes:

* start 1.10.1 build (#11246)
* force nginx rebuild (#11245)
* update k8s version to latest kind release (#11241)
* remove _ssl_expire_time_seconds metric by identifier (#11239)
* update post submit helm ci and clean up (#11221)
* Chart: Add unit tests for default backend & topology spread constraints. (#11219)
* sort default backend hpa metrics (#11217)
* updated certgen image shatag (#11216)
* changed testrunner image sha (#11211)
* bumped certgeimage tag (#11213)
* updated baseimage & deleted a useless file (#11209)
* bump ginkgo to 2-17-1 in testrunner (#11204)
* chunking related faq update (#11205)
* Fix-semver (#11199)
* refactor helm ci tests part I (#11188)
* Proposal: e2e tests for regex patterns (#11185)
* bump ginkgo to v2.17.1 (#11186)
* fixes brotli build issue (#11187)
* fix geoip2 configuration docs (#11151)
* Fix typos in OTel doc (#11081) (#11129)
* Chart: Render `controller.ingressClassResource.parameters` natively. (#11126)
* Fix admission controller logging of `admissionTime` and `testedConfigurationSize` (#11114)
* Chart: Align HPA & KEDA conditions. (#11113)
* Chart: Improve IngressClass documentation. (#11111)
* Chart: Add Gacko to maintainers. Again. (#11112)
* Chart: Deploy `PodDisruptionBudget` with KEDA. (#11105)
* Chores: Pick patches from main. (#11103)

### Dependency updates:

* Bump google.golang.org/grpc from 1.63.0 to 1.63.2 (#11238)
* Bump google.golang.org/grpc from 1.62.1 to 1.63.0 (#11234)
* Bump github.com/prometheus/common from 0.51.1 to 0.52.2 (#11233)
* Bump golang.org/x/crypto from 0.21.0 to 0.22.0 (#11232)
* Bump github.com/prometheus/client_model in the all group (#11231)
* Bump the all group with 3 updates (#11230)
* Bump the all group with 2 updates (#11190)
* Bump actions/add-to-project from 0.6.1 to 1.0.0 (#11189)
* Bump the all group with 3 updates (#11166)
* Bump github.com/prometheus/common from 0.50.0 to 0.51.1 (#11160)
* Bump the all group with 4 updates (#11140)
* Bump the all group with 1 update (#11136)
* Bump google.golang.org/protobuf from 1.32.0 to 1.33.0 in /magefiles (#11127)
* Bump google.golang.org/protobuf in /images/custom-error-pages/rootfs (#11128)
* Bump google.golang.org/protobuf in /images/kube-webhook-certgen/rootfs (#11122)

