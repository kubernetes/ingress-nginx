# ingress-nginx

[ingress-nginx](https://github.com/kubernetes/ingress-nginx) Ingress controller for Kubernetes using NGINX as a reverse proxy and load balancer

To use, add the `kubernetes.io/ingress.class: nginx` annotation to your Ingress resources.

## TL;DR;

```console
$ helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
$ helm install my-release ingress-nginx/ingress-nginx
```

## Introduction

This chart bootstraps an ingress-nginx deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

  - Kubernetes 1.6+

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm install --name my-release ingress-nginx/ingress-nginx
```

The command deploys ingress-nginx on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```console
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the ingress-nginx chart and their default values.

Parameter | Description | Default
--- | --- | ---
`controller.image.repository` | controller container image repository | `quay.io/kubernetes-ingress-controller/nginx-ingress-controller`
`controller.image.tag` | controller container image tag | `0.30.0`
`controller.image.digest` | controller container image digest | `""`
`controller.image.pullPolicy` | controller container image pull policy | `IfNotPresent`
`controller.image.runAsUser` | User ID of the controller process. Value depends on the Linux distribution used inside of the container image. | `101`
`controller.containerPort.http` | The port that the controller container listens on for http connections. | `80`
`controller.containerPort.https` | The port that the controller container listens on for https connections. | `443`
`controller.config` | nginx [ConfigMap](https://github.com/kubernetes/ingress-nginx/blob/master/docs/user-guide/nginx-configuration/configmap.md) entries | none
`controller.configAnnotations` | annotations to be added to controller custom configuration configmap | `{}`
`controller.hostNetwork` | If the nginx deployment / daemonset should run on the host's network namespace. Do not set this when `controller.service.externalIPs` is set and `kube-proxy` is used as there will be a port-conflict for port `80` | false
`controller.dnsPolicy` | If using `hostNetwork=true`, change to `ClusterFirstWithHostNet`. See [pod's dns policy](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy) for details | `ClusterFirst`
`controller.dnsConfig` | custom pod dnsConfig. See [pod's dns config](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-config) for details | `{}`
`controller.reportNodeInternalIp` | If using `hostNetwork=true`, setting `reportNodeInternalIp=true`, will pass the flag `report-node-internal-ip-address` to ingress-nginx. This sets the status of all Ingress objects to the internal IP address of all nodes running the NGINX Ingress controller.
`controller.electionID` | election ID to use for the status update | `ingress-controller-leader`
`controller.extraEnvs` | any additional environment variables to set in the pods | `{}`
`controller.extraContainers` | Sidecar containers to add to the controller pod. See [LemonLDAP::NG controller](https://github.com/lemonldap-ng-controller/lemonldap-ng-controller) as example | `{}`
`controller.extraVolumeMounts` | Additional volumeMounts to the controller main container | `{}`
`controller.extraVolumes` | Additional volumes to the controller pod | `{}`
`controller.extraInitContainers` | Containers, which are run before the app containers are started | `[]`
`controller.healthCheckPath` | Path of the health check endpoint. All requests received on the port defined by the healthz-port parameter are forwarded internally to this path. | `/healthz"`
`controller.ingressClass` | name of the ingress class to route through this controller | `nginx`
`controller.maxmindLicenseKey` | Maxmind license key to download GeoLite2 Databases. See [Accessing and using GeoLite2 database](https://blog.maxmind.com/2019/12/18/significant-changes-to-accessing-and-using-geolite2-databases/) | `""`
`controller.scope.enabled` | limit the scope of the ingress controller | `false` (watch all namespaces)
`controller.scope.namespace` | namespace to watch for ingress | `""` (use the release namespace)
`controller.extraArgs` | Additional controller container arguments | `{}`
`controller.kind` | install as Deployment, DaemonSet or Both | `Deployment`
`controller.annotations` | annotations to be added to the Deployment or Daemonset | `{}`
`controller.autoscaling.enabled` | If true, creates Horizontal Pod Autoscaler | false
`controller.autoscaling.minReplicas` | If autoscaling enabled, this field sets minimum replica count | `2`
`controller.autoscaling.maxReplicas` | If autoscaling enabled, this field sets maximum replica count | `11`
`controller.autoscaling.targetCPUUtilizationPercentage` | Target CPU utilization percentage to scale | `"50"`
`controller.autoscaling.targetMemoryUtilizationPercentage` | Target memory utilization percentage to scale | `"50"`
`controller.autoscaling.autoscalingTemplate` | If autoscaling template provided, creates custom autoscaling metric | false
`controller.hostPort.enabled` | This enable `hostPort` for ports defined in TCP/80 and TCP/443 | false
`controller.hostPort.ports.http` | If `controller.hostPort.enabled` is `true` and this is non-empty, it sets the hostPort | `"80"`
`controller.hostPort.ports.https` | If `controller.hostPort.enabled` is `true` and this is non-empty, it sets the hostPort | `"443"`
`controller.tolerations` | node taints to tolerate (requires Kubernetes >=1.6) | `[]`
`controller.affinity` | node/pod affinities (requires Kubernetes >=1.6) | `{}`
`controller.terminationGracePeriodSeconds` | how many seconds to wait before terminating a pod | `60`
`controller.minReadySeconds` | how many seconds a pod needs to be ready before killing the next, during update | `0`
`controller.nodeSelector` | node labels for pod assignment | `{}`
`controller.podAnnotations` | annotations to be added to pods | `{}`
`controller.podLabels` | labels to add to the pod container metadata | `{}`
`controller.podSecurityContext` | Security context policies to add to the controller pod | `{}`
`controller.sysctls` | Map of optional sysctls to enable in the controller and in the PodSecurityPolicy | `{}`
`controller.replicaCount` | desired number of controller pods | `1`
`controller.minAvailable` | minimum number of available controller pods for PodDisruptionBudget | `1`
`controller.resources` | controller pod resource requests & limits | `{}`
`controller.priorityClassName` | controller priorityClassName | `nil`
`controller.lifecycle` | controller pod lifecycle hooks | `{}`
`controller.publishService.enabled` | if true, the controller will set the endpoint records on the ingress objects to reflect those on the service | `true`
`controller.publishService.pathOverride` | override of the default publish-service name | `""`
`controller.service.annotations` | annotations for controller service | `{}`
`controller.service.labels` | labels for controller service | `{}`
`controller.service.enabled` | if disabled no service will be created. This is especially useful when `controller.kind` is set to `DaemonSet` and `controller.hostPorts.enabled` is `true` | true
`controller.service.clusterIP` | internal controller cluster service IP (set to `"-"` to pass an empty value) | `nil`
`controller.service.omitClusterIP` | (Deprecated) To omit the `clusterIP` from the controller service | `false`
`controller.service.externalIPs` | controller service external IP addresses. Do not set this when `controller.hostNetwork` is set to `true` and `kube-proxy` is used as there will be a port-conflict for port `80` | `[]`
`controller.service.externalTrafficPolicy` | If `controller.service.type` is `NodePort` or `LoadBalancer`, set this to `Local` to enable [source IP preservation](https://kubernetes.io/docs/tutorials/services/source-ip/#source-ip-for-services-with-typenodeport) | `"Cluster"`
`controller.service.sessionAffinity` | Enables client IP based session affinity. Must be `ClientIP` or `None` if set.  | `""`
`controller.service.healthCheckNodePort` | If `controller.service.type` is `NodePort` or `LoadBalancer` and `controller.service.externalTrafficPolicy` is set to `Local`, set this to [the managed health-check port the kube-proxy will expose](https://kubernetes.io/docs/tutorials/services/source-ip/#source-ip-for-services-with-typenodeport). If blank, a random port in the `NodePort` range will be assigned | `""`
`controller.service.loadBalancerIP` | IP address to assign to load balancer (if supported) | `""`
`controller.service.loadBalancerSourceRanges` | list of IP CIDRs allowed access to load balancer (if supported) | `[]`
`controller.service.enableHttp` | if port 80 should be opened for service | `true`
`controller.service.enableHttps` | if port 443 should be opened for service | `true`
`controller.service.targetPorts.http` | Sets the targetPort that maps to the Ingress' port 80 | `80`
`controller.service.targetPorts.https` | Sets the targetPort that maps to the Ingress' port 443 | `443`
`controller.service.ports.http` | Sets service http port | `80`
`controller.service.ports.https` | Sets service https port | `443`
`controller.service.type` | type of controller service to create | `LoadBalancer`
`controller.service.nodePorts.http` | If `controller.service.type` is either `NodePort` or `LoadBalancer` and this is non-empty, it sets the nodePort that maps to the Ingress' port 80 | `""`
`controller.service.nodePorts.https` | If `controller.service.type` is either `NodePort` or `LoadBalancer` and this is non-empty, it sets the nodePort that maps to the Ingress' port 443 | `""`
`controller.service.nodePorts.tcp` | Sets the nodePort for an entry referenced by its key from `tcp` | `{}`
`controller.service.nodePorts.udp` | Sets the nodePort for an entry referenced by its key from `udp` | `{}`
`controller.service.internal.enabled` | Enables an (additional) internal load balancer | false
`controller.service.internal.annotations` | Annotations for configuring the additional internal load balancer | `{}`
`controller.livenessProbe.initialDelaySeconds` | Delay before liveness probe is initiated | 10
`controller.livenessProbe.periodSeconds` | How often to perform the probe | 10
`controller.livenessProbe.timeoutSeconds` | When the probe times out | 5
`controller.livenessProbe.successThreshold` | Minimum consecutive successes for the probe to be considered successful after having failed. | 1
`controller.livenessProbe.failureThreshold` | Minimum consecutive failures for the probe to be considered failed after having succeeded. | 3
`controller.livenessProbe.port` | The port number that the liveness probe will listen on. | 10254
`controller.readinessProbe.initialDelaySeconds` | Delay before readiness probe is initiated | 10
`controller.readinessProbe.periodSeconds` | How often to perform the probe | 10
`controller.readinessProbe.timeoutSeconds` | When the probe times out | 1
`controller.readinessProbe.successThreshold` | Minimum consecutive successes for the probe to be considered successful after having failed. | 1
`controller.readinessProbe.failureThreshold` | Minimum consecutive failures for the probe to be considered failed after having succeeded. | 3
`controller.readinessProbe.port` | The port number that the readiness probe will listen on. | 10254
`controller.metrics.enabled` | if `true`, enable Prometheus metrics | `false`
`controller.metrics.service.annotations` | annotations for Prometheus metrics service | `{}`
`controller.metrics.service.clusterIP` | cluster IP address to assign to service (set to `"-"` to pass an empty value) | `nil`
`controller.metrics.service.omitClusterIP` | (Deprecated) To omit the `clusterIP` from the metrics service | `false`
`controller.metrics.service.externalIPs` | Prometheus metrics service external IP addresses | `[]`
`controller.metrics.service.labels` | labels for metrics service | `{}`
`controller.metrics.service.loadBalancerIP` | IP address to assign to load balancer (if supported) | `""`
`controller.metrics.service.loadBalancerSourceRanges` | list of IP CIDRs allowed access to load balancer (if supported) | `[]`
`controller.metrics.service.servicePort` | Prometheus metrics service port | `9913`
`controller.metrics.service.type` | type of Prometheus metrics service to create | `ClusterIP`
`controller.metrics.serviceMonitor.enabled` | Set this to `true` to create ServiceMonitor for Prometheus operator | `false`
`controller.metrics.serviceMonitor.additionalLabels` | Additional labels that can be used so ServiceMonitor will be discovered by Prometheus | `{}`
`controller.metrics.serviceMonitor.honorLabels` | honorLabels chooses the metric's labels on collisions with target labels. | `false`
`controller.metrics.serviceMonitor.namespace` | namespace where servicemonitor resource should be created | `the same namespace as nginx ingress`
`controller.metrics.serviceMonitor.namespaceSelector` | [namespaceSelector](https://github.com/coreos/prometheus-operator/blob/v0.34.0/Documentation/api.md#namespaceselector) to configure what namespaces to scrape | `will scrape the helm release namespace only`
`controller.metrics.serviceMonitor.scrapeInterval` | interval between Prometheus scraping | `30s`
`controller.metrics.prometheusRule.enabled` | Set this to `true` to create prometheusRules for Prometheus operator | `false`
`controller.metrics.prometheusRule.additionalLabels` | Additional labels that can be used so prometheusRules will be discovered by Prometheus | `{}`
`controller.metrics.prometheusRule.namespace` | namespace where prometheusRules resource should be created | `the same namespace as nginx ingress`
`controller.metrics.prometheusRule.rules` | [rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) to be prometheus in YAML format, check values for an example. | `[]`
`controller.admissionWebhooks.enabled` | Create Ingress admission webhooks. Validating webhook will check the ingress syntax. | `true`
`controller.admissionWebhooks.failurePolicy` | Failure policy for admission webhooks | `Fail`
`controller.admissionWebhooks.port` | Admission webhook port | `8443`
`controller.admissionWebhooks.service.annotations` | Annotations for admission webhook service | `{}`
`controller.admissionWebhooks.service.omitClusterIP` | (Deprecated) To omit the `clusterIP` from the admission webhook service | `false`
`controller.admissionWebhooks.service.clusterIP` | cluster IP address to assign to admission webhook service (set to `"-"` to pass an empty value) | `nil`
`controller.admissionWebhooks.service.externalIPs` | Admission webhook service external IP addresses | `[]`
`controller.admissionWebhooks.service.loadBalancerIP` | IP address to assign to load balancer (if supported) | `""`
`controller.admissionWebhooks.service.loadBalancerSourceRanges` | List of IP CIDRs allowed access to load balancer (if supported) | `[]`
`controller.admissionWebhooks.service.servicePort` | Admission webhook service port | `443`
`controller.admissionWebhooks.service.type` | Type of admission webhook service to create | `ClusterIP`
`controller.admissionWebhooks.patch.enabled` | If true, will use a pre and post install hooks to generate a CA and certificate to use for the prometheus operator tls proxy, and patch the created webhooks with the CA. | `true`
`controller.admissionWebhooks.patch.image.repository` | Repository to use for the webhook integration jobs | `jettech/kube-webhook-certgen`
`controller.admissionWebhooks.patch.image.tag` |  Tag to use for the webhook integration jobs | `v1.2.0`
`controller.admissionWebhooks.patch.image.digest` |  Digest to use for the webhook integration jobs | `""`
`controller.admissionWebhooks.patch.image.pullPolicy` | Image pull policy for the webhook integration jobs | `IfNotPresent`
`controller.admissionWebhooks.patch.priorityClassName` | Priority class for the webhook integration jobs | `""`
`controller.admissionWebhooks.patch.podAnnotations` | Annotations for the webhook job pods | `{}`
`controller.admissionWebhooks.patch.nodeSelector` | Node selector for running admission hook patch jobs | `{}`
`controller.admissionWebhooks.patch.tolerations` | Node taints/tolerations for running admission hook patch jobs | `[]`
`controller.customTemplate.configMapName` | configMap containing a custom nginx template | `""`
`controller.customTemplate.configMapKey` | configMap key containing the nginx template | `""`
`controller.addHeaders` | configMap key:value pairs containing [custom headers](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#add-headers) added before sending response to the client | `{}`
`controller.proxySetHeaders` | configMap key:value pairs containing [custom headers](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#proxy-set-headers) added before sending request to the backends| `{}`
`controller.headers` | DEPRECATED, Use `controller.proxySetHeaders` instead. | `{}`
`controller.updateStrategy` | allows setting of RollingUpdate strategy | `{}`
`controller.configMapNamespace` | The nginx-configmap namespace name | `""`
`controller.tcp.configMapNamespace` | The tcp-services-configmap namespace name | `""`
`controller.tcp.annotations` | annotations to be added to tcp configmap | `{}`
`controller.udp.configMapNamespace` | The udp-services-configmap namespace name | `""`
`controller.udp.annotations` | annotations to be added to udp configmap | `{}`
`defaultBackend.enabled` | Use default backend component | `false`
`defaultBackend.image.repository` | default backend container image repository | `k8s.gcr.io/defaultbackend-amd64`
`defaultBackend.image.tag` | default backend container image tag | `1.5`
`defaultBackend.image.digest` | default backend container image digest | `""`
`defaultBackend.image.pullPolicy` | default backend container image pull policy | `IfNotPresent`
`defaultBackend.image.runAsUser` | User ID of the controller process. Value depends on the Linux distribution used inside of the container image. By default uses nobody user. | `65534`
`defaultBackend.extraArgs` | Additional default backend container arguments | `{}`
`defaultBackend.extraEnvs` | any additional environment variables to set in the defaultBackend pods | `[]`
`defaultBackend.port` | Http port number | `8080`
`defaultBackend.livenessProbe.initialDelaySeconds` | Delay before liveness probe is initiated | 30
`defaultBackend.livenessProbe.periodSeconds` | How often to perform the probe | 10
`defaultBackend.livenessProbe.timeoutSeconds` | When the probe times out | 5
`defaultBackend.livenessProbe.successThreshold` | Minimum consecutive successes for the probe to be considered successful after having failed. | 1
`defaultBackend.livenessProbe.failureThreshold` | Minimum consecutive failures for the probe to be considered failed after having succeeded. | 3
`defaultBackend.readinessProbe.initialDelaySeconds` | Delay before readiness probe is initiated | 0
`defaultBackend.readinessProbe.periodSeconds` | How often to perform the probe | 5
`defaultBackend.readinessProbe.timeoutSeconds` | When the probe times out | 5
`defaultBackend.readinessProbe.successThreshold` | Minimum consecutive successes for the probe to be considered successful after having failed. | 1
`defaultBackend.readinessProbe.failureThreshold` | Minimum consecutive failures for the probe to be considered failed after having succeeded. | 6
`defaultBackend.tolerations` | node taints to tolerate (requires Kubernetes >=1.6) | `[]`
`defaultBackend.affinity` | node/pod affinities (requires Kubernetes >=1.6) | `{}`
`defaultBackend.nodeSelector` | node labels for pod assignment | `{}`
`defaultBackend.podAnnotations` | annotations to be added to pods | `{}`
`defaultBackend.podLabels` | labels to add to the pod container metadata | `{}`
`defaultBackend.replicaCount` | desired number of default backend pods | `1`
`defaultBackend.minAvailable` | minimum number of available default backend pods for PodDisruptionBudget | `1`
`defaultBackend.resources` | default backend pod resource requests & limits | `{}`
`defaultBackend.priorityClassName` | default backend  priorityClassName | `nil`
`defaultBackend.podSecurityContext` | Security context policies to add to the default backend | `{}`
`defaultBackend.service.annotations` | annotations for default backend service | `{}`
`defaultBackend.service.clusterIP` | internal default backend cluster service IP (set to `"-"` to pass an empty value) | `nil`
`defaultBackend.service.omitClusterIP` | (Deprecated) To omit the `clusterIP` from the default backend service | `false`
`defaultBackend.service.externalIPs` | default backend service external IP addresses | `[]`
`defaultBackend.service.loadBalancerIP` | IP address to assign to load balancer (if supported) | `""`
`defaultBackend.service.loadBalancerSourceRanges` | list of IP CIDRs allowed access to load balancer (if supported) | `[]`
`defaultBackend.service.type` | type of default backend service to create | `ClusterIP`
`defaultBackend.serviceAccount.create` | if `true`, create a backend service account. Only useful if you need a pod security policy to run the backend. | `true`
`defaultBackend.serviceAccount.name` | The name of the backend service account to use. If not set and `create` is `true`, a name is generated using the fullname template. Only useful if you need a pod security policy to run the backend. | ``
`imagePullSecrets` | name of Secret resource containing private registry credentials | `nil`
`rbac.create` | if `true`, create & use RBAC resources | `true`
`rbac.scope` | if `true`, do not create & use clusterrole and -binding. Set to `true` in combination with `controller.scope.enabled=true` to disable load-balancer status updates and scope the ingress entirely. | `false`
`podSecurityPolicy.enabled` | if `true`, create & use Pod Security Policy resources | `false`
`serviceAccount.create` | if `true`, create a service account for the controller | `true`
`serviceAccount.name` | The name of the controller service account to use. If not set and `create` is `true`, a name is generated using the fullname template. | ``
`revisionHistoryLimit` | The number of old history to retain to allow rollback. | `10`
`tcp` | TCP service key:value pairs. The value is evaluated as a template. | `{}`
`udp` | UDP service key:value pairs The value is evaluated as a template. | `{}`

These parameters can be passed via Helm's `--set` option
```console
$ helm install ingress-nginx --name my-release \
    --set controller.metrics.enabled=true
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```console
$ helm install ingress-nginx --name my-release -f values.yaml
```

A useful trick to debug issues with ingress is to increase the logLevel
as described [here](https://github.com/kubernetes/ingress-nginx/blob/master/docs/troubleshooting.md#debug)

```console
$ helm install ingress-nginx --set controller.extraArgs.v=2
```
> **Tip**: You can use the default [values.yaml](values.yaml)

## PodDisruptionBudget

Note that the PodDisruptionBudget resource will only be defined if the replicaCount is greater than one,
else it would make it impossible to evacuate a node. See [gh issue #7127](https://github.com/helm/charts/issues/7127) for more info.

## Prometheus Metrics

The Nginx ingress controller can export Prometheus metrics.

```console
$ helm install ingress-nginx --name my-release \
    --set controller.metrics.enabled=true
```

You can add Prometheus annotations to the metrics service using `controller.metrics.service.annotations`. Alternatively, if you use the Prometheus Operator, you can enable ServiceMonitor creation using `controller.metrics.serviceMonitor.enabled`.

## ingress-nginx nginx\_status page/stats server

Previous versions of this chart had a `controller.stats.*` configuration block, which is now obsolete due to the following changes in nginx ingress controller:
* in [0.16.1](https://github.com/kubernetes/ingress-nginx/blob/master/Changelog.md#0161), the vts (virtual host traffic status) dashboard was removed
* in [0.23.0](https://github.com/kubernetes/ingress-nginx/blob/master/Changelog.md#0230), the status page at port 18080 is now a unix socket webserver only available at localhost.
  You can use `curl --unix-socket /tmp/nginx-status-server.sock http://localhost/nginx_status` inside the controller container to access it locally, or use the snippet from [nginx-ingress changelog](https://github.com/kubernetes/ingress-nginx/blob/master/Changelog.md#0230) to re-enable the http server

## ExternalDNS Service configuration

Add an [ExternalDNS](https://github.com/kubernetes-incubator/external-dns) annotation to the LoadBalancer service:

```yaml
controller:
  service:
    annotations:
      external-dns.alpha.kubernetes.io/hostname: kubernetes-example.com.
```

## AWS L7 ELB with SSL Termination

Annotate the controller as shown in the [nginx-ingress l7 patch](https://github.com/kubernetes/ingress-nginx/blob/master/deploy/aws/l7/service-l7.yaml):

```yaml
controller:
  service:
    targetPorts:
      http: http
      https: http
    annotations:
      service.beta.kubernetes.io/aws-load-balancer-ssl-cert: arn:aws:acm:XX-XXXX-X:XXXXXXXXX:certificate/XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXX
      service.beta.kubernetes.io/aws-load-balancer-backend-protocol: "http"
      service.beta.kubernetes.io/aws-load-balancer-ssl-ports: "https"
      service.beta.kubernetes.io/aws-load-balancer-connection-idle-timeout: '3600'
```

## AWS route53-mapper

To configure the LoadBalancer service with the [route53-mapper addon](https://github.com/kubernetes/kops/tree/master/addons/route53-mapper), add the `domainName` annotation and `dns` label:

```yaml
controller:
  service:
    labels:
      dns: "route53"
    annotations:
      domainName: "kubernetes-example.com"
```

## Additional internal load balancer

This setup is useful when you need both external and internal load balancers but don't want to have multiple ingress controllers and multiple ingress objects per application.

By default, the ingress object will point to the external load balancer address, but if correctly configured, you can make use of the internal one if the URL you are looking up resolves to the internal load balancer's URL.

You'll need to set both the following values:

`controller.service.internal.enabled`
`controller.service.internal.annotations`

If one of them is missing the internal load balancer will not be deployed. Example you may have `controller.service.internal.enabled=true` but no annotations set, in this case no action will be taken.

`controller.service.internal.annotations` varies with the cloud service you're using.

Example for AWS
```
controller:
  service:
    internal:
      enabled: true
      annotations:
        # Create internal ELB
        service.beta.kubernetes.io/aws-load-balancer-internal: 0.0.0.0/0
        # Any other annotation can be declared here.
```

Example for GCE
```
controller:
  service:
    internal:
      enabled: true
      annotations:
        # Create internal LB
        cloud.google.com/load-balancer-type: "Internal"
        # Any other annotation can be declared here.
```

An use case for this scenario is having a split-view DNS setup where the public zone CNAME records point to the external balancer URL while the private zone CNAME records point to the internal balancer URL. This way, you only need one ingress kubernetes object.


## Ingress Admission Webhooks

With nginx-ingress-controller version 0.25+, the nginx ingress controller pod exposes an endpoint that will integrate with the `validatingwebhookconfiguration` Kubernetes feature to prevent bad ingress from being added to the cluster.
**This feature is enabled by default since 0.31.0.**

With nginx-ingress-controller in 0.25.* work only with kubernetes 1.14+, 0.26 fix [this issue](https://github.com/kubernetes/ingress-nginx/pull/4521)

## Helm error when upgrading: spec.clusterIP: Invalid value: ""

If you are upgrading this chart from a version between 0.31.0 and 1.2.2 then you may get an error like this:

```
Error: UPGRADE FAILED: Service "?????-controller" is invalid: spec.clusterIP: Invalid value: "": field is immutable
```

Detail of how and why are in [this issue](https://github.com/helm/charts/pull/13646) but to resolve this you can set `xxxx.service.omitClusterIP` to `true` where `xxxx` is the service referenced in the error.

As of version `1.26.0` of this chart, by simply not providing any clusterIP value, `invalid: spec.clusterIP: Invalid value: "": field is immutable` will no longer occur since `clusterIP: ""` will not be rendered.