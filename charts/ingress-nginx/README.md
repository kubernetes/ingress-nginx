# ingress-nginx

[ingress-nginx](https://github.com/kubernetes/ingress-nginx) Ingress controller for Kubernetes using NGINX as a reverse proxy and load balancer

![Version: 4.12.3](https://img.shields.io/badge/Version-4.12.3-informational?style=flat-square) ![AppVersion: 1.12.3](https://img.shields.io/badge/AppVersion-1.12.3-informational?style=flat-square)

To use, add `ingressClassName: nginx` spec field or the `kubernetes.io/ingress.class: nginx` annotation to your Ingress resources.

This chart bootstraps an ingress-nginx deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Requirements

Kubernetes: `>=1.21.0-0`

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

### Migrating from stable/nginx-ingress

There are two main ways to migrate a release from `stable/nginx-ingress` to `ingress-nginx/ingress-nginx` chart:

1. For Nginx Ingress controllers used for non-critical services, the easiest method is to [uninstall](#uninstall-chart) the old release and [install](#install-chart) the new one
1. For critical services in production that require zero-downtime, you will want to:
    1. [Install](#install-chart) a second Ingress controller
    1. Redirect your DNS traffic from the old controller to the new controller
    1. Log traffic from both controllers during this changeover
    1. [Uninstall](#uninstall-chart) the old controller once traffic has fully drained from it

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

The Ingress-Nginx Controller can export Prometheus metrics, by setting `controller.metrics.enabled` to `true`.

You can add Prometheus annotations to the metrics service using `controller.metrics.service.annotations`.
Alternatively, if you use the Prometheus Operator, you can enable ServiceMonitor creation using `controller.metrics.serviceMonitor.enabled`. And set `controller.metrics.serviceMonitor.additionalLabels.release="prometheus"`. "release=prometheus" should match the label configured in the prometheus servicemonitor ( see `kubectl get servicemonitor prometheus-kube-prom-prometheus -oyaml -n prometheus`)

### ingress-nginx nginx\_status page/stats server

Previous versions of this chart had a `controller.stats.*` configuration block, which is now obsolete due to the following changes in Ingress-Nginx Controller:

- In [0.16.1](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md#0161), the vts (virtual host traffic status) dashboard was removed
- In [0.23.0](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md#0230), the status page at port 18080 is now a unix socket webserver only available at localhost.
  You can use `curl --unix-socket /tmp/nginx-status-server.sock http://localhost/nginx_status` inside the controller container to access it locally, or use the snippet from [nginx-ingress changelog](https://github.com/kubernetes/ingress-nginx/blob/main/Changelog.md#0230) to re-enable the http server

### ExternalDNS Service Configuration

Add an [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) annotation to the LoadBalancer service:

```yaml
controller:
  service:
    annotations:
      external-dns.alpha.kubernetes.io/hostname: kubernetes-example.com.
```

### AWS L7 ELB with SSL Termination

Annotate the controller as shown in the [nginx-ingress l7 patch](https://github.com/kubernetes/ingress-nginx/blob/ab3a789caae65eec4ad6e3b46b19750b481b6bce/deploy/aws/l7/service-l7.yaml):

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
        # Create internal NLB
        service.beta.kubernetes.io/aws-load-balancer-scheme: "internal"
        # Create internal ELB(Deprecated)
        # service.beta.kubernetes.io/aws-load-balancer-internal: "true"
        # Any other annotation can be declared here.
```

Example for GCE:

```yaml
controller:
  service:
    internal:
      enabled: true
      annotations:
        # Create internal LB. More information: https://cloud.google.com/kubernetes-engine/docs/how-to/internal-load-balancing
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

The load balancer annotations of more cloud service providers can be found: [Internal load balancer](https://kubernetes.io/docs/concepts/services-networking/service/#internal-load-balancer).

An use case for this scenario is having a split-view DNS setup where the public zone CNAME records point to the external balancer URL while the private zone CNAME records point to the internal balancer URL. This way, you only need one ingress kubernetes object.

Optionally you can set `controller.service.loadBalancerIP` if you need a static IP for the resulting `LoadBalancer`.

### Ingress Admission Webhooks

With nginx-ingress-controller version 0.25+, the Ingress-Nginx Controller pod exposes an endpoint that will integrate with the `validatingwebhookconfiguration` Kubernetes feature to prevent bad ingress from being added to the cluster.
**This feature is enabled by default since 0.31.0.**

With nginx-ingress-controller in 0.25.* work only with kubernetes 1.14+, 0.26 fix [this issue](https://github.com/kubernetes/ingress-nginx/pull/4521)

#### How the Chart Configures the Hooks
A validating and configuration requires the endpoint to which the request is sent to use TLS. It is possible to set up custom certificates to do this, but in most cases, a self-signed certificate is enough. The setup of this component requires some more complex orchestration when using helm. The steps are created to be idempotent and to allow turning the feature on and off without running into helm quirks.

1. A pre-install hook provisions a certificate into the same namespace using a format compatible with provisioning using end user certificates. If the certificate already exists, the hook exits.
2. The Ingress-Nginx Controller pod is configured to use a TLS proxy container, which will load that certificate.
3. Validating and Mutating webhook configurations are created in the cluster.
4. A post-install hook reads the CA from the secret created by step 1 and patches the Validating and Mutating webhook configurations. This process will allow a custom CA provisioned by some other process to also be patched into the webhook configurations. The chosen failure policy is also patched into the webhook configurations

#### Alternatives
It should be possible to use [cert-manager/cert-manager](https://github.com/cert-manager/cert-manager) if a more complete solution is required.

You can enable automatic self-signed TLS certificate provisioning via cert-manager by setting the `controller.admissionWebhooks.certManager.enabled` value to true.

Please ensure that cert-manager is correctly installed and configured.

### Helm Error When Upgrading: spec.clusterIP: Invalid value: ""

If you are upgrading this chart from a version between 0.31.0 and 1.2.2 then you may get an error like this:

```console
Error: UPGRADE FAILED: Service "?????-controller" is invalid: spec.clusterIP: Invalid value: "": field is immutable
```

Detail of how and why are in [this issue](https://github.com/helm/charts/pull/13646) but to resolve this you can set `xxxx.service.omitClusterIP` to `true` where `xxxx` is the service referenced in the error.

As of version `1.26.0` of this chart, by simply not providing any clusterIP value, `invalid: spec.clusterIP: Invalid value: "": field is immutable` will no longer occur since `clusterIP: ""` will not be rendered.

### Pod Security Admission

You can use Pod Security Admission by applying labels to the `ingress-nginx` namespace as instructed by the [documentation](https://kubernetes.io/docs/tasks/configure-pod-container/enforce-standards-namespace-labels).

Example:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: ingress-nginx
  labels:
    kubernetes.io/metadata.name: ingress-nginx
    name: ingress-nginx
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: v1.31
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| commonLabels | object | `{}` |  |
| controller.addHeaders | object | `{}` | Will add custom headers before sending response traffic to the client according to: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#add-headers |
| controller.admissionWebhooks.annotations | object | `{}` |  |
| controller.admissionWebhooks.certManager.admissionCert.duration | string | `""` |  |
| controller.admissionWebhooks.certManager.enabled | bool | `false` |  |
| controller.admissionWebhooks.certManager.rootCert.duration | string | `""` |  |
| controller.admissionWebhooks.certificate | string | `"/usr/local/certificates/cert"` |  |
| controller.admissionWebhooks.createSecretJob.name | string | `"create"` |  |
| controller.admissionWebhooks.createSecretJob.resources | object | `{}` |  |
| controller.admissionWebhooks.createSecretJob.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532,"seccompProfile":{"type":"RuntimeDefault"}}` | Security context for secret creation containers |
| controller.admissionWebhooks.enabled | bool | `true` |  |
| controller.admissionWebhooks.extraEnvs | list | `[]` | Additional environment variables to set |
| controller.admissionWebhooks.failurePolicy | string | `"Fail"` | Admission Webhook failure policy to use |
| controller.admissionWebhooks.key | string | `"/usr/local/certificates/key"` |  |
| controller.admissionWebhooks.labels | object | `{}` | Labels to be added to admission webhooks |
| controller.admissionWebhooks.name | string | `"admission"` |  |
| controller.admissionWebhooks.namespaceSelector | object | `{}` |  |
| controller.admissionWebhooks.objectSelector | object | `{}` |  |
| controller.admissionWebhooks.patch.enabled | bool | `true` |  |
| controller.admissionWebhooks.patch.image.digest | string | `"sha256:7a38cf0f8480775baaee71ab519c7465fd1dfeac66c421f28f087786e631456e"` |  |
| controller.admissionWebhooks.patch.image.image | string | `"ingress-nginx/kube-webhook-certgen"` |  |
| controller.admissionWebhooks.patch.image.pullPolicy | string | `"IfNotPresent"` |  |
| controller.admissionWebhooks.patch.image.tag | string | `"v1.5.4"` |  |
| controller.admissionWebhooks.patch.labels | object | `{}` | Labels to be added to patch job resources |
| controller.admissionWebhooks.patch.networkPolicy.enabled | bool | `false` | Enable 'networkPolicy' or not |
| controller.admissionWebhooks.patch.nodeSelector."kubernetes.io/os" | string | `"linux"` |  |
| controller.admissionWebhooks.patch.podAnnotations | object | `{}` |  |
| controller.admissionWebhooks.patch.priorityClassName | string | `""` | Provide a priority class name to the webhook patching job # |
| controller.admissionWebhooks.patch.rbac | object | `{"create":true}` | Admission webhook patch job RBAC |
| controller.admissionWebhooks.patch.rbac.create | bool | `true` | Create RBAC or not |
| controller.admissionWebhooks.patch.securityContext | object | `{}` | Security context for secret creation & webhook patch pods |
| controller.admissionWebhooks.patch.serviceAccount | object | `{"automountServiceAccountToken":true,"create":true,"name":""}` | Admission webhook patch job service account |
| controller.admissionWebhooks.patch.serviceAccount.automountServiceAccountToken | bool | `true` | Auto-mount service account token or not |
| controller.admissionWebhooks.patch.serviceAccount.create | bool | `true` | Create a service account or not |
| controller.admissionWebhooks.patch.serviceAccount.name | string | `""` | Custom service account name |
| controller.admissionWebhooks.patch.tolerations | list | `[]` |  |
| controller.admissionWebhooks.patchWebhookJob.name | string | `"patch"` |  |
| controller.admissionWebhooks.patchWebhookJob.resources | object | `{}` |  |
| controller.admissionWebhooks.patchWebhookJob.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"readOnlyRootFilesystem":true,"runAsGroup":65532,"runAsNonRoot":true,"runAsUser":65532,"seccompProfile":{"type":"RuntimeDefault"}}` | Security context for webhook patch containers |
| controller.admissionWebhooks.port | int | `8443` |  |
| controller.admissionWebhooks.service.annotations | object | `{}` |  |
| controller.admissionWebhooks.service.externalIPs | list | `[]` |  |
| controller.admissionWebhooks.service.loadBalancerSourceRanges | list | `[]` |  |
| controller.admissionWebhooks.service.servicePort | int | `443` |  |
| controller.admissionWebhooks.service.type | string | `"ClusterIP"` |  |
| controller.affinity | object | `{}` | Affinity and anti-affinity rules for server scheduling to nodes # Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity # |
| controller.allowSnippetAnnotations | bool | `false` | This configuration defines if Ingress Controller should allow users to set their own *-snippet annotations, otherwise this is forbidden / dropped when users add those annotations. Global snippets in ConfigMap are still respected |
| controller.annotations | object | `{}` | Annotations to be added to the controller Deployment or DaemonSet # |
| controller.autoscaling.annotations | object | `{}` |  |
| controller.autoscaling.behavior | object | `{}` |  |
| controller.autoscaling.enabled | bool | `false` |  |
| controller.autoscaling.maxReplicas | int | `11` |  |
| controller.autoscaling.minReplicas | int | `1` |  |
| controller.autoscaling.targetCPUUtilizationPercentage | int | `50` |  |
| controller.autoscaling.targetMemoryUtilizationPercentage | int | `50` |  |
| controller.autoscalingTemplate | list | `[]` |  |
| controller.config | object | `{}` | Global configuration passed to the ConfigMap consumed by the controller. Values may contain Helm templates. Ref.: https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/ |
| controller.configAnnotations | object | `{}` | Annotations to be added to the controller config configuration configmap. |
| controller.configMapNamespace | string | `""` | Allows customization of the configmap / nginx-configmap namespace; defaults to $(POD_NAMESPACE) |
| controller.containerName | string | `"controller"` | Configures the controller container name |
| controller.containerPort | object | `{"http":80,"https":443}` | Configures the ports that the nginx-controller listens on |
| controller.containerSecurityContext | object | `{}` | Security context for controller containers |
| controller.customTemplate.configMapKey | string | `""` |  |
| controller.customTemplate.configMapName | string | `""` |  |
| controller.disableLeaderElection | bool | `false` | This configuration disable Nginx Controller Leader Election |
| controller.dnsConfig | object | `{}` | Optionally customize the pod dnsConfig. |
| controller.dnsPolicy | string | `"ClusterFirst"` | Optionally change this to ClusterFirstWithHostNet in case you have 'hostNetwork: true'. By default, while using host network, name resolution uses the host's DNS. If you wish nginx-controller to keep resolving names inside the k8s network, use ClusterFirstWithHostNet. |
| controller.electionID | string | `""` | Election ID to use for status update, by default it uses the controller name combined with a suffix of 'leader' |
| controller.electionTTL | string | `""` | Duration a leader election is valid before it's getting re-elected, e.g. `15s`, `10m` or `1h`. (Default: 30s) |
| controller.enableAnnotationValidations | bool | `true` |  |
| controller.enableMimalloc | bool | `true` | Enable mimalloc as a drop-in replacement for malloc. # ref: https://github.com/microsoft/mimalloc # |
| controller.enableTopologyAwareRouting | bool | `false` | This configuration enables Topology Aware Routing feature, used together with service annotation service.kubernetes.io/topology-mode="auto" Defaults to false |
| controller.extraArgs | object | `{}` | Additional command line arguments to pass to Ingress-Nginx Controller E.g. to specify the default SSL certificate you can use |
| controller.extraContainers | list | `[]` | Additional containers to be added to the controller pod. See https://github.com/lemonldap-ng-controller/lemonldap-ng-controller as example. |
| controller.extraEnvs | list | `[]` | Additional environment variables to set |
| controller.extraInitContainers | list | `[]` | Containers, which are run before the app containers are started. |
| controller.extraModules | list | `[]` | Modules, which are mounted into the core nginx image. |
| controller.extraVolumeMounts | list | `[]` | Additional volumeMounts to the controller main container. |
| controller.extraVolumes | list | `[]` | Additional volumes to the controller pod. |
| controller.healthCheckHost | string | `""` | Address to bind the health check endpoint. It is better to set this option to the internal node address if the Ingress-Nginx Controller is running in the `hostNetwork: true` mode. |
| controller.healthCheckPath | string | `"/healthz"` | Path of the health check endpoint. All requests received on the port defined by the healthz-port parameter are forwarded internally to this path. |
| controller.hostAliases | list | `[]` | Optionally customize the pod hostAliases. |
| controller.hostNetwork | bool | `false` | Required for use with CNI based kubernetes installations (such as ones set up by kubeadm), since CNI and hostport don't mix yet. Can be deprecated once https://github.com/kubernetes/kubernetes/issues/23920 is merged |
| controller.hostPort.enabled | bool | `false` | Enable 'hostPort' or not |
| controller.hostPort.ports.http | int | `80` | 'hostPort' http port |
| controller.hostPort.ports.https | int | `443` | 'hostPort' https port |
| controller.hostname | object | `{}` | Optionally customize the pod hostname. |
| controller.image.allowPrivilegeEscalation | bool | `false` |  |
| controller.image.chroot | bool | `false` |  |
| controller.image.digest | string | `"sha256:ac444cd9515af325ba577b596fe4f27a34be1aa330538e8b317ad9d6c8fb94ee"` |  |
| controller.image.digestChroot | string | `"sha256:d830fba93e9e0f5ef1462f5fe8a7cd7b167178b79e6c10c041c7da19f1ac66ab"` |  |
| controller.image.image | string | `"ingress-nginx/controller"` |  |
| controller.image.pullPolicy | string | `"IfNotPresent"` |  |
| controller.image.readOnlyRootFilesystem | bool | `false` |  |
| controller.image.runAsGroup | int | `82` | This value must not be changed using the official image. uid=101(www-data) gid=82(www-data) groups=82(www-data) |
| controller.image.runAsNonRoot | bool | `true` |  |
| controller.image.runAsUser | int | `101` | This value must not be changed using the official image. uid=101(www-data) gid=82(www-data) groups=82(www-data) |
| controller.image.seccompProfile.type | string | `"RuntimeDefault"` |  |
| controller.image.tag | string | `"v1.12.3"` |  |
| controller.ingressClass | string | `"nginx"` | For backwards compatibility with ingress.class annotation, use ingressClass. Algorithm is as follows, first ingressClassName is considered, if not present, controller looks for ingress.class annotation |
| controller.ingressClassByName | bool | `false` | Process IngressClass per name (additionally as per spec.controller). |
| controller.ingressClassResource | object | `{"aliases":[],"annotations":{},"controllerValue":"k8s.io/ingress-nginx","default":false,"enabled":true,"name":"nginx","parameters":{}}` | This section refers to the creation of the IngressClass resource. IngressClasses are immutable and cannot be changed after creation. We do not support namespaced IngressClasses, yet, so a ClusterRole and a ClusterRoleBinding is required. |
| controller.ingressClassResource.aliases | list | `[]` | Aliases of this IngressClass. Creates copies with identical settings but the respective alias as name. Useful for development environments with only one Ingress Controller but production-like Ingress resources. `default` gets enabled on the original IngressClass only. |
| controller.ingressClassResource.annotations | object | `{}` | Annotations to be added to the IngressClass resource. |
| controller.ingressClassResource.controllerValue | string | `"k8s.io/ingress-nginx"` | Controller of the IngressClass. An Ingress Controller looks for IngressClasses it should reconcile by this value. This value is also being set as the `--controller-class` argument of this Ingress Controller. Ref: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class |
| controller.ingressClassResource.default | bool | `false` | If true, Ingresses without `ingressClassName` get assigned to this IngressClass on creation. Ingress creation gets rejected if there are multiple default IngressClasses. Ref: https://kubernetes.io/docs/concepts/services-networking/ingress/#default-ingress-class |
| controller.ingressClassResource.enabled | bool | `true` | Create the IngressClass or not |
| controller.ingressClassResource.name | string | `"nginx"` | Name of the IngressClass |
| controller.ingressClassResource.parameters | object | `{}` | A link to a custom resource containing additional configuration for the controller. This is optional if the controller consuming this IngressClass does not require additional parameters. Ref: https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class |
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
| controller.labels | object | `{}` | Labels to be added to the controller Deployment or DaemonSet and other resources that do not have option to specify labels # |
| controller.lifecycle | object | `{"preStop":{"exec":{"command":["/wait-shutdown"]}}}` | Improve connection draining when ingress controller pod is deleted using a lifecycle hook: With this new hook, we increased the default terminationGracePeriodSeconds from 30 seconds to 300, allowing the draining of connections up to five minutes. If the active connections end before that, the pod will terminate gracefully at that time. To effectively take advantage of this feature, the Configmap feature worker-shutdown-timeout new value is 240s instead of 10s. # |
| controller.livenessProbe.failureThreshold | int | `5` |  |
| controller.livenessProbe.httpGet.path | string | `"/healthz"` |  |
| controller.livenessProbe.httpGet.port | int | `10254` |  |
| controller.livenessProbe.httpGet.scheme | string | `"HTTP"` |  |
| controller.livenessProbe.initialDelaySeconds | int | `10` |  |
| controller.livenessProbe.periodSeconds | int | `10` |  |
| controller.livenessProbe.successThreshold | int | `1` |  |
| controller.livenessProbe.timeoutSeconds | int | `1` |  |
| controller.maxmindLicenseKey | string | `""` | Maxmind license key to download GeoLite2 Databases. # https://blog.maxmind.com/2019/12/significant-changes-to-accessing-and-using-geolite2-databases/ |
| controller.metrics.enabled | bool | `false` |  |
| controller.metrics.port | int | `10254` |  |
| controller.metrics.portName | string | `"metrics"` |  |
| controller.metrics.prometheusRule.additionalLabels | object | `{}` |  |
| controller.metrics.prometheusRule.annotations | object | `{}` | Annotations to be added to the PrometheusRule. |
| controller.metrics.prometheusRule.enabled | bool | `false` |  |
| controller.metrics.prometheusRule.rules | list | `[]` |  |
| controller.metrics.service.annotations | object | `{}` |  |
| controller.metrics.service.enabled | bool | `true` | Enable the metrics service or not. |
| controller.metrics.service.externalIPs | list | `[]` | List of IP addresses at which the stats-exporter service is available # Ref: https://kubernetes.io/docs/concepts/services-networking/service/#external-ips # |
| controller.metrics.service.labels | object | `{}` | Labels to be added to the metrics service resource |
| controller.metrics.service.loadBalancerSourceRanges | list | `[]` |  |
| controller.metrics.service.servicePort | int | `10254` |  |
| controller.metrics.service.type | string | `"ClusterIP"` |  |
| controller.metrics.serviceMonitor.additionalLabels | object | `{}` |  |
| controller.metrics.serviceMonitor.annotations | object | `{}` | Annotations to be added to the ServiceMonitor. |
| controller.metrics.serviceMonitor.enabled | bool | `false` |  |
| controller.metrics.serviceMonitor.metricRelabelings | list | `[]` |  |
| controller.metrics.serviceMonitor.namespace | string | `""` |  |
| controller.metrics.serviceMonitor.namespaceSelector | object | `{}` |  |
| controller.metrics.serviceMonitor.relabelings | list | `[]` |  |
| controller.metrics.serviceMonitor.scrapeInterval | string | `"30s"` |  |
| controller.metrics.serviceMonitor.targetLabels | list | `[]` |  |
| controller.minAvailable | int | `1` | Minimum available pods set in PodDisruptionBudget. Define either 'minAvailable' or 'maxUnavailable', never both. |
| controller.minReadySeconds | int | `0` | `minReadySeconds` to avoid killing pods before we are ready # |
| controller.name | string | `"controller"` |  |
| controller.networkPolicy.enabled | bool | `false` | Enable 'networkPolicy' or not |
| controller.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node labels for controller pod assignment # Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/ # |
| controller.podAnnotations | object | `{}` | Annotations to be added to controller pods # |
| controller.podLabels | object | `{}` | Labels to add to the pod container metadata |
| controller.podSecurityContext | object | `{}` | Security context for controller pods |
| controller.priorityClassName | string | `""` |  |
| controller.progressDeadlineSeconds | int | `0` | Specifies the number of seconds you want to wait for the controller deployment to progress before the system reports back that it has failed. Ref.: https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#progress-deadline-seconds |
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
| controller.reportNodeInternalIp | bool | `false` | Bare-metal considerations via the host network https://kubernetes.github.io/ingress-nginx/deploy/baremetal/#via-the-host-network Ingress status was blank because there is no Service exposing the Ingress-Nginx Controller in a configuration using the host network, the default --publish-service flag used in standard cloud setups does not apply |
| controller.resources.requests.cpu | string | `"100m"` |  |
| controller.resources.requests.memory | string | `"90Mi"` |  |
| controller.scope.enabled | bool | `false` | Enable 'scope' or not |
| controller.scope.namespace | string | `""` | Namespace to limit the controller to; defaults to $(POD_NAMESPACE) |
| controller.scope.namespaceSelector | string | `""` | When scope.enabled == false, instead of watching all namespaces, we watching namespaces whose labels only match with namespaceSelector. Format like foo=bar. Defaults to empty, means watching all namespaces. |
| controller.service.annotations | object | `{}` | Annotations to be added to the external controller service. See `controller.service.internal.annotations` for annotations to be added to the internal controller service. |
| controller.service.appProtocol | bool | `true` | Declare the app protocol of the external HTTP and HTTPS listeners or not. Supersedes provider-specific annotations for declaring the backend protocol. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#application-protocol |
| controller.service.clusterIP | string | `""` | Pre-defined cluster internal IP address of the external controller service. Take care of collisions with existing services. This value is immutable. Set once, it can not be changed without deleting and re-creating the service. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#choosing-your-own-ip-address |
| controller.service.enableHttp | bool | `true` | Enable the HTTP listener on both controller services or not. |
| controller.service.enableHttps | bool | `true` | Enable the HTTPS listener on both controller services or not. |
| controller.service.enabled | bool | `true` | Enable controller services or not. This does not influence the creation of either the admission webhook or the metrics service. |
| controller.service.external.enabled | bool | `true` | Enable the external controller service or not. Useful for internal-only deployments. |
| controller.service.externalIPs | list | `[]` | List of node IP addresses at which the external controller service is available. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#external-ips |
| controller.service.externalTrafficPolicy | string | `""` | External traffic policy of the external controller service. Set to "Local" to preserve source IP on providers supporting it. Ref: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip |
| controller.service.internal.annotations | object | `{}` | Annotations to be added to the internal controller service. Mandatory for the internal controller service to be created. Varies with the cloud service. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#internal-load-balancer |
| controller.service.internal.appProtocol | bool | `true` | Declare the app protocol of the internal HTTP and HTTPS listeners or not. Supersedes provider-specific annotations for declaring the backend protocol. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#application-protocol |
| controller.service.internal.clusterIP | string | `""` | Pre-defined cluster internal IP address of the internal controller service. Take care of collisions with existing services. This value is immutable. Set once, it can not be changed without deleting and re-creating the service. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#choosing-your-own-ip-address |
| controller.service.internal.enabled | bool | `false` | Enable the internal controller service or not. Remember to configure `controller.service.internal.annotations` when enabling this. |
| controller.service.internal.externalIPs | list | `[]` | List of node IP addresses at which the internal controller service is available. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#external-ips |
| controller.service.internal.externalTrafficPolicy | string | `""` | External traffic policy of the internal controller service. Set to "Local" to preserve source IP on providers supporting it. Ref: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/#preserving-the-client-source-ip |
| controller.service.internal.ipFamilies | list | `["IPv4"]` | List of IP families (e.g. IPv4, IPv6) assigned to the internal controller service. This field is usually assigned automatically based on cluster configuration and the `ipFamilyPolicy` field. Ref: https://kubernetes.io/docs/concepts/services-networking/dual-stack/#services |
| controller.service.internal.ipFamilyPolicy | string | `"SingleStack"` | Represents the dual-stack capabilities of the internal controller service. Possible values are SingleStack, PreferDualStack or RequireDualStack. Fields `ipFamilies` and `clusterIP` depend on the value of this field. Ref: https://kubernetes.io/docs/concepts/services-networking/dual-stack/#services |
| controller.service.internal.loadBalancerClass | string | `""` | Load balancer class of the internal controller service. Used by cloud providers to select a load balancer implementation other than the cloud provider default. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#load-balancer-class |
| controller.service.internal.loadBalancerIP | string | `""` | Deprecated: Pre-defined IP address of the internal controller service. Used by cloud providers to connect the resulting load balancer service to a pre-existing static IP. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer |
| controller.service.internal.loadBalancerSourceRanges | list | `[]` | Restrict access to the internal controller service. Values must be CIDRs. Allows any source address by default. |
| controller.service.internal.nodePorts.http | string | `""` | Node port allocated for the internal HTTP listener. If left empty, the service controller allocates one from the configured node port range. |
| controller.service.internal.nodePorts.https | string | `""` | Node port allocated for the internal HTTPS listener. If left empty, the service controller allocates one from the configured node port range. |
| controller.service.internal.nodePorts.tcp | object | `{}` | Node port mapping for internal TCP listeners. If left empty, the service controller allocates them from the configured node port range. Example: tcp:   8080: 30080 |
| controller.service.internal.nodePorts.udp | object | `{}` | Node port mapping for internal UDP listeners. If left empty, the service controller allocates them from the configured node port range. Example: udp:   53: 30053 |
| controller.service.internal.ports | object | `{}` |  |
| controller.service.internal.sessionAffinity | string | `""` | Session affinity of the internal controller service. Must be either "None" or "ClientIP" if set. Defaults to "None". Ref: https://kubernetes.io/docs/reference/networking/virtual-ips/#session-affinity |
| controller.service.internal.targetPorts | object | `{}` |  |
| controller.service.internal.type | string | `""` | Type of the internal controller service. Defaults to the value of `controller.service.type`. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types |
| controller.service.ipFamilies | list | `["IPv4"]` | List of IP families (e.g. IPv4, IPv6) assigned to the external controller service. This field is usually assigned automatically based on cluster configuration and the `ipFamilyPolicy` field. Ref: https://kubernetes.io/docs/concepts/services-networking/dual-stack/#services |
| controller.service.ipFamilyPolicy | string | `"SingleStack"` | Represents the dual-stack capabilities of the external controller service. Possible values are SingleStack, PreferDualStack or RequireDualStack. Fields `ipFamilies` and `clusterIP` depend on the value of this field. Ref: https://kubernetes.io/docs/concepts/services-networking/dual-stack/#services |
| controller.service.labels | object | `{}` | Labels to be added to both controller services. |
| controller.service.loadBalancerClass | string | `""` | Load balancer class of the external controller service. Used by cloud providers to select a load balancer implementation other than the cloud provider default. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#load-balancer-class |
| controller.service.loadBalancerIP | string | `""` | Deprecated: Pre-defined IP address of the external controller service. Used by cloud providers to connect the resulting load balancer service to a pre-existing static IP. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer |
| controller.service.loadBalancerSourceRanges | list | `[]` | Restrict access to the external controller service. Values must be CIDRs. Allows any source address by default. |
| controller.service.nodePorts.http | string | `""` | Node port allocated for the external HTTP listener. If left empty, the service controller allocates one from the configured node port range. |
| controller.service.nodePorts.https | string | `""` | Node port allocated for the external HTTPS listener. If left empty, the service controller allocates one from the configured node port range. |
| controller.service.nodePorts.tcp | object | `{}` | Node port mapping for external TCP listeners. If left empty, the service controller allocates them from the configured node port range. Example: tcp:   8080: 30080 |
| controller.service.nodePorts.udp | object | `{}` | Node port mapping for external UDP listeners. If left empty, the service controller allocates them from the configured node port range. Example: udp:   53: 30053 |
| controller.service.ports.http | int | `80` | Port the external HTTP listener is published with. |
| controller.service.ports.https | int | `443` | Port the external HTTPS listener is published with. |
| controller.service.sessionAffinity | string | `""` | Session affinity of the external controller service. Must be either "None" or "ClientIP" if set. Defaults to "None". Ref: https://kubernetes.io/docs/reference/networking/virtual-ips/#session-affinity |
| controller.service.targetPorts.http | string | `"http"` | Port of the ingress controller the external HTTP listener is mapped to. |
| controller.service.targetPorts.https | string | `"https"` | Port of the ingress controller the external HTTPS listener is mapped to. |
| controller.service.type | string | `"LoadBalancer"` | Type of the external controller service. Ref: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types |
| controller.shareProcessNamespace | bool | `false` |  |
| controller.sysctls | object | `{}` | sysctls for controller pods # Ref: https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/ |
| controller.tcp.annotations | object | `{}` | Annotations to be added to the tcp config configmap |
| controller.tcp.configMapNamespace | string | `""` | Allows customization of the tcp-services-configmap; defaults to $(POD_NAMESPACE) |
| controller.terminationGracePeriodSeconds | int | `300` | `terminationGracePeriodSeconds` to avoid killing pods before we are ready # wait up to five minutes for the drain of connections # |
| controller.tolerations | list | `[]` | Node tolerations for server scheduling to nodes with taints # Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ # |
| controller.topologySpreadConstraints | list | `[]` | Topology spread constraints rely on node labels to identify the topology domain(s) that each Node is in. # Ref: https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/ # |
| controller.udp.annotations | object | `{}` | Annotations to be added to the udp config configmap |
| controller.udp.configMapNamespace | string | `""` | Allows customization of the udp-services-configmap; defaults to $(POD_NAMESPACE) |
| controller.unhealthyPodEvictionPolicy | string | `""` | Eviction policy for unhealthy pods guarded by PodDisruptionBudget. Ref: https://kubernetes.io/blog/2023/01/06/unhealthy-pod-eviction-policy-for-pdbs/ |
| controller.updateStrategy | object | `{}` | The update strategy to apply to the Deployment or DaemonSet # |
| controller.watchIngressWithoutClass | bool | `false` | Process Ingress objects without ingressClass annotation/ingressClassName field Overrides value for --watch-ingress-without-class flag of the controller binary Defaults to false |
| defaultBackend.affinity | object | `{}` | Affinity and anti-affinity rules for server scheduling to nodes # Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#affinity-and-anti-affinity |
| defaultBackend.autoscaling.annotations | object | `{}` |  |
| defaultBackend.autoscaling.enabled | bool | `false` |  |
| defaultBackend.autoscaling.maxReplicas | int | `2` |  |
| defaultBackend.autoscaling.minReplicas | int | `1` |  |
| defaultBackend.autoscaling.targetCPUUtilizationPercentage | int | `50` |  |
| defaultBackend.autoscaling.targetMemoryUtilizationPercentage | int | `50` |  |
| defaultBackend.containerSecurityContext | object | `{}` | Security context for default backend containers |
| defaultBackend.enabled | bool | `false` |  |
| defaultBackend.extraArgs | object | `{}` |  |
| defaultBackend.extraConfigMaps | list | `[]` |  |
| defaultBackend.extraEnvs | list | `[]` | Additional environment variables to set for defaultBackend pods |
| defaultBackend.extraVolumeMounts | list | `[]` |  |
| defaultBackend.extraVolumes | list | `[]` |  |
| defaultBackend.image.allowPrivilegeEscalation | bool | `false` |  |
| defaultBackend.image.image | string | `"defaultbackend-amd64"` |  |
| defaultBackend.image.pullPolicy | string | `"IfNotPresent"` |  |
| defaultBackend.image.readOnlyRootFilesystem | bool | `true` |  |
| defaultBackend.image.runAsGroup | int | `65534` |  |
| defaultBackend.image.runAsNonRoot | bool | `true` |  |
| defaultBackend.image.runAsUser | int | `65534` |  |
| defaultBackend.image.seccompProfile.type | string | `"RuntimeDefault"` |  |
| defaultBackend.image.tag | string | `"1.5"` |  |
| defaultBackend.labels | object | `{}` | Labels to be added to the default backend resources |
| defaultBackend.livenessProbe.failureThreshold | int | `3` |  |
| defaultBackend.livenessProbe.initialDelaySeconds | int | `30` |  |
| defaultBackend.livenessProbe.periodSeconds | int | `10` |  |
| defaultBackend.livenessProbe.successThreshold | int | `1` |  |
| defaultBackend.livenessProbe.timeoutSeconds | int | `5` |  |
| defaultBackend.minAvailable | int | `1` | Minimum available pods set in PodDisruptionBudget. Define either 'minAvailable' or 'maxUnavailable', never both. |
| defaultBackend.minReadySeconds | int | `0` | `minReadySeconds` to avoid killing pods before we are ready # |
| defaultBackend.name | string | `"defaultbackend"` |  |
| defaultBackend.networkPolicy.enabled | bool | `false` | Enable 'networkPolicy' or not |
| defaultBackend.nodeSelector | object | `{"kubernetes.io/os":"linux"}` | Node labels for default backend pod assignment # Ref: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/ # |
| defaultBackend.podAnnotations | object | `{}` | Annotations to be added to default backend pods # |
| defaultBackend.podLabels | object | `{}` | Labels to add to the pod container metadata |
| defaultBackend.podSecurityContext | object | `{}` | Security context for default backend pods |
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
| defaultBackend.service.externalIPs | list | `[]` | List of IP addresses at which the default backend service is available # Ref: https://kubernetes.io/docs/concepts/services-networking/service/#external-ips # |
| defaultBackend.service.loadBalancerSourceRanges | list | `[]` |  |
| defaultBackend.service.servicePort | int | `80` |  |
| defaultBackend.service.type | string | `"ClusterIP"` |  |
| defaultBackend.serviceAccount.automountServiceAccountToken | bool | `true` |  |
| defaultBackend.serviceAccount.create | bool | `true` |  |
| defaultBackend.serviceAccount.name | string | `""` |  |
| defaultBackend.tolerations | list | `[]` | Node tolerations for server scheduling to nodes with taints # Ref: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/ # |
| defaultBackend.topologySpreadConstraints | list | `[]` | Topology spread constraints rely on node labels to identify the topology domain(s) that each Node is in. Ref.: https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/ |
| defaultBackend.unhealthyPodEvictionPolicy | string | `""` | Eviction policy for unhealthy pods guarded by PodDisruptionBudget. Ref: https://kubernetes.io/blog/2023/01/06/unhealthy-pod-eviction-policy-for-pdbs/ |
| defaultBackend.updateStrategy | object | `{}` | The update strategy to apply to the Deployment or DaemonSet # |
| dhParam | string | `""` | A base64-encoded Diffie-Hellman parameter. This can be generated with: `openssl dhparam 4096 2> /dev/null | base64` # Ref: https://github.com/kubernetes/ingress-nginx/tree/main/docs/examples/customization/ssl-dh-param |
| global.image.registry | string | `"registry.k8s.io"` | Registry host to pull images from. |
| imagePullSecrets | list | `[]` | Optional array of imagePullSecrets containing private registry credentials # Ref: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/ |
| namespaceOverride | string | `""` | Override the deployment namespace; defaults to .Release.Namespace |
| portNamePrefix | string | `""` | Prefix for TCP and UDP ports names in ingress controller service # Some cloud providers, like Yandex Cloud may have a requirements for a port name regex to support cloud load balancer integration |
| rbac.create | bool | `true` |  |
| rbac.scope | bool | `false` |  |
| revisionHistoryLimit | int | `10` | Rollback limit # |
| serviceAccount.annotations | object | `{}` | Annotations for the controller service account |
| serviceAccount.automountServiceAccountToken | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| tcp | object | `{}` | TCP service key-value pairs # Ref: https://github.com/kubernetes/ingress-nginx/blob/main/docs/user-guide/exposing-tcp-udp-services.md # |
| udp | object | `{}` | UDP service key-value pairs # Ref: https://github.com/kubernetes/ingress-nginx/blob/main/docs/user-guide/exposing-tcp-udp-services.md # |
