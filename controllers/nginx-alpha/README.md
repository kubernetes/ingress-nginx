# Nginx Ingress Controller

This is a simple nginx Ingress controller. Expect it to grow up. See [Ingress controller documentation](../README.md) for details on how it works.

## Deploying the controller

Deploying the controller is as easy as creating the RC in this directory. Having done so you can test it with the following echoheaders application:

```yaml
# 3 Services for the 3 endpoints of the Ingress
apiVersion: v1
kind: Service
metadata:
  name: echoheaders-x
  labels:
    app: echoheaders
spec:
  type: NodePort
  ports:
  - port: 80
    nodePort: 30301
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: echoheaders
---
apiVersion: v1
kind: Service
metadata:
  name: echoheaders-default
  labels:
    app: echoheaders
spec:
  type: NodePort
  ports:
  - port: 80
    nodePort: 30302
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: echoheaders
---
apiVersion: v1
kind: Service
metadata:
  name: echoheaders-y
  labels:
    app: echoheaders
spec:
  type: NodePort
  ports:
  - port: 80
    nodePort: 30284
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: echoheaders
---
# A single RC matching all Services
apiVersion: v1
kind: ReplicationController
metadata:
  name: echoheaders
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: echoheaders
    spec:
      containers:
      - name: echoheaders
        image: gcr.io/google_containers/echoserver:1.0
        ports:
        - containerPort: 8080
---
# An Ingress with 2 hosts and 3 endpoints
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: echomap
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: /foo
        backend:
          serviceName: echoheaders-x
          servicePort: 80
  - host: bar.baz.com
    http:
      paths:
      - path: /bar
        backend:
          serviceName: echoheaders-y
          servicePort: 80
      - path: /foo
        backend:
          serviceName: echoheaders-x
          servicePort: 80
```
You should be able to access the Services on the public IP of the node the nginx pod lands on.

## Wishlist

* SSL/TLS
* Production ready options
* Dynamic adding backends
* Varied loadbalancing algorithms

... this list goes on. If you feel you know nginx better than I do, please contribute.

