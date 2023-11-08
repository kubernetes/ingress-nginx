# Changelog

### 1.8.1
Images:

 * registry.k8s.io/ingress-nginx/controller:v1.8.1@sha256:e5c4824e7375fcf2a393e1c03c293b69759af37a9ca6abdb91b13d78a93da8bd
 * registry.k8s.io/ingress-nginx/controller-chroot:v1.8.1@sha256:e0d4121e3c5e39de9122e55e331a32d5ebf8d4d257227cb93ab54a1b912a7627
 
### All Changes:

* netlify: Only trigger preview when there are changes in docs. (#10144)
* changed to updated baseimage and reverted tag (#10143)
* Fix loadBalancerClass value (#10139)
* Added a doc line to the missing helm value service.internal.loadBalancerIP (#9406)
* Set grpc :authority header from request header (#8912)
* bump pinned golang to 1.20.5 (#10127)
* update test runner (#10125)
* chore: remove echo from snippet tests (#10110)
* Update typo in docs for lb scheme (#10117)
* golang 1.20.5 bump (#10120)
* feat(helm): Add loadBalancerClass (#9562)
* chore: remove echo friom canary tests (#10089)
* fix: obsolete warnings (#10029)
* docs: change Dockefile url ref main (#10087)
* Revert "Remove fastcgi feature" (#10081)
* docs: add netlify configuration (#10073)
* add distroless otel init (#10035)
* chore: move httpbun to be part of framework (#9955)
* Remove fastcgi feature (#9864)
* Fix mirror-target values without path separator and port (#9889)
* Adding feature to upgrade Oracle Cloud Infrastructure's Flexible Load Balancer and adjusting Health Check that were critical in the previous configuration (#9961)
* add support for keda fallback settings (#9993)
* unnecessary use of fmt.Sprint (S1039) (#10049)
* chore: pkg imported more than once (#10048)
* tracing: upgrade to dd-opentracing-cpp v1.3.7 (#10031)
* fix: add canary to sidebar in examples (#10068)
* docs: add lua testing documentation (#10060)
* docs: canary weighted deployments example (#10067)
* Update Internal Load Balancer docs (#10062)
* fix broken kubernetes.io/user-guide/ docs links (#10055)
* docs: Updated the content of deploy/rbac.md (#10054)
* ensured hpa mem spec before cpu spec (#10043)
* Fix typo in controller_test (#10034)
* chore(dep): upgrade github.com/emicklei/go-restful/v3 to 3.10 (#10028)
* Upgrade to Golang 1.20.4 (#10016)
* perf: avoid unnecessary byte/string conversion (#10012)
* added note on dns for localtesting (#10021)
* added helmshowvalues example (#10019)
* release controller 1.8.0 and chart 4.7.0 (#10017)

### Dependencies updates: 
* Bump ossf/scorecard-action from 2.1.3 to 2.2.0 (#10133)
* Bump google.golang.org/grpc from 1.56.0 to 1.56.1 (#10134)
* Bump github.com/prometheus/client_golang from 1.15.1 to 1.16.0 (#10106)
* Bump golang.org/x/crypto from 0.9.0 to 0.10.0 (#10105)
* Bump google.golang.org/grpc from 1.55.0 to 1.56.0 (#10103)
* Bump goreleaser/goreleaser-action from 4.2.0 to 4.3.0 (#10101)
* Bump docker/setup-buildx-action from 2.6.0 to 2.7.0 (#10102)
* Bump actions/checkout from 3.5.2 to 3.5.3 (#10076)
* Bump docker/setup-qemu-action from 2.1.0 to 2.2.0 (#10075)
* Bump aquasecurity/trivy-action from 0.10.0 to 0.11.2 (#10078)
* Bump docker/setup-buildx-action from 2.5.0 to 2.6.0 (#10077)
* Bump actions/dependency-review-action from 3.0.4 to 3.0.6 (#10042)
* Bump github.com/stretchr/testify from 1.8.3 to 1.8.4 (#10041)
* Bump github.com/stretchr/testify from 1.8.2 to 1.8.3 (#10005)
 
**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-controller-v1.8.0...controller-controller-v1.8.1
