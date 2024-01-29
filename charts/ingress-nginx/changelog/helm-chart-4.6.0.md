# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 4.5.3

* docs(helm): fix value key in readme for enabling certManager (#9640)
* Upgrade alpine 3.17.2
* Upgrade golang 1.20
* Drop testing/support for Kubernetes 1.23
* docs(helm): fix value key in readme for enabling certManager (#9640)
* Update Ingress-Nginx version controller-v1.7.0
* feat: OpenTelemetry module integration (#9062)
* canary-weight-total annotation ignored in rule backends (#9729)
* fix controller psp's volume config (#9740)
* Fix several Helm YAML issues with extraModules and extraInitContainers (#9709)
* Chart: Drop `controller.headers`, rework DH param secret. (#9659)
* Deployment/DaemonSet: Label pods using `ingress-nginx.labels`. (#9732)
* HPA: autoscaling/v2beta1 deprecated, bump apiVersion to v2 for defaultBackend (#9731)
* Fix incorrect annotation name in upstream hashing configuration (#9617)

* Update Ingress-Nginx version controller-v1.7.0

**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.5.2...helm-chart-4.6.0
