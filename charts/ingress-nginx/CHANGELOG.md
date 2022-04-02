# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.0.18
"[8291](https://github.com/kubernetes/ingress-nginx/pull/8291) remove git tag env from cloud build"
"[8286](https://github.com/kubernetes/ingress-nginx/pull/8286) Fix OpenTelemetry sidecar image build"
"[8277](https://github.com/kubernetes/ingress-nginx/pull/8277) Add OpenSSF Best practices badge"
"[8273](https://github.com/kubernetes/ingress-nginx/pull/8273) Issue#8241"
"[8267](https://github.com/kubernetes/ingress-nginx/pull/8267) Add fsGroup value to admission-webhooks/job-patch charts"
"[8262](https://github.com/kubernetes/ingress-nginx/pull/8262) Updated confusing error"
"[8256](https://github.com/kubernetes/ingress-nginx/pull/8256) fix: deny locations with invalid auth-url annotation"
"[8253](https://github.com/kubernetes/ingress-nginx/pull/8253) Add a certificate info metric"
"[8236](https://github.com/kubernetes/ingress-nginx/pull/8236) webhook: remove useless code."
"[8227](https://github.com/kubernetes/ingress-nginx/pull/8227) Update libraries in webhook image"
"[8225](https://github.com/kubernetes/ingress-nginx/pull/8225) fix inconsistent-label-cardinality for prometheus metrics: nginx_ingress_controller_requests"
"[8221](https://github.com/kubernetes/ingress-nginx/pull/8221) Do not validate ingresses with unknown ingress class in admission webhook endpoint"
"[8210](https://github.com/kubernetes/ingress-nginx/pull/8210) Bump github.com/prometheus/client_golang from 1.11.0 to 1.12.1"
"[8209](https://github.com/kubernetes/ingress-nginx/pull/8209) Bump google.golang.org/grpc from 1.43.0 to 1.44.0"
"[8204](https://github.com/kubernetes/ingress-nginx/pull/8204) Add Artifact Hub lint"
"[8203](https://github.com/kubernetes/ingress-nginx/pull/8203) Fix Indentation of example and link to cert-manager tutorial"
"[8201](https://github.com/kubernetes/ingress-nginx/pull/8201) feat(metrics): add path and method labels to requests countera"
"[8199](https://github.com/kubernetes/ingress-nginx/pull/8199) use functional options to reduce number of methods creating an EchoDeployment"
"[8196](https://github.com/kubernetes/ingress-nginx/pull/8196) docs: fix inconsistent controller annotation"
"[8191](https://github.com/kubernetes/ingress-nginx/pull/8191) Using Go install for misspell"
"[8186](https://github.com/kubernetes/ingress-nginx/pull/8186) prometheus+grafana using servicemonitor"
"[8185](https://github.com/kubernetes/ingress-nginx/pull/8185) Append elements on match, instead of removing for cors-annotations"
"[8179](https://github.com/kubernetes/ingress-nginx/pull/8179) Bump github.com/opencontainers/runc from 1.0.3 to 1.1.0"
"[8173](https://github.com/kubernetes/ingress-nginx/pull/8173) Adding annotations to the controller service account"
"[8163](https://github.com/kubernetes/ingress-nginx/pull/8163) Update the $req_id placeholder description"
"[8162](https://github.com/kubernetes/ingress-nginx/pull/8162) Versioned static manifests"
"[8159](https://github.com/kubernetes/ingress-nginx/pull/8159) Adding some geoip variables and default values"
"[8155](https://github.com/kubernetes/ingress-nginx/pull/8155) #7271 feat: avoid-pdb-creation-when-default-backend-disabled-and-replicas-gt-1"
"[8151](https://github.com/kubernetes/ingress-nginx/pull/8151) Automatically generate helm docs"
"[8143](https://github.com/kubernetes/ingress-nginx/pull/8143) Allow to configure delay before controller exits"
"[8136](https://github.com/kubernetes/ingress-nginx/pull/8136) add ingressClass option to helm chart - back compatibility with ingress.class annotations"
"[8126](https://github.com/kubernetes/ingress-nginx/pull/8126) Example for JWT"


