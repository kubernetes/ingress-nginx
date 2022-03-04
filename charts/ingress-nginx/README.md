# ingress-nginx

[ingress-nginx](https://github.com/kubernetes/ingress-nginx) Ingress controller for Kubernetes using NGINX as a reverse proxy and load balancer

![Version: 4.0.18](https://img.shields.io/badge/Version-4.0.18-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 1.1.2](https://img.shields.io/badge/AppVersion-1.1.2-informational?style=flat-square)

To use, add `ingressClassName: nginx` spec field or the `kubernetes.io/ingress.class: nginx` annotation to your Ingress resources.

This chart bootstraps an ingress-nginx deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites

- Chart version 3.x.x: Kubernetes v1.16+
- Chart version 4.x.x and above: Kubernetes v1.19+

## Get Repo Info

```console
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
```

## Install Chart

**Important:** only helm3 is supported

```console
helm install [RELEASE_NAME] ingress-nginx/ingress-nginx
```

The command deploys ingress-nginx on the Kubernetes cluster in the default configuration.

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] [CHART] --install
```

_See [helm upgrade](https://helm.sh/docs/helm/helm_upgrade/) for command documentation._

### Upgrading With Zero Downtime in Production

By default the ingress-nginx controller has service interruptions whenever it's pods are restarted or redeployed. In order to fix that, see the excellent blog post by Lindsay Landry from Codecademy: [Kubernetes: Nginx and Zero Downtime in Production](https://medium.com/codecademy-engineering/kubernetes-nginx-and-zero-downtime-in-production-2c910c6a5ed8).

### Migrating from stable/nginx-ingress

There are two main ways to migrate a release from `stable/nginx-ingress` to `ingress-nginx/ingress-nginx` chart:

1. For Nginx Ingress controllers used for non-critical services, the easiest method is to [uninstall](#uninstall-chart) the old release and [install](#install-chart) the new one
1. For critical services in production that require zero-downtime, you will want to:
    1. [Install](#install-chart) a second Ingress controller
    1. Redirect your DNS traffic from the old controller to the new controller
    1. Log traffic from both controllers during this changeover
    1. [Uninstall](#uninstall-chart) the old controller once traffic has fully drained from it
    1. For details on all of these steps see [Upgrading With Zero Downtime in Production](#upgrading-with-zero-downtime-in-production)

Note that there are some different and upgraded configurations between the two charts, described by Rimas Mocevicius from JFrog in the "Upgrading to ingress-nginx Helm chart" section of [Migrating from Helm chart nginx-ingress to ingress-nginx](https://rimusz.net/migrating-to-ingress-nginx). As the `ingress-nginx/ingress-nginx` chart continues to update, you will want to check current differences by running [helm configuration](#configuration) commands on both charts.

## Configuration

See [Customizing the Chart Before Installing](https://helm.sh/docs/intro/using_helm/#customizing-the-chart-before-installing). To see all configurable options with detailed comments, visit the chart's [values.yaml](./values.yaml), or run these configuration commands:

```console
helm show values ingress-nginx/ingress-nginx
```

### PodDisruptionBudget

Note that the PodDisruptionBudget resource will only be defined if the replicaCount is greater than one,
else it would make it impossible to evacuate a node. See [gh issue #7127](https://github.com/helm/charts/issues/7127) for more info.

### Prometheus Metrics

The Nginx ingress controller can export Prometheus metrics, by setting `controller.metrics.enabled` to `true`.

You can add Prometheus annotations to the metrics service using `controller.metrics.service.annotations`.
Alternatively, if you use the Prometheus Operator, you can enable ServiceMonitor creation using `controller.metrics.serviceMonitor.enabled`. And set `controller.metrics.serviceMonitor.additionalLabels.release="prometheus"`. "release=prometheus" should match the label configured in the prometheus servicemonitor ( see `kubectl get servicemonitor prometheus-kube-prom-prometheus -oyaml -n prometheus`)

### ingress-nginx nginx\_status page/stats server

Previous versions of this chart had a `controller.stats.*` configuration block, which is now obsolete due to the following changes in nginx ingress controller:

- In [0.16.1](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md#0161), the vts (virtual host traffic status) dashboard was removed
- In [0.23.0](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md#0230), the status page at port 18080 is now a unix socket webserver only available at localhost.
  You can use `curl --unix-socket /tmp/nginx-status-server.sock http://localhost/nginx_status` inside the controller container to access it locally, or use the snippet from [nginx-ingress changelog](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md#0230) to re-enable the http server

### ExternalDNS Service Configuration

Add an [ExternalDNS](https://github.com/kubernetes-incubator/external-dns) annotation to the LoadBalancer service:

```yaml
controller:
  service:
    annotations:
      external-dns.alpha.kubernetes.io/hostname: kubernetes-example.com.
```

### AWS L7 ELB with SSL Termination

Annotate the controller as shown in the [nginx-ingress l7 patch](https://github.com/kubernetes/ingress-nginx/blob/main/deploy/aws/l7/service-l7.yaml):

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

### AWS route53-mapper

To configure the LoadBalancer service with the [route53-mapper addon](https://github.com/kubernetes/kops/tree/master/addons/route53-mapper), add the `domainName` annotation and `dns` label:

```yaml
controller:
  service:
    labels:
      dns: "route53"
    annotations:
      domainName: "kubernetes-example.com"
```

### Additional Internal Load Balancer

This setup is useful when you need both external and internal load balancers but don't want to have multiple ingress controllers and multiple ingress objects per application.

By default, the ingress object will point to the external load balancer address, but if correctly configured, you can make use of the internal one if the URL you are looking up resolves to the internal load balancer's URL.

You'll need to set both the following values:

`controller.service.internal.enabled`
`controller.service.internal.annotations`

If one of them is missing the internal load balancer will not be deployed. Example you may have `controller.service.internal.enabled=true` but no annotations set, in this case no action will be taken.

`controller.service.internal.annotations` varies with the cloud service you're using.

Example for AWS:

```yaml
controller:
  service:
    internal:
      enabled: true
      annotations:
        # Create internal ELB
        service.beta.kubernetes.io/aws-load-balancer-internal: "true"
        # Any other annotation can be declared here.
```

Example for GCE:

```yaml
controller:
  service:
    internal:
      enabled: true
      annotations:
        # Create internal LB. More informations: https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balancing
        # For GKE versions 1.17 and later
        networking.gke.io/load-balancer-type: "Internal"
        # For earlier versions
        # cloud.google.com/load-balancer-type: "Internal"

        # Any other annotation can be declared here.
```

Example for Azure:

```yaml
controller:
  service:
      annotations:
        # Create internal LB
        service.beta.kubernetes.io/azure-load-balancer-internal: "true"
        # Any other annotation can be declared here.
```

Example for Oracle Cloud Infrastructure:

```yaml
controller:
  service:
      annotations:
        # Create internal LB
        service.beta.kubernetes.io/oci-load-balancer-internal: "true"
        # Any other annotation can be declared here.
```

An use case for this scenario is having a split-view DNS setup where the public zone CNAME records point to the external balancer URL while the private zone CNAME records point to the internal balancer URL. This way, you only need one ingress kubernetes object.

Optionally you can set `controller.service.loadBalancerIP` if you need a static IP for the resulting `LoadBalancer`.

### Ingress Admission Webhooks

With nginx-ingress-controller version 0.25+, the nginx ingress controller pod exposes an endpoint that will integrate with the `validatingwebhookconfiguration` Kubernetes feature to prevent bad ingress from being added to the cluster.
**This feature is enabled by default since 0.31.0.**

With nginx-ingress-controller in 0.25.* work only with kubernetes 1.14+, 0.26 fix [this issue](https://github.com/kubernetes/ingress-nginx/pull/4521)

### Helm Error When Upgrading: spec.clusterIP: Invalid value: ""

If you are upgrading this chart from a version between 0.31.0 and 1.2.2 then you may get an error like this:

```console
Error: UPGRADE FAILED: Service "?????-controller" is invalid: spec.clusterIP: Invalid value: "": field is immutable
```

Detail of how and why are in [this issue](https://github.com/helm/charts/pull/13646) but to resolve this you can set `xxxx.service.omitClusterIP` to `true` where `xxxx` is the service referenced in the error.

As of version `1.26.0` of this chart, by simply not providing any clusterIP value, `invalid: spec.clusterIP: Invalid value: "": field is immutable` will no longer occur since `clusterIP: ""` will not be rendered.

## Requirements

Kubernetes: `>=1.19.0-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| commonLabels | object | `{}` |  |
| controller.addHeaders | object | `{}` | Will add custom headers before sending response traffic to the client according to: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#add-headers |
| controller.admissionWebhooks.annotations | object | `{}` |  |
| controller.admissionWebhooks.certificate | string | `"/usr/local/certificates/cert"` |  |
| controller.admissionWebhooks.createSecretJob.resources | object | `{}` |  |
| controller.admissionWebhooks.enabled | bool | `true` |  |
| controller.admissionWebhooks.existingPsp | string | `""` | Use an existing PSP instead of creating one |
| controller.admissionWebhooks.failurePolicy | string | `"Fail"` |  |
| controller.admissionWebhooks.key | string | `"/usr/local/certificates/key"` |  |
| controller.admissionWebhooks.labels | object | `{}` | Labels to be added to admission webhooks |
| controller.admissionWebhooks.namespaceSelector | object | `{}` |  |
| controller.admissionWebhooks.objectSelector | object | `{}` |  |
| controller.admissionWebhooks.patch.enabled | bool | `true` |  |
| controller.admissionWebhooks.patch.fsGroup | int | `2000` |  |
| controller.admissionWebhooks.patch.image.digest | string | `"sha256:64d8c73dca984af206adf9d6d7e46aa550362b1d7a01f3a0a91b20cc67868660"` |  |
| controller.admissionWebhooks.patch.image.image | string | `"ingress-nginx/kube-webhook-certgen"` |  |
| controller.admissionWebhooks.patch.image.pullPolicy | string | `"IfNotPresent"` |  |
| controller.admissionWebhooks.patch.image.registry | string | `"k8s.gcr.io"` |  |
| controller.admissionWebhooks.patch.image.tag | string | `"v1.1.1"` |  |
| controller.admissionWebhooks.patch.labels | object | `{}` | Labels to be added to patch job resources |
| controller.admissionWebhooks.patch.nodeSelector."kubernetes.io/os" | string | `"linux"` |  |
| controller.admissionWebhooks.patch.podAnnotations | object | `{}` |  |
| controller.admissionWebhooks.patch.priorityClassName | string | `""` | Provide a priority class name to the webhook patching job |
| controller.admissionWebhooks.patch.runAsUser | int | `2000` |  |
| controller.admissionWebhooks.patch.tolerations | list | `[]` |  |
| controller.admissionWebhooks.patchWebhookJob.resources | object | `{}` |  |
| controller.admissionWebhooks.port | int | `8443` |  |
| controller.admissionWebhooks.service.annotations | object | `{}` |  |
| controller.admissionWebhooks.service.externalIPs | list | `[]` |  |
| controller.admissionWebhooks.service.loadBalancerSourceRanges | list | `[]` |  |
| controller.admissionWebhooks.service.servicePort | int | `443` |  |
| controller.admissionWebhooks.service.type | string | `"ClusterIP"` |  |
| controller.affinity | object | `{}` | Affinity and anti-affinity rules for server scheduling to nodes |
| controller.allowSnippetAnnotations | bool | `true` | This configuration defines if Ingress Controller should allow users to set their own *-snippet annotations, otherwise this is forbidden / dropped when users add those annotations. Global snippets in ConfigMap are still respected |
| controller.annotations | object | `{}` | Annotations to be added to the controller Deployment or DaemonSet |
| controller.autoscaling.behavior | object | `{}` |  |
| controller.autoscaling.enabled | bool | `false` |  |
| controller.autoscaling.maxReplicas | int | `11` |  |
| controller.autoscaling.minReplicas | int | `1` |  |
| controller.autoscaling.targetCPUUtilizationPercentage | int | `50` |  |
| controller.autoscaling.targetMemoryUtilizationPercentage | int | `50` |  |
| controller.autoscalingTemplate | list | `[]` |  |
| controller.config | object | `{}` | Will add custom configuration options to Nginx https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/ |
| controller.configAnnotations | object | `{}` | Annotations to be added to the controller config configuration configmap. |
| controller.configMapNamespace | string | `""` | Allows customization of the configmap / nginx-configmap namespace; defaults to $(POD_NAMESPACE) |
| controller.containerName | string | `"controller"` | Configures the controller container name |
| controller.containerPort | object | `{"http":80,"https":443}` | Configures the ports that the nginx-controller listens on |
| controller.customTemplate.configMapKey | string | `""` |  |
| controller.customTemplate.configMapName | string | `""` |  |
| controller.dnsConfig | object | `{}` | Optionally customize the pod dnsConfig. |
| controller.dnsPolicy | string | `"ClusterFirst"` | Optionally change this to ClusterFirstWithHostNet in case you have 'hostNetwork: true'. By default, while using host network, name resolution uses the host's DNS. If you wish nginx-controller to keep resolving names inside the k8s network, use ClusterFirstWithHostNet. |
| controller.electionID | string | `"ingress-controller-leader"` | Election ID to use for status update |
| controller.enableMimalloc | bool | `true` | Enable mimalloc as a drop-in replacement for malloc. |
| controller.existingPsp | string | `""` | Use an existing PSP instead of creating one |
| controller.extraArgs | object | `{}` | Additional command line arguments to pass to nginx-ingress-controller E.g. to specify the default SSL certificate you can use |
| controller.extraContainers | list | `[]` | Additional containers to be added to the controller pod. See https://github.com/lemonldap-ng-controller/lemonldap-ng-controller as example. |
| controller.extraEnvs | list | `[]` | Additional environment variables to set |
| controller.extraInitContainers | list | `[]` | Containers, which are run before the app containers are started. |
| controller.extraModules | list | `[]` |  |
| controller.extraVolumeMounts | list | `[]` | Additional volumeMounts to the controller main container. |
| controller.extraVolumes | list | `[]` | Additional volumes to the controller pod. |
| controller.healthCheckHost | string | `""` | Address to bind the health check endpoint. It is better to set this option to the internal node address if the ingress nginx controller is running in the `hostNetwork: true` mode. |
| controller.healthCheckPath | string | `"/healthz"` | Path of the health check endpoint. All requests received on the port defined by the healthz-port parameter are forwarded internally to this path. |
| controller.hostNetwork | bool | `false` | Required for use with CNI based kubernetes installations (such as ones set up by kubeadm), since CNI and hostport don't mix yet. Can be deprecated once https://github.com/kubernetes/kubernetes/issues/23920 is merged |
| controller.hostPort.enabled | bool | `false` | Enable 'hostPort' or not |
| controller.hostPort.ports.http | int | `80` | 'hostPort' http port |
| controller.hostPort.ports.https | int | `443` | 'hostPort' https port |
| controller.hostname | object | `{}` | Optionally customize the pod hostname. |
| controller.image.allowPrivilegeEscalation | bool | `true` |  |
| controller.image.digest | string | `"sha256:28b11ce69e57843de44e3db6413e98d09de0f6688e33d4bd384002a44f78405c"` |  |
| controller.image.image | string | `"ingress-nginx/controller"` |  |
| controller.image.pullPolicy | string | `"IfNotPresent"` |  |
| controller.image.registry | string | `"k8s.gcr.io"` |  |
| controller.image.runAsUser | int | `101` |  |
| controller.image.tag | string | `"v1.1.2"` |  |
| controller.ingressClass | string | `"nginx"` | For backwards compatibility with ingress.class annotation, use ingressClass. Algorithm is as follows, first ingressClassName is considered, if not present, controller looks for ingress.class annotation |
| controller.ingressClassByName | bool | `false` | Process IngressClass per name (additionally as per spec.controller). |
| controller.ingressClassResource.controllerValue | string | `"k8s.io/ingress-nginx"` | Controller-value of the controller that is processing this ingressClass |
| controller.ingressClassResource.default | bool | `false` | Is this the default ingressClass for the cluster |
| controller.ingressClassResource.enabled | bool | `true` | Is this ingressClass enabled or not |
| controller.ingressClassResource.name | string | `"nginx"` | Name of the ingressClass |
| controller.ingressClassResource.parameters | object | `{}` | Parameters is a link to a custom resource containing additional configuration for the controller. This is optional if the controller does not require extra parameters. |
| controller.keda.apiVersion | string | `"keda.sh/v1alpha1"` |  |
| controller.keda.behavior | object | `{}` |  |
| controller.keda.cooldownPeriod | int | `300` |  |
| controller.keda.enabled | bool | `false` |  |
| controller.keda.maxReplicas | int | `11` |  |
| controller.keda.minReplicas | int | `1` |  |
| controller.keda.pollingInterval | int | `30` |  |
| controller.keda.restoreToOriginalReplicaCount | bool | `false` |  |
| controller.keda.scaledObject.annotations | object | `{}` |  |
| controller.keda.triggers | list | `[]` |  |
| controller.kind | string | `"Deployment"` | Use a `DaemonSet` or `Deployment` |
| controller.labels | object | `{}` | Labels to be added to the controller Deployment or DaemonSet and other resources that do not have option to specify labels |
| controller.lifecycle | object | `{"preStop":{"exec":{"command":["/wait-shutdown"]}}}` | Improve connection draining when ingress controller pod is deleted using a lifecycle hook: With this new hook, we increased the default terminationGracePeriodSeconds from 30 seconds to 300, allowing the draining of connections up to five minutes. If the active connections end before that, the pod will terminate gracefully at that time. To effectively take advantage of this feature, the Configmap feature worker-shutdown-timeout new value is 240s instead of 10s. |
| controller.livenessProbe.failureThreshold | int | `5` |  |
| controller.livenessProbe.httpGet.path | string | `"/healthz"` |  |
| controller.livenessProbe.httpGet.port | int | `10254` |  |
| controller.livenessProbe.httpGet.scheme | string | `"HTTP"` |  |
| controller.livenessProbe.initialDelaySeconds | int | `10` |  |
| controller.livenessProbe.periodSeconds | int | `10` |  |
| controller.livenessProbe.successThreshold | int | `1` |  |
| controller.livenessProbe.timeoutSeconds | int | `1` |  |
| controller.maxmindLicenseKey | string | `""` | Maxmind license key to download GeoLite2 Databases. |
| controller.metrics.enabled | bool | `false` |  |
| controller.metrics.port | int | `10254` |  |
| controller.metrics.prometheusRule.additionalLabels | object | `{}` |  |
| controller.metrics.prometheusRule.enabled | bool | `false` |  |
| controller.metrics.prometheusRule.rules | list | `[]` |  |
| controller.metrics.service.annotations | object | `{}` |  |
| controller.metrics.service.externalIPs | list | `[]` | List of IP addresses at which the stats-exporter service is available |
| controller.metrics.service.loadBalancerSourceRanges | list | `[]` |  |
| controller.metrics.service.servicePort | int | `10254` |  |
| controller.metrics.service.type | string | `"ClusterIP"` |  |
| controller.metrics.serviceMonitor.additionalLabels | object | `{}` |  |
| controller.metrics.serviceMonitor.enabled | bool | `false` |  |
| controller.metrics.serviceMonitor.metricRelabelings | list | `[]` |  |
| controller.metrics.serviceMonitor.namespace | string | `""` |  |
| controller.metrics.serviceMonitor.namespaceSelector | object | `{}` |  |
| controller.metrics.serviceMonitor.relabelings | list | `[]` |  |
| controller.metrics.serviceMonitor.scrapeInterval | string | `"30s"` |  |
| controller.metrics.serviceMonitor.targetLabels | list | `[]` |  |
| controller.minAvailable | int | `1` |  |
| controller.minReadySeconds | int | `0` | `minReadySeconds` to avoid killing pods before we are ready |
| controller.name | string | `"controller"` |  |
| controller.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node labels for controller pod assignment |
| controller.podAnnotations | object | `{}` | Annotations to be added to controller pods |
| controller.podLabels | object | `{}` | Labels to add to the pod container metadata |
| controller.podSecurityContext | object | `{}` | Security Context policies for controller pods |
| controller.priorityClassName | string | `""` |  |
| controller.proxySetHeaders | object | `{}` | Will add custom headers before sending traffic to backends according to https://github.com/kubernetes/ingress-nginx/tree/main/docs/examples/customization/custom-headers |
| controller.publishService | object | `{"enabled":true,"pathOverride":""}` | Allows customization of the source of the IP address or FQDN to report in the ingress status field. By default, it reads the information provided by the service. If disable, the status field reports the IP address of the node or nodes where an ingress controller pod is running. |
| controller.publishService.enabled | bool | `true` | Enable 'publishService' or not |
| controller.publishService.pathOverride | string | `""` | Allows overriding of the publish service to bind to Must be <namespace>/<service_name> |
| controller.readinessProbe.failureThreshold | int | `3` |  |
| controller.readinessProbe.httpGet.path | string | `"/healthz"` |  |
| controller.readinessProbe.httpGet.port | int | `10254` |  |
| controller.readinessProbe.httpGet.scheme | string | `"HTTP"` |  |
| controller.readinessProbe.initialDelaySeconds | int | `10` |  |
| controller.readinessProbe.periodSeconds | int | `10` |  |
| controller.readinessProbe.successThreshold | int | `1` |  |
| controller.readinessProbe.timeoutSeconds | int | `1` |  |
| controller.replicaCount | int | `1` |  |
| controller.reportNodeInternalIp | bool | `false` | Bare-metal considerations via the host network https://kubernetes.github.io/ingress-nginx/deploy/baremetal/#via-the-host-network Ingress status was blank because there is no Service exposing the NGINX Ingress controller in a configuration using the host network, the default --publish-service flag used in standard cloud setups does not apply |
| controller.resources.requests.cpu | string | `"100m"` |  |
| controller.resources.requests.memory | string | `"90Mi"` |  |
| controller.scope.enabled | bool | `false` | Enable 'scope' or not |
| controller.scope.namespace | string | `""` | Namespace to limit the controller to; defaults to $(POD_NAMESPACE) |
| controller.scope.namespaceSelector | string | `""` | When scope.enabled == false, instead of watching all namespaces, we watching namespaces whose labels only match with namespaceSelector. Format like foo=bar. Defaults to empty, means watching all namespaces. |
| controller.service.annotations | object | `{}` |  |
| controller.service.appProtocol | bool | `true` | If enabled is adding an appProtocol option for Kubernetes service. An appProtocol field replacing annotations that were using for setting a backend protocol. Here is an example for AWS: service.beta.kubernetes.io/aws-load-balancer-backend-protocol: http It allows choosing the protocol for each backend specified in the Kubernetes service. See the following GitHub issue for more details about the purpose: https://github.com/kubernetes/kubernetes/issues/40244 Will be ignored for Kubernetes versions older than 1.20 |
| controller.service.enableHttp | bool | `true` |  |
| controller.service.enableHttps | bool | `true` |  |
| controller.service.enabled | bool | `true` |  |
| controller.service.external.enabled | bool | `true` |  |
| controller.service.externalIPs | list | `[]` | List of IP addresses at which the controller services are available |
| controller.service.internal.annotations | object | `{}` | Annotations are mandatory for the load balancer to come up. Varies with the cloud service. |
| controller.service.internal.enabled | bool | `false` | Enables an additional internal load balancer (besides the external one). |
| controller.service.internal.loadBalancerSourceRanges | list | `[]` | Restrict access For LoadBalancer service. Defaults to 0.0.0.0/0. |
| controller.service.ipFamilies | list | `["IPv4"]` | List of IP families (e.g. IPv4, IPv6) assigned to the service. This field is usually assigned automatically based on cluster configuration and the ipFamilyPolicy field. |
| controller.service.ipFamilyPolicy | string | `"SingleStack"` | Represents the dual-stack-ness requested or required by this Service. Possible values are SingleStack, PreferDualStack or RequireDualStack. The ipFamilies and clusterIPs fields depend on the value of this field. |
| controller.service.labels | object | `{}` |  |
| controller.service.loadBalancerSourceRanges | list | `[]` |  |
| controller.service.nodePorts.http | string | `""` |  |
| controller.service.nodePorts.https | string | `""` |  |
| controller.service.nodePorts.tcp | object | `{}` |  |
| controller.service.nodePorts.udp | object | `{}` |  |
| controller.service.ports.http | int | `80` |  |
| controller.service.ports.https | int | `443` |  |
| controller.service.targetPorts.http | string | `"http"` |  |
| controller.service.targetPorts.https | string | `"https"` |  |
| controller.service.type | string | `"LoadBalancer"` |  |
| controller.sysctls | object | `{}` | See https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/ for notes on enabling and using sysctls |
| controller.tcp.annotations | object | `{}` | Annotations to be added to the tcp config configmap |
| controller.tcp.configMapNamespace | string | `""` | Allows customization of the tcp-services-configmap; defaults to $(POD_NAMESPACE) |
| controller.terminationGracePeriodSeconds | int | `300` | `terminationGracePeriodSeconds` to avoid killing pods before we are ready |
| controller.tolerations | list | `[]` | Node tolerations for server scheduling to nodes with taints |
| controller.topologySpreadConstraints | list | `[]` | Topology spread constraints rely on node labels to identify the topology domain(s) that each Node is in. |
| controller.udp.annotations | object | `{}` | Annotations to be added to the udp config configmap |
| controller.udp.configMapNamespace | string | `""` | Allows customization of the udp-services-configmap; defaults to $(POD_NAMESPACE) |
| controller.updateStrategy | object | `{}` | The update strategy to apply to the Deployment or DaemonSet |
| controller.watchIngressWithoutClass | bool | `false` | Process Ingress objects without ingressClass annotation/ingressClassName field Overrides value for --watch-ingress-without-class flag of the controller binary Defaults to false |
| defaultBackend.affinity | object | `{}` |  |
| defaultBackend.autoscaling.annotations | object | `{}` |  |
| defaultBackend.autoscaling.enabled | bool | `false` |  |
| defaultBackend.autoscaling.maxReplicas | int | `2` |  |
| defaultBackend.autoscaling.minReplicas | int | `1` |  |
| defaultBackend.autoscaling.targetCPUUtilizationPercentage | int | `50` |  |
| defaultBackend.autoscaling.targetMemoryUtilizationPercentage | int | `50` |  |
| defaultBackend.containerSecurityContext | object | `{}` | Security Context policies for controller main container. See https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/ for notes on enabling and using sysctls |
| defaultBackend.enabled | bool | `false` |  |
| defaultBackend.existingPsp | string | `""` | Use an existing PSP instead of creating one |
| defaultBackend.extraArgs | object | `{}` |  |
| defaultBackend.extraEnvs | list | `[]` | Additional environment variables to set for defaultBackend pods |
| defaultBackend.extraVolumeMounts | list | `[]` |  |
| defaultBackend.extraVolumes | list | `[]` |  |
| defaultBackend.image.allowPrivilegeEscalation | bool | `false` |  |
| defaultBackend.image.image | string | `"defaultbackend-amd64"` |  |
| defaultBackend.image.pullPolicy | string | `"IfNotPresent"` |  |
| defaultBackend.image.readOnlyRootFilesystem | bool | `true` |  |
| defaultBackend.image.registry | string | `"k8s.gcr.io"` |  |
| defaultBackend.image.runAsNonRoot | bool | `true` |  |
| defaultBackend.image.runAsUser | int | `65534` |  |
| defaultBackend.image.tag | string | `"1.5"` |  |
| defaultBackend.labels | object | `{}` | Labels to be added to the default backend resources |
| defaultBackend.livenessProbe.failureThreshold | int | `3` |  |
| defaultBackend.livenessProbe.initialDelaySeconds | int | `30` |  |
| defaultBackend.livenessProbe.periodSeconds | int | `10` |  |
| defaultBackend.livenessProbe.successThreshold | int | `1` |  |
| defaultBackend.livenessProbe.timeoutSeconds | int | `5` |  |
| defaultBackend.minAvailable | int | `1` |  |
| defaultBackend.name | string | `"defaultbackend"` |  |
| defaultBackend.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node labels for default backend pod assignment |
| defaultBackend.podAnnotations | object | `{}` | Annotations to be added to default backend pods |
| defaultBackend.podLabels | object | `{}` | Labels to add to the pod container metadata |
| defaultBackend.podSecurityContext | object | `{}` | Security Context policies for controller pods See https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/ for notes on enabling and using sysctls |
| defaultBackend.port | int | `8080` |  |
| defaultBackend.priorityClassName | string | `""` |  |
| defaultBackend.readinessProbe.failureThreshold | int | `6` |  |
| defaultBackend.readinessProbe.initialDelaySeconds | int | `0` |  |
| defaultBackend.readinessProbe.periodSeconds | int | `5` |  |
| defaultBackend.readinessProbe.successThreshold | int | `1` |  |
| defaultBackend.readinessProbe.timeoutSeconds | int | `5` |  |
| defaultBackend.replicaCount | int | `1` |  |
| defaultBackend.resources | object | `{}` |  |
| defaultBackend.service.annotations | object | `{}` |  |
| defaultBackend.service.externalIPs | list | `[]` | List of IP addresses at which the default backend service is available |
| defaultBackend.service.loadBalancerSourceRanges | list | `[]` |  |
| defaultBackend.service.servicePort | int | `80` |  |
| defaultBackend.service.type | string | `"ClusterIP"` |  |
| defaultBackend.serviceAccount.automountServiceAccountToken | bool | `true` |  |
| defaultBackend.serviceAccount.create | bool | `true` |  |
| defaultBackend.serviceAccount.name | string | `""` |  |
| defaultBackend.tolerations | list | `[]` | Node tolerations for server scheduling to nodes with taints |
| dhParam | string | `nil` | A base64-encoded Diffie-Hellman parameter. This can be generated with: `openssl dhparam 4096 2> /dev/null | base64` |
| imagePullSecrets | list | `[]` | Optional array of imagePullSecrets containing private registry credentials |
| podSecurityPolicy.enabled | bool | `false` |  |
| rbac.create | bool | `true` |  |
| rbac.scope | bool | `false` |  |
| revisionHistoryLimit | int | `10` | Rollback limit |
| serviceAccount.annotations | object | `{}` | Annotations for the controller service account |
| serviceAccount.automountServiceAccountToken | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tcp | object | `{}` | TCP service key:value pairs |
| udp | object | `{}` | UDP service key:value pairs |

