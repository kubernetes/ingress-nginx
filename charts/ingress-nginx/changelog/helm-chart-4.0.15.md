# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.0.15

* [8120] https://github.com/kubernetes/ingress-nginx/pull/8120    Update go in runner and release v1.1.1
* [8119] https://github.com/kubernetes/ingress-nginx/pull/8119    Update to go v1.17.6
* [8118] https://github.com/kubernetes/ingress-nginx/pull/8118    Remove deprecated libraries, update other libs
* [8117] https://github.com/kubernetes/ingress-nginx/pull/8117    Fix codegen errors
* [8115] https://github.com/kubernetes/ingress-nginx/pull/8115    chart/ghaction: set the correct permission to have access to push a release
* [8098] https://github.com/kubernetes/ingress-nginx/pull/8098    generating SHA for CA only certs in backend_ssl.go + comparison of P…
* [8088] https://github.com/kubernetes/ingress-nginx/pull/8088    Fix Edit this page link to use main branch
* [8072] https://github.com/kubernetes/ingress-nginx/pull/8072    Expose GeoIP2 Continent code as variable
* [8061] https://github.com/kubernetes/ingress-nginx/pull/8061    docs(charts): using helm-docs for chart
* [8058] https://github.com/kubernetes/ingress-nginx/pull/8058    Bump github.com/spf13/cobra from 1.2.1 to 1.3.0
* [8054] https://github.com/kubernetes/ingress-nginx/pull/8054    Bump google.golang.org/grpc from 1.41.0 to 1.43.0
* [8051] https://github.com/kubernetes/ingress-nginx/pull/8051    align bug report with feature request regarding kind documentation
* [8046] https://github.com/kubernetes/ingress-nginx/pull/8046    Report expired certificates (#8045)
* [8044] https://github.com/kubernetes/ingress-nginx/pull/8044    remove G109 check till gosec resolves issues
* [8042] https://github.com/kubernetes/ingress-nginx/pull/8042    docs_multiple_instances_one_cluster_ticket_7543
* [8041] https://github.com/kubernetes/ingress-nginx/pull/8041    docs: fix typo'd executable name
* [8035] https://github.com/kubernetes/ingress-nginx/pull/8035    Comment busy owners
* [8029] https://github.com/kubernetes/ingress-nginx/pull/8029    Add stream-snippet as a ConfigMap and Annotation option
* [8023] https://github.com/kubernetes/ingress-nginx/pull/8023    fix nginx compilation flags
* [8021] https://github.com/kubernetes/ingress-nginx/pull/8021    Disable default modsecurity_rules_file if modsecurity-snippet is specified
* [8019] https://github.com/kubernetes/ingress-nginx/pull/8019    Revise main documentation page
* [8018] https://github.com/kubernetes/ingress-nginx/pull/8018    Preserve order of plugin invocation
* [8015] https://github.com/kubernetes/ingress-nginx/pull/8015    Add newline indenting to admission webhook annotations
* [8014] https://github.com/kubernetes/ingress-nginx/pull/8014    Add link to example error page manifest in docs
* [8009] https://github.com/kubernetes/ingress-nginx/pull/8009    Fix spelling in documentation and top-level files
* [8008] https://github.com/kubernetes/ingress-nginx/pull/8008    Add relabelings in controller-servicemonitor.yaml
* [8003] https://github.com/kubernetes/ingress-nginx/pull/8003    Minor improvements (formatting, consistency) in install guide
* [8001] https://github.com/kubernetes/ingress-nginx/pull/8001    fix: go-grpc Dockerfile
* [7999] https://github.com/kubernetes/ingress-nginx/pull/7999    images: use k8s-staging-test-infra/gcb-docker-gcloud
* [7996] https://github.com/kubernetes/ingress-nginx/pull/7996    doc: improvement
* [7983] https://github.com/kubernetes/ingress-nginx/pull/7983    Fix a couple of misspellings in the annotations documentation.
* [7979] https://github.com/kubernetes/ingress-nginx/pull/7979    allow set annotations for admission Jobs
* [7977] https://github.com/kubernetes/ingress-nginx/pull/7977    Add ssl_reject_handshake to default server
* [7975] https://github.com/kubernetes/ingress-nginx/pull/7975    add legacy version update v0.50.0 to main changelog
* [7972] https://github.com/kubernetes/ingress-nginx/pull/7972    updated service upstream definition

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.0.14...helm-chart-4.0.15