### 4.0.15

- [8120] https://github.com/kubernetes/ingress-nginx/pull/8120    Update go in runner and release v1.1.1
- [8119] https://github.com/kubernetes/ingress-nginx/pull/8119    Update to go v1.17.6
- [8118] https://github.com/kubernetes/ingress-nginx/pull/8118    Remove deprecated libraries, update other libs
- [8117] https://github.com/kubernetes/ingress-nginx/pull/8117    Fix codegen errors
- [8115] https://github.com/kubernetes/ingress-nginx/pull/8115    chart/ghaction: set the correct permission to have access to push a release
- [8098] https://github.com/kubernetes/ingress-nginx/pull/8098    generating SHA for CA only certs in backend_ssl.go + comparision of P…
- [8088] https://github.com/kubernetes/ingress-nginx/pull/8088    Fix Edit this page link to use main branch
- [8072] https://github.com/kubernetes/ingress-nginx/pull/8072    Expose GeoIP2 Continent code as variable
- [8061] https://github.com/kubernetes/ingress-nginx/pull/8061    docs(charts): using helm-docs for chart
- [8058] https://github.com/kubernetes/ingress-nginx/pull/8058    Bump github.com/spf13/cobra from 1.2.1 to 1.3.0
- [8054] https://github.com/kubernetes/ingress-nginx/pull/8054    Bump google.golang.org/grpc from 1.41.0 to 1.43.0
- [8051] https://github.com/kubernetes/ingress-nginx/pull/8051    align bug report with feature request regarding kind documentation
- [8046] https://github.com/kubernetes/ingress-nginx/pull/8046    Report expired certificates (#8045)
- [8044] https://github.com/kubernetes/ingress-nginx/pull/8044    remove G109 check till gosec resolves issues
- [8042] https://github.com/kubernetes/ingress-nginx/pull/8042    docs_multiple_instances_one_cluster_ticket_7543
- [8041] https://github.com/kubernetes/ingress-nginx/pull/8041    docs: fix typo'd executible name
- [8035] https://github.com/kubernetes/ingress-nginx/pull/8035    Comment busy owners
- [8029] https://github.com/kubernetes/ingress-nginx/pull/8029    Add stream-snippet as a ConfigMap and Annotation option
- [8023] https://github.com/kubernetes/ingress-nginx/pull/8023    fix nginx compilation flags
- [8021] https://github.com/kubernetes/ingress-nginx/pull/8021    Disable default modsecurity_rules_file if modsecurity-snippet is specified
- [8019] https://github.com/kubernetes/ingress-nginx/pull/8019    Revise main documentation page
- [8018] https://github.com/kubernetes/ingress-nginx/pull/8018    Preserve order of plugin invocation
- [8015] https://github.com/kubernetes/ingress-nginx/pull/8015    Add newline indenting to admission webhook annotations
- [8014] https://github.com/kubernetes/ingress-nginx/pull/8014    Add link to example error page manifest in docs
- [8009] https://github.com/kubernetes/ingress-nginx/pull/8009    Fix spelling in documentation and top-level files
- [8008] https://github.com/kubernetes/ingress-nginx/pull/8008    Add relabelings in controller-servicemonitor.yaml
- [8003] https://github.com/kubernetes/ingress-nginx/pull/8003    Minor improvements (formatting, consistency) in install guide
- [8001] https://github.com/kubernetes/ingress-nginx/pull/8001    fix: go-grpc Dockerfile
- [7999] https://github.com/kubernetes/ingress-nginx/pull/7999    images: use k8s-staging-test-infra/gcb-docker-gcloud
- [7996] https://github.com/kubernetes/ingress-nginx/pull/7996    doc: improvement
- [7983] https://github.com/kubernetes/ingress-nginx/pull/7983    Fix a couple of misspellings in the annotations documentation.
- [7979] https://github.com/kubernetes/ingress-nginx/pull/7979    allow set annotations for admission Jobs
- [7977] https://github.com/kubernetes/ingress-nginx/pull/7977    Add ssl_reject_handshake to defaul server
- [7975] https://github.com/kubernetes/ingress-nginx/pull/7975    add legacy version update v0.50.0 to main changelog
- [7972] https://github.com/kubernetes/ingress-nginx/pull/7972    updated service upstream definition

