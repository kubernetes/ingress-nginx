# Changelog

### controller-v1.10.2

Images:

* registry.k8s.io/ingress-nginx/controller:v1.10.2@sha256:e3311b3d9671bc52d90572bcbfb7ee5b71c985d6d6cffd445c241f1e2703363c
* registry.k8s.io/ingress-nginx/controller-chroot:v1.10.2@sha256:c4395cba98f9721e3381d3c06e7994371bae20f5ab30e457cd7debe44a8c8c54

### All changes:

* update test runner to latest build (#11557)
* add k8s 1.30 to ci build (#11553)
* update test runner go base to 3.20 (#11550)
* tag new test runner image with new nginx base 0.0.8 (#11549)
* bump NGINX_BASE to v0.0.8 (#11543)
* trigger build for NGINX-1.25 v0.0.8 (#11542)
* Upgrade OWASP_MODSECURITY_CRS_VERSION 3.3.5 to 4.4.0 and update docs (#11548)
* [feature] bump nginx to 1.25.5 and add http3 module (#11541)
* add ssl patches to nginx-1.25 image for coroutines to work in lua client hello and cert ssl blocks (#11534)
* bump alpine version to 3.20 to custom-error-pages (#11537)
* fix: Ensure changes in MatchCN annotation are detected (#11528)
* Docs: Add information about HTTP/3 support. (#11525)
* Docs: Specify `ingressClass` for multi-controller setup. (#11520)
* Docs: Improve default certificate usage. (#11519)
* docs: Update Ingress-NGINX v1.10.1 compatibility with Kubernetes v1.30 (#11500)
* Update getting-started.md with new prerequisites (#11487)
* Fix boolean configuration (#11484)
* Chores: Align security contacts & chart maintainers to actual owners. (#11480)
* CI: Bump forgotten Ginkgo versions. (#11469)
* Tests: Replace deprecated `grpc.Dial` by `grpc.NewClient`. (#11468)
* Owners: Promote Gacko to admin. (#11464)
* fixed fastcgi userguide (#11455)
* Remove unnecessary space character (#11451)
* fix for docs issue 11432 (#11446)
* Update index.md (#11445)
* upgrade to alpine 3.20 (#11438)
* update golang to 1.22.4 (#11431)
* Adapt dashboards for Grafana 11 compatibility (#11414)
* Rename variable to fix typo (#11413)
* Fix helm install on cloud provider admonition block (#11412)
* edited helm-install tips (#11411)
* added info for aws helm install (#11410)
* added multiplecontrollers-howto to faq (#11409)
* removed tlsv1 & tlsv1.1 (#11408)
* Docs: Remove opentracing and zipkin from docs (#11405)
* Go: Sync modules from `main`. (#11398)
* add workflow to helm release and update ct for branch (#11317)
* Merge pull request #11277 from strongjz/chart-1.10.1 (#11314)
* Release Helm Chart on branch update (#11306)
* Release controller 1.10.1 (#11298)
* fix path in file changed detected message (#11286)
* chore: fix function names in comment (#11281)
* fix: update kube version requirement to 1.21 (#11279)
* release helm chart from release branch (#11278)
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
* Start the release of v1.10.0 (#11038)

### Dependency updates:

* Bump the all group with 2 updates (#11524)
* Bump k8s.io/klog/v2 from 2.130.0 to 2.130.1 in the all group (#11521)
* Bump aquasecurity/trivy-action from 0.22.0 to 0.23.0 in the all group (#11501)
* Bump k8s.io/klog/v2 from 2.120.1 to 2.130.0 (#11479)
* Bump the all group with 3 updates (#11478)
* Bump the all group with 2 updates (#11477)
* Bump golang.org/x/crypto from 0.23.0 to 0.24.0 (#11471)
* Bump sigs.k8s.io/controller-runtime in the all group (#11449)
* Bump github.com/prometheus/common from 0.53.0 to 0.54.0 (#11447)
* Bump the all group with 3 updates (#11450)
* Bump goreleaser/goreleaser-action from 5.1.0 to 6.0.0 (#11448)
* Bump github.com/onsi/ginkgo/v2 from 2.17.2 to 2.19.0 (#11422)
* Bump the all group with 2 updates (#11421)
* Bump google.golang.org/grpc from 1.63.2 to 1.64.0 (#11423)
* Bump the all group across 1 directory with 6 updates (#11407)
* Bump golangci/golangci-lint-action from 5.3.0 to 6.0.1 (#11406)
* Bump the all group with 3 updates (#11404)
* Bump Kubernetes version on images (#11403)
* Bump golangci/golangci-lint-action from 4.0.0 to 5.0.0 (#11402)
* Bump the all group with 4 updates (#11380)
* Bump k8s.io/component-base from 0.29.3 to 0.30.0 (#11301)
* Bump github.com/prometheus/common from 0.52.3 to 0.53.0 (#11300)
* Bump golang.org/x/net from 0.22.0 to 0.23.0 (#11285)
* Bump golang.org/x/net in /images/kube-webhook-certgen/rootfs (#11284)
* Bump the all group with 2 updates (#11266)
* Bump azure/setup-helm from 3.5 to 4 (#11265)
* Bump actions/add-to-project from 1.0.0 to 1.0.1 in the all group (#11264)
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

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-v1.10.1...controller-v1.10.2
