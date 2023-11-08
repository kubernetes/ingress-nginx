# Changelog

### 1.9.3
Images:

 * registry.k8s.io/ingress-nginx/controller:v1.9.3@sha256:8fd21d59428507671ce0fb47f818b1d859c92d2ad07bb7c947268d433030ba98
 * registry.k8s.io/ingress-nginx/controller-chroot:v1.9.3@sha256:df4931fd6859fbf1a71e785f02a44b2f9a16f010ae852c442e9bb779cbefdc86
 
### All Changes:

* update nginx base, httpbun, e2e, helm webhook cert gen (#10506)
* added warning for configuration-snippets usage (#10492)
* Remove legacy GeoIP from controller (#10495)
* add upstream patch for CVE-2023-44487 (#10494)
* Revert "Remove curl from nginx base image (#10477)" (#10479)
* update error and otel to have all the arch we support (#10476)
* Remove curl from nginx base image (#10477)

### Dependencies updates: 
* Bump x/net (#10514)
* Bump curl and Go version (#10503)
* Bump google.golang.org/grpc from 1.58.2 to 1.58.3 (#10496)
* Bump github.com/prometheus/client_model (#10486)
* Bump ossf/scorecard-action from 2.2.0 to 2.3.0 (#10487)
* Bump golang.org/x/crypto from 0.13.0 to 0.14.0 (#10485)
 
**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-controller-v1.9.1...controller-controller-v1.9.3