### 4.0.14

- [8061] https://github.com/kubernetes/ingress-nginx/pull/8061 Using helm-docs to populate values table in README.md

### 4.0.13

- [8008] https://github.com/kubernetes/ingress-nginx/pull/8008 Add relabelings in controller-servicemonitor.yaml

### 4.0.12

- [7978] https://github.com/kubernetes/ingress-nginx/pull/7979 Support custom annotations in admissions Jobs

### 4.0.11

- [7873] https://github.com/kubernetes/ingress-nginx/pull/7873 Makes the [appProtocol](https://kubernetes.io/docs/concepts/services-networking/_print/#application-protocol) field optional.

### 4.0.10

- [7964] https://github.com/kubernetes/ingress-nginx/pull/7964 Update controller version to v1.1.0

### 4.0.9

- [6992] https://github.com/kubernetes/ingress-nginx/pull/6992 Add ability to specify labels for all resources

### 4.0.7

- [7923] https://github.com/kubernetes/ingress-nginx/pull/7923 Release v1.0.5 of ingress-nginx
- [7806] https://github.com/kubernetes/ingress-nginx/pull/7806 Choice option for internal/external loadbalancer type service

### 4.0.6

- [7804] https://github.com/kubernetes/ingress-nginx/pull/7804 Release v1.0.4 of ingress-nginx
- [7651] https://github.com/kubernetes/ingress-nginx/pull/7651 Support ipFamilyPolicy and ipFamilies fields in Helm Chart
- [7798] https://github.com/kubernetes/ingress-nginx/pull/7798 Exoscale: use HTTP Healthcheck mode
- [7793] https://github.com/kubernetes/ingress-nginx/pull/7793 Update kube-webhook-certgen to v1.1.1

### 4.0.5

- [7740] https://github.com/kubernetes/ingress-nginx/pull/7740 Release v1.0.3 of ingress-nginx

### 4.0.3

- [7707] https://github.com/kubernetes/ingress-nginx/pull/7707 Release v1.0.2 of ingress-nginx

### 4.0.2 

- [7681] https://github.com/kubernetes/ingress-nginx/pull/7681 Release v1.0.1 of ingress-nginx

### 4.0.1 

- [7535] https://github.com/kubernetes/ingress-nginx/pull/7535 Release v1.0.0 ingress-nginx

### 3.34.0

- [7256] https://github.com/kubernetes/ingress-nginx/pull/7256 Add namespace field in the namespace scoped resource templates

### 3.33.0

- [7164] https://github.com/kubernetes/ingress-nginx/pull/7164 Update nginx to v1.20.1

### 3.32.0

- [7117] https://github.com/kubernetes/ingress-nginx/pull/7117 Add annotations for HPA

### 3.31.0

- [7137] https://github.com/kubernetes/ingress-nginx/pull/7137 Add support for custom probes

### 3.30.0

