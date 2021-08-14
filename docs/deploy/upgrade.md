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
          image: k8s.gcr.io/ingress-nginx/controller:v0.34.0@sha256:56633bd00dab33d92ba14c6e709126a762d54a75a6e72437adefeaaca0abb069
          args: ...
```

simply change the `0.34.0` tag to the version you wish to upgrade to.
The easiest way to do this is e.g. (do note you may need to change the name parameter according to your installation):

```
kubectl set image deployment/nginx-ingress-controller \
  nginx-ingress-controller=k8s.gcr.io/ingress-nginx/controller:v1.0.0-beta.3@sha256:44a7a06b71187a4529b0a9edee5cc22bdf71b414470eff696c3869ea8d90a695 \
  -n ingress-nginx
```

For interactive editing, use `kubectl edit deployment nginx-ingress-controller -n ingress-nginx`.

## With Helm

If you installed ingress-nginx using the Helm command in the deployment docs so its name is `ngx-ingress`, you should be able to upgrade using

```
helm upgrade --reuse-values ngx-ingress ingress-nginx/ingress-nginx
```

### Migrating from stable/nginx-ingress

See detailed steps in the upgrading section of the `ingress-nginx` chart [README](https://github.com/kubernetes/ingress-nginx/blob/master/charts/ingress-nginx/README.md#migrating-from-stablenginx-ingress).
