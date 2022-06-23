# OpenTelemetry sidecar

Enables distributed tracing using [OpenTelemetry](https://opentelemetry.io/).

Using the OpenTelemetry sidecar enables the OpenTelemetry nginx module [OpenTelemetry nginx module](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx) for the NGINX ingress controller.
By default this feature is disabled.

## Usage

To enable the OpenTelemetry nginx module, the sidecar must be activated and configured.

### Enable sidecar

The OpenTelemetry sidecar load the OpenTelemetry module into the ingress-nginx container.
To enable the sidecar use the [helm chart](https://github.com/kubernetes/ingress-nginx/tree/main/charts/ingress-nginx):
```
  extraModules: 
  - name: opentelemetry
    image: registry.k8s.io/ingress-nginx/opentelemetry
```
<!-- TODO: correct image name -->

Otherwise add to the controller deployment configuration:
```
spec:
  template:
    spec:
      containers:
          volumeMounts:
            - name: modules
              mountPath: /modules_mount
      initContainers:
        - name: opentelemetry
          image: registry.k8s.io/ingress-nginx/opentelemetry
          command: ['sh', '-c', '/usr/local/bin/init_module.sh']
          volumeMounts:
            - name: modules
              mountPath: /modules_mount
      volumes:
        - name: modules
          emptyDir: {}
```

### enable module

To enable the Opentelemetry nginx module add to the configuration ConfigMap:
```
data:
  enable-opentelemetry: "true"
```

The [OpenTelemetry C++ library](https://github.com/open-telemetry/opentelemetry-cpp), base of the OpenTelemetry nginx module, requires a configuration file for processor and exporter configuration. The configuration file is `mandatory`.
```
data:
  opentelemetry-config: /conf/otel-nginx.toml
```

#### module configuration
The module configuration for processors and exporters can be added by an additional ConfigMap:
```
{
    "kind": "ConfigMap",
    "apiVersion": "v1",
    "metadata": {
        "name": "nginx-otel",
        "namespace": "ingress-nginx",
        "creationTimestamp": null
    },
    "data": {
        "otel-nginx.toml": "exporter = \"otlp\"\nprocessor = \"batch\"\n\n[exporters.otlp]\n# Alternatively the OTEL_EXPORTER_OTLP_ENDPOINT environment variable can also be used.\nhost = \"localhost\"\nport = 4317\n\n[processors.batch]\nmax_queue_size = 2048\nschedule_delay_millis = 5000\nmax_export_batch_size = 512\n\n[service]\nname = \"nginx-proxy\" # Opentelemetry resource name\n\n[sampler]\nname = \"AlwaysOn\" # Also: AlwaysOff, TraceIdRatioBased\nratio = 0.1\nparent_based = false\n"
    }
}
```
Currently this ConfigMap must be added manually.
The configuration here is equal to [OpenTelemetry C++ usage description](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx#usage)

Additionally the `otel-nginx.toml` file must be mounted to the ingress-nginx container
```
spec:
  template:
    spec:
      containers:
          volumeMounts:
            - name: otel-nginx
              readOnly: true
              mountPath: /conf
      volumes:
        - name: otel-nginx
          configMap:
            name: nginx-otel
            items:
              - key: otel-nginx.toml
                path: otel-nginx.toml
```

### nginx configuration
The OpenTelemetry nginx module provides several [nginx directives](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx#nginx-directives).

`Comming soon.`

### helm chart

The easiest way to activate the sidecar is to use the [helm chart](https://github.com/kubernetes/ingress-nginx/tree/main/charts/ingress-nginx). The values.yaml file have to contain

```
  extraModules: 
  - name: opentelemetry
    image: registry.k8s.io/ingress-nginx/opentelemetry:v20220415-controller-v1.2.0-beta.0-2-g81c2afd97@sha256:ce61e2cf0b347dffebb2dcbf57c33891d2217c1bad9c0959c878e5be671ef941

  config:
    enable-opentelemetry: true
    opentelemetry-config: "/conf/otel-nginx.toml"

  extraVolumeMounts:
  - name: otel-nginx
    readOnly: true
    mountPath: /conf

  extraVolumes:
   - name: otel-nginx
     configMap:
      name: nginx-otel
```

## Collector
As OpenTelemetry collector use for example [OpenTelemetry Collector Helm Chart](https://github.com/open-telemetry/opentelemetry-helm-charts/tree/main/charts/opentelemetry-collector)