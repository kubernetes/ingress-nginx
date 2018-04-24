# Multiple ingress controllers

## Running multiple ingress controllers

If you're running multiple ingress controllers, or running on a cloud provider that natively handles ingress, you need to specify the annotation `kubernetes.io/ingress.class: "nginx"` in all ingresses that you would like this controller to claim.  This mechanism also provides users the ability to run _multiple_ NGINX ingress controllers (e.g. one which serves public traffic, one which serves "internal" traffic).  When utilizing this functionality the option `--ingress-class` should be changed to a value unique for the cluster within the definition of the replication controller. Here is a partial example:

```
spec:
  template:
     spec:
       containers:
         - name: nginx-ingress-internal-controller
           args:
             - /nginx-ingress-controller
             - '--default-backend-service=ingress/nginx-ingress-default-backend'
             - '--election-id=ingress-controller-leader-internal'
             - '--ingress-class=nginx-internal'
             - '--configmap=ingress/nginx-ingress-internal-controller'
```


## Annotation ingress.class

If you have multiple Ingress controllers in a single cluster, you can pick one by specifying the `ingress.class` 
annotation, eg creating an Ingress with an annotation like

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

__Note__: Deploying multiple ingress controller and not specifying the annotation will result in both controllers fighting to satisfy the Ingress.

## Disabling NGINX ingress controller

Setting the annotation `kubernetes.io/ingress.class` to any other value  which does not match a valid ingress class will force the NGINX Ingress controller to ignore your Ingress.  If you are only running a single NGINX ingress controller, this can be achieved by setting this to any value except "nginx" or an empty string.

Do this if you wish to use one of the other Ingress controllers at the same time as the NGINX controller.