- [#7092](https://github.com/kubernetes/ingress-nginx/pull/7092) Removes the possibility of using localhost in ExternalNames as endpoints

### 3.29.0

- [X] [#6945](https://github.com/kubernetes/ingress-nginx/pull/7020) Add option to specify job label for ServiceMonitor

### 3.28.0

- [ ] [#6900](https://github.com/kubernetes/ingress-nginx/pull/6900) Support existing PSPs

### 3.27.0

- Update ingress-nginx v0.45.0

### 3.26.0

- [X] [#6979](https://github.com/kubernetes/ingress-nginx/pull/6979) Changed servicePort value for metrics

### 3.25.0

- [X] [#6957](https://github.com/kubernetes/ingress-nginx/pull/6957) Add ability to specify automountServiceAccountToken

### 3.24.0

- [X] [#6908](https://github.com/kubernetes/ingress-nginx/pull/6908) Add volumes to default-backend deployment

### 3.23.0

- Update ingress-nginx v0.44.0

### 3.22.0

- [X] [#6802](https://github.com/kubernetes/ingress-nginx/pull/6802) Add value for configuring a custom Diffie-Hellman parameters file
- [X] [#6815](https://github.com/kubernetes/ingress-nginx/pull/6815) Allow use of numeric namespaces in helm chart

### 3.21.0

- [X] [#6783](https://github.com/kubernetes/ingress-nginx/pull/6783) Add custom annotations to ScaledObject
- [X] [#6761](https://github.com/kubernetes/ingress-nginx/pull/6761) Adding quotes in the serviceAccount name in Helm values
- [X] [#6767](https://github.com/kubernetes/ingress-nginx/pull/6767) Remove ClusterRole when scope option is enabled
- [X] [#6785](https://github.com/kubernetes/ingress-nginx/pull/6785) Update kube-webhook-certgen image to v1.5.1

### 3.20.1

- Do not create KEDA in case of DaemonSets.
- Fix KEDA v2 definition

### 3.20.0

- [X] [#6730](https://github.com/kubernetes/ingress-nginx/pull/6730) Do not create HPA for defaultBackend if not enabled.

### 3.19.0

- Update ingress-nginx v0.43.0

### 3.18.0

- [X] [#6688](https://github.com/kubernetes/ingress-nginx/pull/6688) Allow volume-type emptyDir in controller podsecuritypolicy
- [X] [#6691](https://github.com/kubernetes/ingress-nginx/pull/6691) Improve parsing of helm parameters

### 3.17.0

- Update ingress-nginx v0.42.0

### 3.16.1

- Fix chart-releaser action

### 3.16.0

- [X] [#6646](https://github.com/kubernetes/ingress-nginx/pull/6646) Added LoadBalancerIP value for internal service

### 3.15.1

- Fix chart-releaser action

### 3.15.0

- [X] [#6586](https://github.com/kubernetes/ingress-nginx/pull/6586) Fix 'maxmindLicenseKey' location in values.yaml

### 3.14.0

- [X] [#6469](https://github.com/kubernetes/ingress-nginx/pull/6469) Allow custom service names for controller and backend

### 3.13.0

- [X] [#6544](https://github.com/kubernetes/ingress-nginx/pull/6544) Fix default backend HPA name variable

### 3.12.0

- [X] [#6514](https://github.com/kubernetes/ingress-nginx/pull/6514) Remove helm2 support and update docs

### 3.11.1

- [X] [#6505](https://github.com/kubernetes/ingress-nginx/pull/6505) Reorder HPA resource list to work with GitOps tooling

### 3.11.0

- Support Keda Autoscaling

### 3.10.1

- Fix regression introduced in 0.41.0 with external authentication

### 3.10.0

- Fix routing regression introduced in 0.41.0 with PathType Exact

### 3.9.0

- [X] [#6423](https://github.com/kubernetes/ingress-nginx/pull/6423) Add Default backend HPA autoscaling

### 3.8.0

- [X] [#6395](https://github.com/kubernetes/ingress-nginx/pull/6395) Update jettech/kube-webhook-certgen image
- [X] [#6377](https://github.com/kubernetes/ingress-nginx/pull/6377) Added loadBalancerSourceRanges for internal lbs
- [X] [#6356](https://github.com/kubernetes/ingress-nginx/pull/6356) Add securitycontext settings on defaultbackend
- [X] [#6401](https://github.com/kubernetes/ingress-nginx/pull/6401) Fix controller service annotations
- [X] [#6403](https://github.com/kubernetes/ingress-nginx/pull/6403) Initial helm chart changelog

### 3.7.1

- [X] [#6326](https://github.com/kubernetes/ingress-nginx/pull/6326) Fix liveness and readiness probe path in daemonset chart

### 3.7.0

- [X] [#6316](https://github.com/kubernetes/ingress-nginx/pull/6316) Numerals in podAnnotations in quotes [#6315](https://github.com/kubernetes/ingress-nginx/issues/6315)

### 3.6.0

- [X] [#6305](https://github.com/kubernetes/ingress-nginx/pull/6305) Add default linux nodeSelector

### 3.5.1

- [X] [#6299](https://github.com/kubernetes/ingress-nginx/pull/6299) Fix helm chart release

### 3.5.0

- [X] [#6260](https://github.com/kubernetes/ingress-nginx/pull/6260) Allow Helm Chart to customize admission webhook's annotations, timeoutSeconds, namespaceSelector, objectSelector and cert files locations

### 3.4.0

- [X] [#6268](https://github.com/kubernetes/ingress-nginx/pull/6268) Update to 0.40.2 in helm chart #6288

### 3.3.1

- [X] [#6259](https://github.com/kubernetes/ingress-nginx/pull/6259) Release helm chart
- [X] [#6258](https://github.com/kubernetes/ingress-nginx/pull/6258) Fix chart markdown link
- [X] [#6253](https://github.com/kubernetes/ingress-nginx/pull/6253) Release v0.40.0

### 3.3.1

- [X] [#6233](https://github.com/kubernetes/ingress-nginx/pull/6233) Add admission controller e2e test

### 3.3.0

- [X] [#6203](https://github.com/kubernetes/ingress-nginx/pull/6203) Refactor parsing of key values
- [X] [#6162](https://github.com/kubernetes/ingress-nginx/pull/6162) Add helm chart options to expose metrics service as NodePort
- [X] [#6180](https://github.com/kubernetes/ingress-nginx/pull/6180) Fix helm chart admissionReviewVersions regression
- [X] [#6169](https://github.com/kubernetes/ingress-nginx/pull/6169) Fix Typo in example prometheus rules

### 3.0.0

- [X] [#6167](https://github.com/kubernetes/ingress-nginx/pull/6167) Update chart requirements

### 2.16.0

- [X] [#6154](https://github.com/kubernetes/ingress-nginx/pull/6154) add `topologySpreadConstraint` to controller

### 2.15.0

- [X] [#6087](https://github.com/kubernetes/ingress-nginx/pull/6087) Adding parameter for externalTrafficPolicy in internal controller service spec

### 2.14.0

- [X] [#6104](https://github.com/kubernetes/ingress-nginx/pull/6104) Misc fixes for nginx-ingress chart for better keel and prometheus-operator integration

### 2.13.0

- [X] [#6093](https://github.com/kubernetes/ingress-nginx/pull/6093) Release v0.35.0

### 2.13.0

- [X] [#6093](https://github.com/kubernetes/ingress-nginx/pull/6093) Release v0.35.0
- [X] [#6080](https://github.com/kubernetes/ingress-nginx/pull/6080) Switch images to k8s.gcr.io after Vanity Domain Flip

### 2.12.1

- [X] [#6075](https://github.com/kubernetes/ingress-nginx/pull/6075) Sync helm chart affinity examples

### 2.12.0

- [X] [#6039](https://github.com/kubernetes/ingress-nginx/pull/6039) Add configurable serviceMonitor metricRelabelling and targetLabels
- [X] [#6044](https://github.com/kubernetes/ingress-nginx/pull/6044) Fix YAML linting

### 2.11.3

- [X] [#6038](https://github.com/kubernetes/ingress-nginx/pull/6038) Bump chart version PATCH

### 2.11.2

- [X] [#5951](https://github.com/kubernetes/ingress-nginx/pull/5951) Bump chart patch version

### 2.11.1

- [X] [#5900](https://github.com/kubernetes/ingress-nginx/pull/5900) Release helm chart for v0.34.1

### 2.11.0

- [X] [#5879](https://github.com/kubernetes/ingress-nginx/pull/5879) Update helm chart for v0.34.0
- [X] [#5671](https://github.com/kubernetes/ingress-nginx/pull/5671) Make liveness probe more fault tolerant than readiness probe

### 2.10.0

- [X] [#5843](https://github.com/kubernetes/ingress-nginx/pull/5843) Update jettech/kube-webhook-certgen image

### 2.9.1

- [X] [#5823](https://github.com/kubernetes/ingress-nginx/pull/5823) Add quoting to sysctls because numeric values need to be presented as strings (#5823)

### 2.9.0

- [X] [#5795](https://github.com/kubernetes/ingress-nginx/pull/5795) Use fully qualified images to avoid cri-o issues


### TODO

Keep building the changelog using *git log charts* checking the tag
