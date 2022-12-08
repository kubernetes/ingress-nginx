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
  name: ingress-nginx-controller
  namespace: ingress-nginx
spec:
  replicas: 1
  selector: ...
  template:
    metadata: ...
    spec:
      containers:
        - name: ingress-nginx-controller
          image: registry.k8s.io/ingress-nginx/controller:v1.0.4@sha256:545cff00370f28363dad31e3b59a94ba377854d3a11f18988f5f9e56841ef9ef
          args: ...
```

simply change the `v1.0.4` tag to the version you wish to upgrade to.
The easiest way to do this is e.g. (do note you may need to change the name parameter according to your installation):

```
kubectl set image deployment/ingress-nginx-controller \
  controller=registry.k8s.io/ingress-nginx/controller:v1.0.5@sha256:55a1fcda5b7657c372515fe402c3e39ad93aa59f6e4378e82acd99912fe6028d \
  -n ingress-nginx
```

For interactive editing, use `kubectl edit deployment ingress-nginx-controller -n ingress-nginx`.

## With Helm

If you installed ingress-nginx using the Helm command in the deployment docs so its name is `ingress-nginx`,
you should be able to upgrade using

```shell
helm upgrade --reuse-values ingress-nginx ingress-nginx/ingress-nginx
```

### Migrating from stable/nginx-ingress

See detailed steps in the upgrading section of the `ingress-nginx` chart [README](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/README.md#migrating-from-stablenginx-ingress).
