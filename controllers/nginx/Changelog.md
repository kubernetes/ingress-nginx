Changelog

### 0.9-beta.1

**Image:**  `gcr.io/google_containers/nginx-ingress-controller:0.9.0-beta.1`

*New Features:*

- SSL Passthrough
- New Flag `--publish-service` that set the Service fronting the ingress controllers
- Ingress status shows the correct IP/hostname address without duplicates
- Custom body sizes per Ingress
- Prometheus metrics


*Breaking changes:*

- Flag `--nginx-configmap` was replaced with `--configmap`
- Configmap field `body-size` was replaced with `proxy-body-size`

*Changes:*

- [X] [#184](https://github.com/kubernetes/ingress/pull/184) Fix template error
- [X] [#179](https://github.com/kubernetes/ingress/pull/179) Allows the usage of Default SSL Cert
- [X] [#178](https://github.com/kubernetes/ingress/pull/178) Add initialization of proxy variable
- [X] [#177](https://github.com/kubernetes/ingress/pull/177) Refactoring sysctlFSFileMax helper
- [X] [#176](https://github.com/kubernetes/ingress/pull/176) Fix TLS does not get updated when changed
- [X] [#174](https://github.com/kubernetes/ingress/pull/174) Update nginx to 1.11.9
- [X] [#172](https://github.com/kubernetes/ingress/pull/172) add some unit test cases for some packages under folder "core.pkg.ingress"
- [X] [#168](https://github.com/kubernetes/ingress/pull/168) Changes the SSL Temp file to something inside the same SSL Directory
- [X] [#165](https://github.com/kubernetes/ingress/pull/165) Fix rate limit issue when more than 2 servers enabled in ingress
- [X] [#161](https://github.com/kubernetes/ingress/pull/161) Document some missing parameters and their defaults for NGINX controller
- [X] [#158](https://github.com/kubernetes/ingress/pull/158) prefect unit test cases for annotation.proxy
- [X] [#156](https://github.com/kubernetes/ingress/pull/156) Fix issue for ratelimit
- [X] [#154](https://github.com/kubernetes/ingress/pull/154) add unit test cases for core.pkg.ingress.annotations.cors
- [X] [#151](https://github.com/kubernetes/ingress/pull/151) Port in redirect
- [X] [#150](https://github.com/kubernetes/ingress/pull/150) Add support for custom header sizes
- [X] [#149](https://github.com/kubernetes/ingress/pull/149) Add flag to allow switch off the update of Ingress status
- [X] [#148](https://github.com/kubernetes/ingress/pull/148) Add annotation to allow custom body sizes
- [X] [#145](https://github.com/kubernetes/ingress/pull/145) fix wrong links and punctuations
- [X] [#144](https://github.com/kubernetes/ingress/pull/144) add unit test cases for core.pkg.k8s
- [X] [#143](https://github.com/kubernetes/ingress/pull/143) Use protobuf instead of rest to connect to apiserver host and add troubleshooting doc
- [X] [#142](https://github.com/kubernetes/ingress/pull/142) Use system fs.max-files as limits instead of hard-coded value
- [X] [#141](https://github.com/kubernetes/ingress/pull/141) Add reuse port and backlog to port 80 and 443
- [X] [#138](https://github.com/kubernetes/ingress/pull/138) reference to const
- [X] [#136](https://github.com/kubernetes/ingress/pull/136) Add content and descriptions about nginx's configuration
- [X] [#135](https://github.com/kubernetes/ingress/pull/135) correct improper punctuation
- [X] [#134](https://github.com/kubernetes/ingress/pull/134) fix typo
- [X] [#133](https://github.com/kubernetes/ingress/pull/133) Add TCP and UDP services removed in migration
- [X] [#132](https://github.com/kubernetes/ingress/pull/132) Document nginx controller configuration tweaks
- [X] [#128](https://github.com/kubernetes/ingress/pull/128) Add tests and godebug to compare structs
- [X] [#126](https://github.com/kubernetes/ingress/pull/126) change the type of imagePullPolicy
- [X] [#123](https://github.com/kubernetes/ingress/pull/123) Add resolver configuration to nginx
- [X] [#119](https://github.com/kubernetes/ingress/pull/119) add unit test case for annotations.service
- [X] [#115](https://github.com/kubernetes/ingress/pull/115) add default_server to listen statement for default backend
- [X] [#114](https://github.com/kubernetes/ingress/pull/114) fix typo
- [X] [#113](https://github.com/kubernetes/ingress/pull/113) Add condition of enqueue and unit test cases for task.Queue
- [X] [#108](https://github.com/kubernetes/ingress/pull/108) annotations: print error and skip if malformed
- [X] [#107](https://github.com/kubernetes/ingress/pull/107) fix some wrong links of examples which to be used for nginx
- [X] [#103](https://github.com/kubernetes/ingress/pull/103) Update the nginx controller manifests
- [X] [#101](https://github.com/kubernetes/ingress/pull/101) Add unit test for strings.StringInSlice
- [X] [#99](https://github.com/kubernetes/ingress/pull/99) Update nginx to 1.11.8
- [X] [#97](https://github.com/kubernetes/ingress/pull/97) Fix gofmt
- [X] [#96](https://github.com/kubernetes/ingress/pull/96) Fix typo PassthrougBackends -> PassthroughBackends
- [X] [#95](https://github.com/kubernetes/ingress/pull/95) Deny location mapping in case of specific errors
- [X] [#94](https://github.com/kubernetes/ingress/pull/94) Add support to disable server_tokens directive
- [X] [#93](https://github.com/kubernetes/ingress/pull/93) Fix sort for catch all server
- [X] [#92](https://github.com/kubernetes/ingress/pull/92) Refactoring of nginx configuration deserialization
- [X] [#91](https://github.com/kubernetes/ingress/pull/91) Fix x-forwarded-port mapping
- [X] [#90](https://github.com/kubernetes/ingress/pull/90) fix the wrong link to build/test/release
- [X] [#89](https://github.com/kubernetes/ingress/pull/89) fix the wrong links to the examples and developer documentation
- [X] [#88](https://github.com/kubernetes/ingress/pull/88) Fix multiple tls hosts sharing the same secretName
- [X] [#86](https://github.com/kubernetes/ingress/pull/86) Update X-Forwarded-Port
- [X] [#82](https://github.com/kubernetes/ingress/pull/82) Fix incorrect X-Forwarded-Port for TLS
- [X] [#81](https://github.com/kubernetes/ingress/pull/81) Do not push containers to remote repo as part of test-e2e
- [X] [#78](https://github.com/kubernetes/ingress/pull/78) Fix #76: hardcode X-Forwarded-Port due to SSL Passthrough
- [X] [#77](https://github.com/kubernetes/ingress/pull/77) Add support for IPV6 in dns resolvers
- [X] [#66](https://github.com/kubernetes/ingress/pull/66) Start FAQ docs
- [X] [#65](https://github.com/kubernetes/ingress/pull/65) Support hostnames in Ingress status
- [X] [#64](https://github.com/kubernetes/ingress/pull/64) Sort whitelist list to avoid random orders
- [X] [#62](https://github.com/kubernetes/ingress/pull/62) Fix e2e make targets
- [X] [#61](https://github.com/kubernetes/ingress/pull/61) Ignore coverage profile files
- [X] [#58](https://github.com/kubernetes/ingress/pull/58) Fix "invalid port in upstream" on nginx controller
- [X] [#57](https://github.com/kubernetes/ingress/pull/57) Fix invalid port in upstream
- [X] [#54](https://github.com/kubernetes/ingress/pull/54) Expand developer docs
- [X] [#52](https://github.com/kubernetes/ingress/pull/52) fix typo in variable ProxyRealIPCIDR
- [X] [#44](https://github.com/kubernetes/ingress/pull/44) Bump nginx version to one higher than that in contrib
- [X] [#36](https://github.com/kubernetes/ingress/pull/36) Add nginx metrics to prometheus
- [X] [#34](https://github.com/kubernetes/ingress/pull/34) nginx: also listen on ipv6
- [X] [#32](https://github.com/kubernetes/ingress/pull/32) Restart nginx if master process dies
- [X] [#31](https://github.com/kubernetes/ingress/pull/31) Add healthz checker
- [X] [#25](https://github.com/kubernetes/ingress/pull/25) Fix a data race in TestFileWatcher
- [X] [#12](https://github.com/kubernetes/ingress/pull/12) Split implementations from generic code
- [X] [#10](https://github.com/kubernetes/ingress/pull/10) Copy Ingress history from kubernetes/contrib
- [X] [#1498](https://github.com/kubernetes/contrib/pull/1498) Refactoring of template handling
- [X] [#1571](https://github.com/kubernetes/contrib/pull/1571) use POD_NAMESPACE as a namespace in cli parameters
- [X] [#1591](https://github.com/kubernetes/contrib/pull/1591) Always listen on port 443, even without ingress rules
- [X] [#1596](https://github.com/kubernetes/contrib/pull/1596) Adapt nginx hash sizes to the number of ingress
- [X] [#1653](https://github.com/kubernetes/contrib/pull/1653) Update image version
- [X] [#1672](https://github.com/kubernetes/contrib/pull/1672) Add firewall rules and ing class clarifications
- [X] [#1711](https://github.com/kubernetes/contrib/pull/1711) Add function helpers to nginx template
- [X] [#1743](https://github.com/kubernetes/contrib/pull/1743) Allow customisation of the nginx proxy_buffer_size directive via ConfigMap
- [X] [#1749](https://github.com/kubernetes/contrib/pull/1749) Readiness probe that works behind a CP lb
- [X] [#1751](https://github.com/kubernetes/contrib/pull/1751) Add the name of the upstream in the log
- [X] [#1758](https://github.com/kubernetes/contrib/pull/1758) Update nginx to 1.11.4
- [X] [#1759](https://github.com/kubernetes/contrib/pull/1759) Add support for default backend in Ingress rule
- [X] [#1762](https://github.com/kubernetes/contrib/pull/1762) Add cloud detection
- [X] [#1766](https://github.com/kubernetes/contrib/pull/1766) Clarify the controller uses endpoints and not services
- [X] [#1767](https://github.com/kubernetes/contrib/pull/1767) Update godeps
- [X] [#1772](https://github.com/kubernetes/contrib/pull/1772) Avoid replacing nginx.conf file if the new configuration is invalid
- [X] [#1773](https://github.com/kubernetes/contrib/pull/1773) Add annotation to add CORS support
- [X] [#1786](https://github.com/kubernetes/contrib/pull/1786) Add docs about go template
- [X] [#1796](https://github.com/kubernetes/contrib/pull/1796) Add external authentication support using auth_request
- [X] [#1802](https://github.com/kubernetes/contrib/pull/1802) Initialize proxy_upstream_name variable
- [X] [#1806](https://github.com/kubernetes/contrib/pull/1806) Add docs about the log format
- [X] [#1808](https://github.com/kubernetes/contrib/pull/1808) WebSocket documentation
- [X] [#1847](https://github.com/kubernetes/contrib/pull/1847) Change structure of packages
- [X] Add annotation for custom upstream timeouts
- [X] Mutual TLS auth (https://github.com/kubernetes/contrib/issues/1870)

### 0.8.3

- [X] [#1450](https://github.com/kubernetes/contrib/pull/1450) Check for errors in nginx template
- [ ] [#1498](https://github.com/kubernetes/contrib/pull/1498) Refactoring of template handling
- [X] [#1467](https://github.com/kubernetes/contrib/pull/1467) Use ClientConfig to configure connection
- [X] [#1575](https://github.com/kubernetes/contrib/pull/1575) Update nginx to 1.11.3

### 0.8.2

- [X] [#1336](https://github.com/kubernetes/contrib/pull/1336) Add annotation to skip ingress rule
- [X] [#1338](https://github.com/kubernetes/contrib/pull/1338) Add HTTPS default backend
- [X] [#1351](https://github.com/kubernetes/contrib/pull/1351) Avoid generation of invalid ssl certificates
- [X] [#1379](https://github.com/kubernetes/contrib/pull/1379) improve nginx performance
- [X] [#1350](https://github.com/kubernetes/contrib/pull/1350) Improve performance (listen backlog=net.core.somaxconn)
- [X] [#1384](https://github.com/kubernetes/contrib/pull/1384) Unset Authorization header when proxying
- [X] [#1398](https://github.com/kubernetes/contrib/pull/1398) Mitigate HTTPoxy Vulnerability

### 0.8.1

- [X] [#1317](https://github.com/kubernetes/contrib/pull/1317) Fix duplicated real_ip_header
- [X] [#1315](https://github.com/kubernetes/contrib/pull/1315) Addresses #1314

### 0.8

- [X] [#1063](https://github.com/kubernetes/contrib/pull/1063) watches referenced tls secrets
- [X] [#850](https://github.com/kubernetes/contrib/pull/850) adds configurable SSL redirect nginx controller
- [X] [#1136](https://github.com/kubernetes/contrib/pull/1136) Fix nginx rewrite rule order
- [X] [#1144](https://github.com/kubernetes/contrib/pull/1144) Add cidr whitelist support
- [X] [#1230](https://github.com/kubernetes/contrib/pull/1130) Improve docs and examples
- [X] [#1258](https://github.com/kubernetes/contrib/pull/1258) Avoid sync without a reachable
- [X] [#1235](https://github.com/kubernetes/contrib/pull/1235) Fix stats by country in nginx status page
- [X] [#1236](https://github.com/kubernetes/contrib/pull/1236) Update nginx to add dynamic TLS records and spdy
- [X] [#1238](https://github.com/kubernetes/contrib/pull/1238) Add support for dynamic TLS records and spdy
- [X] [#1239](https://github.com/kubernetes/contrib/pull/1239) Add support for conditional log of urls
- [X] [#1253](https://github.com/kubernetes/contrib/pull/1253) Use delayed queue
- [X] [#1296](https://github.com/kubernetes/contrib/pull/1296) Fix formatting
- [X] [#1299](https://github.com/kubernetes/contrib/pull/1299) Fix formatting

### 0.7

- [X] [#898](https://github.com/kubernetes/contrib/pull/898) reorder locations. Location / must be the last one to avoid errors routing to subroutes
- [X] [#946](https://github.com/kubernetes/contrib/pull/946) Add custom authentication (Basic or Digest) to ingress rules
- [X] [#926](https://github.com/kubernetes/contrib/pull/926) Custom errors should be optional
- [X] [#1002](https://github.com/kubernetes/contrib/pull/1002) Use k8s probes (disable NGINX checks)
- [X] [#962](https://github.com/kubernetes/contrib/pull/962) Make optional http2
- [X] [#1054](https://github.com/kubernetes/contrib/pull/1054) force reload if some certificate change
- [X] [#958](https://github.com/kubernetes/contrib/pull/958) update NGINX to 1.11.0 and add digest module
- [X] [#960](https://github.com/kubernetes/contrib/issues/960) https://trac.nginx.org/nginx/changeset/ce94f07d50826fcc8d48f046fe19d59329420fdb/nginx
- [X] [#1057](https://github.com/kubernetes/contrib/pull/1057) Remove loadBalancer ip on shutdown
- [X] [#1079](https://github.com/kubernetes/contrib/pull/1079) path rewrite
- [X] [#1093](https://github.com/kubernetes/contrib/pull/1093) rate limiting
- [X] [#1102](https://github.com/kubernetes/contrib/pull/1102) geolocation of traffic in stats
- [X] [#884](https://github.com/kubernetes/contrib/issues/884) support services running ssl
- [X] [#930](https://github.com/kubernetes/contrib/issues/930) detect changes in configuration configmaps
