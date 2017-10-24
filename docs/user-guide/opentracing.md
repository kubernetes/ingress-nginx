# Opentracing

Using the third party module [rnburn/nginx-opentracing](https://github.com/rnburn/nginx-opentracing) the NGINX ingress controller can configure NGINX to enable [OpenTracing](http://opentracing.io) instrumentation.
By default this feature is disabled.

To enable the instrumentation we just need to enable the instrumentation in the configuration configmap and set the host where we should send the traces.

In the [rnburn/zipkin-date-server](https://github.com/rnburn/zipkin-date-server)
github repository is an example of a dockerized date service. Using [minikube](https://github.com/kubernetes/minikube), to install the example and zipkin collector run:

```
kubectl create -f https://raw.githubusercontent.com/rnburn/zipkin-date-server/master/kubernetes/zipkin.yaml
kubectl create -f https://raw.githubusercontent.com/rnburn/zipkin-date-server/master/kubernetes/deployment.yaml
```

In the [aledbf/zipkin-js-example](https://github.com/aledbf/zipkin-js-example) github repository is possible to see a dockerized version of zipkin-js-example with the required Kubernetes descriptors.
To install the example and the zipkin collector we just need to run:

```
kubectl create -f https://raw.githubusercontent.com/aledbf/zipkin-js-example/kubernetes/kubernetes/zipkin.yaml
kubectl create -f https://raw.githubusercontent.com/aledbf/zipkin-js-example/kubernetes/kubernetes/deployment.yaml
```

Also we need to configure the NGINX controller configmap with the required values:

```
$ echo '
apiVersion: v1
kind: ConfigMap
data:
  enable-opentracing: "true"
  zipkin-collector-host: zipkin.default.svc.cluster.local
metadata:
  name: nginx-configuration
  namespace: ingress-nginx
  labels:
    app: ingress-nginx
' | kubectl replace -f -
```

Using curl we can generate some traces:

```console
$ curl -v http://$(minikube ip)
$ curl -v http://$(minikube ip)
```

In the zipkin inteface we can see the details:

![zipkin screenshot](../images/zipkin-demo.png "zipkin collector screenshot")
