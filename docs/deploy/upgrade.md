# Upgrading

!!! important
    No matter the method you use for upgrading, _if you use template overrides,
    make sure your templates are compatible with the new version of ingress-nginx_.

## Without Helm

To upgrade your ingress-nginx installation, it should be enough to change the version of the image
in the controller Deployment.

I.e. if your deployment resource looks like (partial example):

```yaml
kind: Deployment
metadata:
  name: nginx-ingress-controller
  namespace: ingress-nginx
spec:
  replicas: 1
  selector: ...
  template:
    metadata: ...
    spec:
      containers:
        - name: nginx-ingress-controller
          image: quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.30.0
          args: ...
```

simply change the `0.30.0` tag to the version you wish to upgrade to.
The easiest way to do this is e.g. (do note you may need to change the name parameter according to your installation):

```
kubectl set image deployment/nginx-ingress-controller \
  nginx-ingress-controller=quay.io/kubernetes-ingress-controller/nginx-ingress-controller:0.33.0
```

For interactive editing, use `kubectl edit deployment nginx-ingress-controller`.

## With Helm

If you installed ingress-nginx using the Helm command in the deployment docs so its name is `ngx-ingress`,
you should be able to upgrade using

```shell
helm upgrade --reuse-values ngx-ingress ingress-nginx/ingress-nginx
```
