#  Basic usage - host based routing

ingress-nginx can be used for many use cases, inside various cloud providers and supports a lot of configurations. In this section you can find a common usage scenario where a single load balancer powered by ingress-nginx will route traffic to 2 different HTTP backend services based on the host name.

First of all follow the instructions to install ingress-nginx. Then imagine that you need to expose 2 HTTP services already installed, `myServiceA`, `myServiceB`, and configured as `type: ClusterIP`. 

Let's say that you want to expose the first at `myServiceA.foo.org` and the second at `myServiceB.foo.org`.

If the cluster version is < 1.19, you can create two **ingress** resources like this:

```
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress-myservicea
spec:
  ingressClassName: nginx
  rules:
  - host: myservicea.foo.org
    http:
      paths:
      - path: /
        backend:
          serviceName: myservicea
          servicePort: 80
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: ingress-myserviceb
  annotations:
    # use the shared ingress-nginx
    kubernetes.io/ingress.class: "nginx"
spec:
  rules:
  - host: myserviceb.foo.org
    http:
      paths:
      - path: /
        backend:
          serviceName: myserviceb
          servicePort: 80
```

If the cluster uses Kubernetes version >= 1.19.x, then its suggested to create 2 ingress resources, using yaml examples shown below. These examples are in conformity with the `networking.kubernetes.io/v1` api.

```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-myservicea
spec:
  rules:
  - host: myservicea.foo.org
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: myservicea
            port:
              number: 80
  ingressClassName: nginx
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: ingress-myserviceb
spec:
  rules:
  - host: myserviceb.foo.org
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: myserviceb
            port:
              number: 80
  ingressClassName: nginx
```

When you apply this yaml, 2 ingress resources will be created managed by the **ingress-nginx** instance. Nginx is configured to automatically discover all ingress with the `kubernetes.io/ingress.class: "nginx"` annotation or where `ingressClassName: nginx` is present.
Please note that the ingress resource should be placed inside the same namespace of the backend resource.

On many cloud providers ingress-nginx will also create the corresponding Load Balancer resource. All you have to do is get the external IP and add a DNS `A record` inside your DNS provider that point myservicea.foo.org and myserviceb.foo.org to the nginx external IP. Get the external IP by running:

```
kubectl get services -n ingress-nginx
```

To test inside minikube refer to this documentation: [Set up Ingress on Minikube with the NGINX Ingress Controller](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/)
