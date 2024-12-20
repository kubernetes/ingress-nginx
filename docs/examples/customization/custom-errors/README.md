# Custom Errors

This example demonstrates how to use a custom backend to render custom error pages.

If you are using the Helm Chart, look at [example values](https://github.com/kubernetes/ingress-nginx/blob/main/docs/examples/customization/custom-errors/custom-default-backend.helm.values.yaml) and don't forget to add the [ConfigMap](https://github.com/kubernetes/ingress-nginx/blob/main/docs/examples/customization/custom-errors/custom-default-backend-error_pages.configMap.yaml) to your deployment. Otherwise, continue with [Customized default backend](#customized-default-backend) manual deployment.

## Customized default backend

First, create the custom `default-backend`. It will be used by the Ingress controller later on.

To do that, you can take a look at the [example manifest](https://github.com/kubernetes/ingress-nginx/blob/main/docs/examples/customization/custom-errors/custom-default-backend.yaml)
in this project's GitHub repository.

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

If you do not already have an instance of the Ingress-Nginx Controller running, deploy it according to the
[deployment guide][deploy], then follow these steps:

1. Edit the `ingress-nginx-controller` Deployment and set the value of the `--default-backend-service` flag to the name of the
   newly created error backend.

2. Edit the `ingress-nginx-controller` ConfigMap and create the key `custom-http-errors` with a value of `404,503`.

3. Take note of the IP address assigned to the Ingress-Nginx Controller Service.
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

## Maintenance page

You can also leverage custom error pages to set a **"_Service under maintenance_" page** for the whole cluster, useful to prevent users from accessing your services while you are performing planned scheduled maintenance.

When enabled, the maintenance page is served to the clients with an HTTP [**503 Service Unavailable**](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/503) response **status code**.

To do that:

- Enable a **custom error page for the 503 HTTP error**, by following the guide above
- Set the value of the `--watch-namespace-selector` flag to the name of some non-existent namespace, e.g. `nonexistent-namespace`
  - This effectively prevents the NGINX Ingress Controller from reading `Ingress` resources from any namespace in the Kubernetes cluster
- Set your `location-snippet` to `return 503;`, to make the NGINX Ingress Controller always return the 503 HTTP error page for all the requests
