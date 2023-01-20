# Changelog

This file documents all notable changes to [ingress-nginx](https://github.com/kubernetes/ingress-nginx) Helm Chart. The release numbering uses [semantic versioning](http://semver.org).

### 5.0.0

* Upgrade golang to 1.19.4
* Upgrade to Alpine 3.17
* A breaking change on how ingress deals with pathType and paths on ingress objects. Implement pathType validation (#9511)
* Values: Add missing `controller.metrics.service.labels`. (#9501)

This release adds a breaking change on how ingress deals with pathType and paths on ingress objects.

Previously, one could add an ingress like:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: grafana
  namespace: default
  annotations:
    nginx.kubernetes.io/use-regex: true
spec:
  rules:
    - host: grafana.bla.com
      http:
        paths:
          - backend:
              service:
                name: grafana
                port:
                  number: 80
            path: /grafana[0-9]{6}
            pathType: Prefix
 ```

But this path is not conformant with RFC or even Kubernetes specifications that states that path field should only
contain a valid path, and specific cases should have a [pathType: ImplementationSpecific](https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types)

Now, to use any regex on path, **EVEN IF REGEX ANNOTATION OR REWRITE ANNOTATION ARE PRESENT** the pathType should be of type ImplementationSpecific.

- Accepted characters now are A-Za-z0-9-._~/
- The validation above is more wide if pathType is ImplementationSpecific, allowing also common regex characters like ^%$[](){}*+?
- The set of characters is configurable in ConfigMap option `path-additional-allowed-chars`
- A user can disable this behavior setting the ConfigMap entry `disable-pathtype-validation` to true

Disabling pathType validation will disable the uncommon characters on pathType Exact and Prefix,
but the additional Allowed Characters will still be respected. If user wants to add more characters they should change the default


**Full Changelog**: https://github.com/kubernetes/ingress-nginx/compare/helm-chart-4.4.3...helm-chart-5.0.0
