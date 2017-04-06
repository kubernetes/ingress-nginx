# Customize the HAProxy configuration

This example use a [ConfigMap](https://kubernetes.io/docs/user-guide/configmap/) to customize the HAProxy configuration.

## Prerequisites

This document has the following prerequisites:

Deploy only the tls-secret and the default backend from the [deployment instructions](../../../deployment/haproxy/)

As mentioned in the deployment instructions, you MUST turn down any existing
ingress controllers before running HAProxy Ingress.

## Customize the HAProxy configuration

Using a [ConfigMap](https://kubernetes.io/docs/user-guide/configmap/) is possible to customize the HAProxy configuration.

For example, if we want to change the syslog-endpoint we need to create a ConfigMap:

```
$ kubectl create configmap haproxy-conf --from-literal=syslog-endpoint=172.17.8.101

```

Create the HAProxy Ingress deployment:
```
$ kubectl create -f haproxy-custom-configuration.yaml
```

The only difference from the deployment instructions is the --configmap parameter:
```
- --configmap=default/haproxy-conf
```

If the Configmap it is updated, HAProxy will be reloaded with the new configuration.

Check all the config options in the [HAProxy Ingress docs](https://github.com/jcmoraisjr/haproxy-ingress#configmap)