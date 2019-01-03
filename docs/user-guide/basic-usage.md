#  Basic usage - host based routing

ingress-nginx can be used for many use cases, inside various cloud provider and supports a lot of configurations. In this section you can find a common usage scenario where a single load balancer powerd by ingress-nginx will route traffic to 2 different HTTP backend services based on the host name.

First of all follow the instructions to install ingress-nginx. Then imagine that you need to expose 2 HTTP services already installed: `myServiceA`, `myServiceB`. Let's say that you want to expose the first at `myServiceA.foo.org` and the second at `myServiceB.foo.org`. One possible solution is to create two **ingress** resources:

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: ingress-myServiceA
  annotations:
    # use the shared ingress-nginx
    kubernetes.io/ingress.class: "nginx"
spec:
  rules:
  - host: myServiceA.foo.org
    http:
      paths:
      - path: /
        backend:
          serviceName: myServiceA
          servicePort: 80
---
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: ingress-myServiceB
  annotations:
    # use the shared ingress-nginx
    kubernetes.io/ingress.class: "nginx"
spec:
  rules:
  - host: myServiceB.foo.org
    http:
      paths:
      - path: /
        backend:
          serviceName: myServiceB
          servicePort: 80
```

When you apply this yaml, 2 ingress resources will be created managed by the **ingress-nginx** instance. Nginx is configured to automatically discover all ingress with the `kubernetes.io/ingress.class: "nginx"` annotation.
Please note that the ingress resource should be placed inside the same namespace of the backend resource.

On many cloud providers ingress-nginx will also create the corresponding Load Balancer resource. All you have to do is get the external IP and add a DNS `A record` inside your DNS provider that point myServiceA.foo.org and myServiceB.foo.org to the nginx external IP. Get the external IP by running:

```
kubectl get services -n ingress-nginx
```
