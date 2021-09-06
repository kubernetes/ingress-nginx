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

You can mark a particular `IngressClass` as default for your cluster. Setting the `ingressclass.kubernetes.io/is-default-class` annotation to `true` on an `IngressClass` resource will ensure that new Ingresses without an `ingressClassName `field specified will be assigned this default `IngressClass`. But be aware that `ingressClass` works in a very specific way: you will need to change the `.spec.controller` value in your `IngressClass` and point the controller to the relevant `ingressClass.`
Here is a partial example:

    Ingress-Nginx-IngressClass-1 with .spec.controller equals to "k8s.io/ingress-nginx1"
    Ingress-Nginx-IngressClass-2 with .spec.controller equals to "k8s.io/ingress-nginx2" When deploying your ingress controllers, you will have to change the --controller-class field as follows:

Ingress-Nginx-Controller-nginx1 with k8s.io/ingress-nginx1 Ingress-Nginx-Controller-nginx2 with k8s.io/ingress-nginx2 Then, when you create an Ingress Object with IngressClassName = ingress-nginx2, it will look for controllers with controller-class=k8s.io/ingress-nginx2 and as Ingress-Nginx-Controller-nginx2 is watching objects that points to ingressClass="k8s.io/ingress-nginx2, it will serve that object, while Ingress-Nginx-Controller-nginx1 will ignore the ingress object.

Bear in mind that, if your Ingress-Nginx-Controller-nginx2 is started with the flag --watch-ingress-without-class=true, then it will serve ; - objects without ingress-class - objects with the annotation configured in flag --ingress-class and same class value - and also objects pointing to the ingressClass that have the same .spec.controller as configured in --controller-class


The flags regarding the option `--ingress-class` will be disabled soon instead use `ingressClassName` ingressClassName is a field in the specs of a ingress object.

```
% k explain ingress.spec.ingressClassName
KIND:     Ingress
VERSION:  networking.k8s.io/v1

FIELD:    ingressClassName <string>

DESCRIPTION:
     IngressClassName is the name of the IngressClass cluster resource. The
     associated IngressClass defines which controller will implement the
     resource. This replaces the deprecated `kubernetes.io/ingress.class`
     annotation. For backwards compatibility, when that annotation is set, it must be given precedence over this field. The controller may emit a warning if the field and annotation have different values. Implementations of this API should ignore Ingresses without a class specified. An IngressClass resource may be marked as default, which can be used to set a default value for this field. For more information, refer to the IngressClass documentation.
```
the `spec.ingressClassName` behavior has precedence over the annotation.

```yaml
spec:
  template:
     spec:
       containers:
         - name: nginx-ingress-internal-controller
           args:
             - /nginx-ingress-controller
             - '--ingress-class=nginx-internal'
             - '--configmap=ingress/nginx-ingress-internal-controller'
```

!!! important
    Deploying multiple Ingress controllers, of different types (e.g., `ingress-nginx` & `gce`), and not specifying a class annotation will
    result in both or all controllers fighting to satisfy the Ingress, and all of them racing to update Ingress status field in confusing ways.

    When running multiple ingress-nginx controllers, it will only process an unset class annotation if one of the controllers uses the default
    `--ingress-class` value (see `IsValid` method in `internal/ingress/annotations/class/main.go`), otherwise the class annotation become required.

    If `--ingress-class` is set to the default value of `nginx`, the controller will monitor Ingresses with no class annotation *and* Ingresses with annotation class set to `nginx`. Use a non-default value for `--ingress-class`, to ensure that the controller only satisfied the specific class of Ingresses.
