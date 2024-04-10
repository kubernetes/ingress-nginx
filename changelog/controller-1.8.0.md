# Changelog

### 1.8.0

Images:

* registry.k8s.io/ingress-nginx/controller:v1.8.0@sha256:744ae2afd433a395eeb13dc03d3313facba92e96ad71d9feaafc85925493fee3
* registry.k8s.io/ingress-nginx/controller-chroot:v1.8.0@sha256:a45e41cd2b7670adf829759878f512d4208d0aec1869dae593a0fecd09a5e49e

### Important changes:

* Validate path types (#9967)
* images: upgrade to Alpine 3.18 (#9997)
* Update documentation to reflect project name; Ingress-Nginx Controller

For improving security, our 1.8.0 release includes a [new, **optional** validation ](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#strict-validate-path-type) that limits the characters accepted  on ".spec paths.path" when pathType=Exact or athType=Prefix, to alphanumeric characters only.

More information can be found on our [Google doc](https://docs.google.com/document/d/1HPvaEwHRuMSkXYkVIJ-w7IpijKdHfNynm_4N2Akt0CQ/edit?usp=sharing), our new [ingress-nginx-dev mailing list](https://groups.google.com/a/kubernetes.io/g/ingress-nginx-dev/c/ebbBMo-zX-w) or in our [docs](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#strict-validate-path-type)

### Community Updates

We are now posting updates and release to our twitter handle, [@IngressNginx](https://twitter.com/IngressNGINX) and
on our new [ingress-nginx-dev mailing list](https://groups.google.com/a/kubernetes.io/g/ingress-nginx-dev/c/ebbBMo-zX-w)

### All changes:

* Add legacy to OpenTelemetry migration doc (#10011)
* changed tagsha to recent builds (#10001)
* change to alpine318 baseimage (#10000)
* images: upgrade to Alpine 3.18 (#9997)
* openssl CVE fix (#9996)
* PodDisruptionBudget spec logic update (#9904)
* Admission warning (#9975)
* Add OPA examples on pathType restrictions (#9992)
* updated testrunner image tag+sha (#9987)
* bumped ginkgo to v2.9.5 (#9985)
* helm: Fix opentelemetry module installation for daemonset (#9792)
* OpenTelemetry default config (#9978)
* Correct annotations in monitoring docs (#9976)
* fix: avoid builds and tests for changes to markdown (#9962)
* Validate path types (#9967)
* HPA: Use capabilites & align manifests. (#9521)
* Use dl.k8s.io instead of hardcoded GCS URIs (#9946)
* add option for annotations in PodDisruptionBudget (#9843)
* chore: update httpbin to httpbun (#9919)
* image_update (#9942)
* Add geoname id value into $geoip2_*_geoname_id variables (#9527)
* Update annotations.md (#9933)
* Update charts/* to keep project name display aligned (#9931)
* Keep project name display aligned (#9920)

### Dependencies updates:
* Bump github.com/imdario/mergo from 0.3.15 to 0.3.16 (#10008)
* Bump github.com/prometheus/common from 0.43.0 to 0.44.0 (#10007)
* Bump k8s.io/klog/v2 from 2.90.1 to 2.100.1 (#9913)
* Bump github.com/onsi/ginkgo/v2 from 2.9.0 to 2.9.5 (#9980)
* Bump golang.org/x/crypto from 0.8.0 to 0.9.0 (#9982)
* Bump actions/setup-go from 4.0.0 to 4.0.1 (#9984)
* Bump securego/gosec from 2.15.0 to 2.16.0 (#9983)
* Bump github.com/prometheus/common from 0.42.0 to 0.43.0 (#9981)
* Bump github.com/prometheus/client_model from 0.3.0 to 0.4.0 (#9937)
* Bump google.golang.org/grpc from 1.54.0 to 1.55.0 (#9936)

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/controller-controller-v1.7.1...controller-controller-v1.8.0
