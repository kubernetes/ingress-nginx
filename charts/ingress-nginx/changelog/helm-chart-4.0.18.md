# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.0.18

* [8291](https://github.com/kubernetes/ingress-nginx/pull/8291) remove git tag env from cloud build
* [8286](https://github.com/kubernetes/ingress-nginx/pull/8286) Fix OpenTelemetry sidecar image build
* [8277](https://github.com/kubernetes/ingress-nginx/pull/8277) Add OpenSSF Best practices badge
* [8273](https://github.com/kubernetes/ingress-nginx/pull/8273) Issue#8241
* [8267](https://github.com/kubernetes/ingress-nginx/pull/8267) Add fsGroup value to admission-webhooks/job-patch charts
* [8262](https://github.com/kubernetes/ingress-nginx/pull/8262) Updated confusing error
* [8256](https://github.com/kubernetes/ingress-nginx/pull/8256) fix: deny locations with invalid auth-url annotation
* [8253](https://github.com/kubernetes/ingress-nginx/pull/8253) Add a certificate info metric
* [8236](https://github.com/kubernetes/ingress-nginx/pull/8236) webhook: remove useless code.
* [8227](https://github.com/kubernetes/ingress-nginx/pull/8227) Update libraries in webhook image
* [8225](https://github.com/kubernetes/ingress-nginx/pull/8225) fix inconsistent-label-cardinality for prometheus metrics: nginx_ingress_controller_requests
* [8221](https://github.com/kubernetes/ingress-nginx/pull/8221) Do not validate ingresses with unknown ingress class in admission webhook endpoint
* [8210](https://github.com/kubernetes/ingress-nginx/pull/8210) Bump github.com/prometheus/client_golang from 1.11.0 to 1.12.1
* [8209](https://github.com/kubernetes/ingress-nginx/pull/8209) Bump google.golang.org/grpc from 1.43.0 to 1.44.0
* [8204](https://github.com/kubernetes/ingress-nginx/pull/8204) Add Artifact Hub lint
* [8203](https://github.com/kubernetes/ingress-nginx/pull/8203) Fix Indentation of example and link to cert-manager tutorial
* [8201](https://github.com/kubernetes/ingress-nginx/pull/8201) feat(metrics): add path and method labels to requests countera
* [8199](https://github.com/kubernetes/ingress-nginx/pull/8199) use functional options to reduce number of methods creating an EchoDeployment
* [8196](https://github.com/kubernetes/ingress-nginx/pull/8196) docs: fix inconsistent controller annotation
* [8191](https://github.com/kubernetes/ingress-nginx/pull/8191) Using Go install for misspell
* [8186](https://github.com/kubernetes/ingress-nginx/pull/8186) prometheus+grafana using servicemonitor
* [8185](https://github.com/kubernetes/ingress-nginx/pull/8185) Append elements on match, instead of removing for cors-annotations
* [8179](https://github.com/kubernetes/ingress-nginx/pull/8179) Bump github.com/opencontainers/runc from 1.0.3 to 1.1.0
* [8173](https://github.com/kubernetes/ingress-nginx/pull/8173) Adding annotations to the controller service account
* [8163](https://github.com/kubernetes/ingress-nginx/pull/8163) Update the $req_id placeholder description
* [8162](https://github.com/kubernetes/ingress-nginx/pull/8162) Versioned static manifests
* [8159](https://github.com/kubernetes/ingress-nginx/pull/8159) Adding some geoip variables and default values
* [8155](https://github.com/kubernetes/ingress-nginx/pull/8155) #7271 feat: avoid-pdb-creation-when-default-backend-disabled-and-replicas-gt-1
* [8151](https://github.com/kubernetes/ingress-nginx/pull/8151) Automatically generate helm docs
* [8143](https://github.com/kubernetes/ingress-nginx/pull/8143) Allow to configure delay before controller exits
* [8136](https://github.com/kubernetes/ingress-nginx/pull/8136) add ingressClass option to helm chart - back compatibility with ingress.class annotations
* [8126](https://github.com/kubernetes/ingress-nginx/pull/8126) Example for JWT

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.0.15...helm-chart-4.0.18
