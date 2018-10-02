# Multiple Ingress controllers

If you're running multiple ingress controllers, or running on a cloud provider that natively handles ingress such as GKE,
you need to specify the annotation `kubernetes.io/ingress.class: "nginx"` in all ingresses that you would like the ingress-nginx controller to claim.


For instance,

```yaml
metadata:
  name: foo
  annotations:
    kubernetes.io/ingress.class: "gce"
```

will target the GCE controller, forcing the nginx controller to ignore it, while an annotation like

```yaml
metadata:
  name: foo
  annotations:
    kubernetes.io/ingress.class: "nginx"
```

will target the nginx controller, forcing the GCE controller to ignore it.

To reiterate, setting the annotation to any value which does not match a valid ingress class will force the NGINX Ingress controller to ignore your Ingress.
If you are only running a single NGINX ingress controller, this can be achieved by setting the annotation to any value except "nginx" or an empty string.

Do this if you wish to use one of the other Ingress controllers at the same time as the NGINX controller.

## Multiple ingress-nginx controllers

This mechanism also provides users the ability to run _multiple_ NGINX ingress controllers (e.g. one which serves public traffic, one which serves "internal" traffic).
To do this, the option `--ingress-class` must be changed to a value unique for the cluster within the definition of the replication controller.
Here is a partial example:

```yaml
spec:
  template:
     spec:
       containers:
         - name: nginx-ingress-internal-controller
           args:
             - /nginx-ingress-controller
             - '--election-id=ingress-controller-leader-internal'
             - '--ingress-class=nginx-internal'
             - '--configmap=ingress/nginx-ingress-internal-controller'
```

!!! important
    Deploying multiple Ingress controllers, of different types (e.g., `ingress-nginx` & `gce`), and not specifying a class annotation will
    result in both or all controllers fighting to satisfy the Ingress, and all of them racing to update Ingress status field in confusing ways.

    When running multiple ingress-nginx controllers, it will only process an unset class annotation if one of the controllers uses the default
    `--ingress-class` value (see `IsValid` method in `internal/ingress/annotations/class/main.go`), otherwise the class annotation become required.
