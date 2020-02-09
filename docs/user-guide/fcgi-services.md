

# Exposing FastCGI Servers

> **FastCGI** is a [binary protocol](https://en.wikipedia.org/wiki/Binary_protocol "Binary protocol") for interfacing interactive programs with a [web server](https://en.wikipedia.org/wiki/Web_server "Web server"). [...] (It's) aim is to reduce the overhead related to interfacing between web server and CGI programs, allowing a server to handle more web page requests per unit of time.
>
> &mdash; Wikipedia

The _ingress-nginx_ ingress controller can be used to directly expose [FastCGI](https://en.wikipedia.org/wiki/FastCGI) servers.  Enabling FastCGI in your Ingress only requires setting the _backend-protocol_ annotation to `FCGI`, and with a couple more annotations you can customize the way _ingress-nginx_ handles the communication with your FastCGI _server_.


## Example Objects to Expose a FastCGI Pod

The _Pod_ example object below exposes port `9000`, which is the conventional FastCGI port.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: example-app
labels:
  app: example-app
spec:
  containers:
  - name: example-app
    image: example-app:1.0
    ports:
    - containerPort: 9000
      name: fastcgi
```

The _Service_ object example below matches port `9000` from the _Pod_ object above.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: example-service
spec:
  selector:
    app: example-app
  ports:
  - port: 9000
    targetPort: 9000
    name: fastcgi
```

And the _Ingress_ and _ConfigMap_ objects below demonstrates the supported _FastCGI_ specific annotations (NGINX actually has 50 FastCGI directives, all of which have not been exposed in the ingress yet), and matches the service `example-service`, and the port named `fastcgi` from above. The _ConfigMap_ **must** be created first for the _Ingress Controller_ to be able to find it when the _Ingress_ object is created, otherwise you will need to restart the _Ingress Controller_ pods.

```yaml
# The ConfigMap MUST be created first for the ingress controller to be able to
# find it when the Ingress object is created.

apiVersion: v1
kind: ConfigMap
metadata:
  name: example-cm
data:
  SCRIPT_FILENAME: "/example/index.php"

---

apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: "nginx"
    nginx.ingress.kubernetes.io/backend-protocol: "FCGI"
    nginx.ingress.kubernetes.io/fastcgi-index: "index.php"
    nginx.ingress.kubernetes.io/fastcgi-params-configmap: "example-cm"
  name: example-app
spec:
  rules:
  - host: app.example.com
    http:
      paths:
      - backend:
          serviceName: example-service
          servicePort: fastcgi
```

## FastCGI Ingress Annotations

To enable FastCGI, the `nginx.ingress.kubernetes.io/backend-protocol` annotation needs to be set to `FCGI`, which overrides the default `HTTP` value.

> `nginx.ingress.kubernetes.io/backend-protocol: "FCGI"`

**This enables the _FastCGI_ mode for all paths defined in the _Ingress_ object**

### The `nginx.ingress.kubernetes.io/fastcgi-index` Annotation

To specify an index file, the `fastcgi-index` annotation value can optionally be set.  In the example below, the value is set to `index.php`.  This annotation corresponds to [the _NGINX_ `fastcgi_index` directive](http://nginx.org/en/docs/http/ngx_http_fastcgi_module.html#fastcgi_index).

> `nginx.ingress.kubernetes.io/fastcgi-index: "index.php"`

### The `nginx.ingress.kubernetes.io/fastcgi-params-configmap` Annotation

To specify [_NGINX_ `fastcgi_param` directives](http://nginx.org/en/docs/http/ngx_http_fastcgi_module.html#fastcgi_param), the `fastcgi-params-configmap` annotation is used, which in turn must lead to a _ConfigMap_ object containing the _NGINX_ `fastcgi_param` directives as key/values.

> `nginx.ingress.kubernetes.io/fastcgi-params-configmap: "example-configmap"`

And the _ConfigMap_ object to specify the `SCRIPT_FILENAME` and `HTTP_PROXY`  _NGINX's_ `fastcgi_param` directives will look like the following:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-configmap
data:
  SCRIPT_FILENAME: "/example/index.php"
  HTTP_PROXY: ""
```
Using the _namespace/_ prefix is also supported, for example:

> `nginx.ingress.kubernetes.io/fastcgi-params-configmap: "example-namespace/example-configmap"`
