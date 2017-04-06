# Customize the HAProxy configuration

This example use a [ConfigMap](https://kubernetes.io/docs/user-guide/configmap/) to customize the HAProxy configuration.

## Prerequisites

This document has the following prerequisites:

* Deploy a service named `ingress-default-backend` to be used as default backend service
* Create a [TLS secret](/examples/PREREQUISITES.md#tls-certificates) named `tls-secret` to be used as default TLS certificate

As mentioned in the deployment instructions, you MUST turn down any existing
ingress controllers before running HAProxy Ingress.

## Customize the HAProxy configuration

Using a [ConfigMap](https://kubernetes.io/docs/user-guide/configmap/) is possible to customize the HAProxy configuration

For example, if we want to change the syslog-endpoint we need to create a ConfigMap:

```
$ kubectl create configmap haproxy-conf --from-literal=syslog-endpoint=172.17.8.101
```

Please check the example `haproxy-custom-configuration.yaml`

If the Configmap it is updated, HAProxy will be reloaded with the new configuration.
