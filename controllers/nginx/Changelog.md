Changelog

### 0.9

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
