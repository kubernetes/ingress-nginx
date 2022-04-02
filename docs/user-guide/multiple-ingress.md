# Multiple Ingress controllers

By default, deploying multiple Ingress controllers (e.g., `ingress-nginx` & `gce`) will result in all controllers simultaneously racing to update Ingress status fields in confusing ways.

To fix this problem, use [IngressClasses](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class). The `kubernetes.io/ingress.class` annotation  is deprecated from kubernetes v1.22+.

## Using IngressClasses

If all ingress controllers respect IngressClasses (e.g. multiple instances of ingress-nginx v1.0), you can deploy two Ingress controllers by granting them control over two different IngressClasses, then selecting one of the two IngressClasses with `ingressClassName`.

First, ensure the `--controller-class=` and `--ingress-class` are set to something different on each ingress controller:

```yaml
# ingress-nginx Deployment/Statfulset
spec:
  template:
     spec:
       containers:
         - name: ingress-nginx-internal-controller
           args:
             - /nginx-ingress-controller
             - '--controller-class=k8s.io/internal-ingress-nginx'
             - '--ingress-class=k8s.io/internal-nginx'
            ...
```

Then use the same value in the IngressClass:

```yaml
# ingress-nginx IngressClass
apiVersion: networking.k8s.io/v1
kind: IngressClass
metadata:
  name: internal-nginx
spec:
  controller: k8s.io/internal-ingress-nginx
  ...
```

And refer to that IngressClass in your Ingress:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  ingressClassName: internal-nginx
  ...
```

or if installing with Helm:

```yaml
controller:
  ingressClassResource:
    name: internal-nginx  # default: nginx
    enabled: true
    default: false
    controllerValue: "k8s.io/internal-ingress-nginx"  # default: k8s.io/ingress-nginx
```

!!! important

    When running multiple ingress-nginx controllers, it will only process an unset class annotation if one of the controllers uses the default
    `--controller-class` value (see `IsValid` method in `internal/ingress/annotations/class/main.go`), otherwise the class annotation becomes required.

    If `--controller-class` is set to the default value of `k8s.io/ingress-nginx`, the controller will monitor Ingresses with no class annotation *and* Ingresses with annotation class set to `nginx`. Use a non-default value for `--controller-class`, to ensure that the controller only satisfied the specific class of Ingresses.

## Using the kubernetes.io/ingress.class annotation (in deprecation)

If you're running multiple ingress controllers where one or more do not support IngressClasses, you must specify the annotation `kubernetes.io/ingress.class: "nginx"` in all ingresses that you would like ingress-nginx to claim.


For instance,

```yaml
metadata:
  name: foo
  annotations:
    kubernetes.io/ingress.class: "gce"
```

will target the GCE controller, forcing the Ingress-NGINX controller to ignore it, while an annotation like:

```yaml
metadata:
  name: foo
  annotations:
    kubernetes.io/ingress.class: "nginx"
```

will target the Ingress-NGINX controller, forcing the GCE controller to ignore it.

You can change the value "nginx" to something else by setting the `--ingress-class` flag:

```yaml
spec:
  template:
     spec:
       containers:
         - name: ingress-nginx-internal-controller
           args:
             - /nginx-ingress-controller
             - --ingress-class=internal-nginx
```

then setting the corresponding `kubernetes.io/ingress.class: "internal-nginx"` annotation on your Ingresses.

To reiterate, setting the annotation to any value which does not match a valid ingress class will force the NGINX Ingress controller to ignore your Ingress.
If you are only running a single NGINX ingress controller, this can be achieved by setting the annotation to any value except "nginx" or an empty string.

Do this if you wish to use one of the other Ingress controllers at the same time as the NGINX controller.
