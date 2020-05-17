# Custom Errors

This example demonstrates how to use a custom backend to render custom error pages.

## Customized default backend

First, create the custom `default-backend`. It will be used by the Ingress controller later on.

```
$ kubectl create -f custom-default-backend.yaml
service "nginx-errors" created
deployment.apps "nginx-errors" created
```

This should have created a Deployment and a Service with the name `nginx-errors`.

```
$ kubectl get deploy,svc
NAME                           DESIRED   CURRENT   READY     AGE
deployment.apps/nginx-errors   1         1         1         10s

NAME                   TYPE        CLUSTER-IP  EXTERNAL-IP   PORT(S)   AGE
service/nginx-errors   ClusterIP   10.0.0.12   <none>        80/TCP    10s
```

## Ingress controller configuration

If you do not already have an instance of the NGINX Ingress controller running, deploy it according to the
[deployment guide][deploy], then follow these steps:

1. Edit the `ingress-nginx-controller` Deployment and set the value of the `--default-backend-service` flag to the name of the
   newly created error backend.

2. Edit the `ingress-nginx-controller` ConfigMap and create the key `custom-http-errors` with a value of `404,503`.

3. Take note of the IP address assigned to the NGINX Ingress controller Service.
    ```
    $ kubectl get svc ingress-nginx
    NAME            TYPE        CLUSTER-IP  EXTERNAL-IP   PORT(S)          AGE
    ingress-nginx   ClusterIP   10.0.0.13   <none>        80/TCP,443/TCP   10m
    ```

!!! note
    The `ingress-nginx` Service is of type `ClusterIP` in this example. This may vary depending on your environment.
    Make sure you can use the Service to reach NGINX before proceeding with the rest of this example.

[deploy]: ../../../deploy/

## Testing error pages

Let us send a couple of HTTP requests using cURL and validate everything is working as expected.

A request to the default backend returns a 404 error with a custom message:

```
$ curl -D- http://10.0.0.13/
HTTP/1.1 404 Not Found
Server: nginx/1.13.12
Date: Tue, 12 Jun 2018 19:11:24 GMT
Content-Type: */*
Transfer-Encoding: chunked
Connection: keep-alive

<span>The page you're looking for could not be found.</span>
```

A request with a custom `Accept` header returns the corresponding document type (JSON):

```
$ curl -D- -H 'Accept: application/json' http://10.0.0.13/
HTTP/1.1 404 Not Found
Server: nginx/1.13.12
Date: Tue, 12 Jun 2018 19:12:36 GMT
Content-Type: application/json
Transfer-Encoding: chunked
Connection: keep-alive
Vary: Accept-Encoding

{ "message": "The page you're looking for could not be found" }
```

To go further with this example, feel free to deploy your own applications and Ingress objects, and validate that the
responses are still in the correct format when a backend returns 503 (eg. if you scale a Deployment down to 0 replica).
