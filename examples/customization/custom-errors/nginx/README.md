This example shows how is possible to use a custom backend to render custom error pages. The code of this example is located here [nginx-debug-server](https://github.com/aledbf/contrib/tree/nginx-debug-server)


The idea is to use the headers `X-Code` and `X-Format` that NGINX pass to the backend in case of an error to find out the best existent representation of the response to be returned. i.e. if the request contains an `Accept` header of type `json` the error should be in that format and not in `html` (the default in NGINX).

First create the custom backend to use in the Ingress controller

```
$ kubectl create -f custom-default-backend.yaml
service "nginx-errors" created
replicationcontroller "nginx-errors" created
```

```
$ kubectl get svc
NAME                    CLUSTER-IP   EXTERNAL-IP   PORT(S)         AGE
echoheaders             10.3.0.7     nodes         80/TCP          23d
kubernetes              10.3.0.1     <none>        443/TCP         34d
nginx-errors            10.3.0.102   <none>        80/TCP          11s
```

```
$ kubectl get rc
CONTROLLER             REPLICAS   AGE
echoheaders            1          19d
nginx-errors           1          19s
```

Next create the Ingress controller executing
```
$ kubectl create -f rc-custom-errors.yaml
```

Now to check if this is working we use curl:

```
$ curl -v http://172.17.4.99/
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 80 (#0)
> GET / HTTP/1.1
> Host: 172.17.4.99
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 404 Not Found
< Server: nginx/1.10.0
< Date: Wed, 04 May 2016 02:53:45 GMT
< Content-Type: text/html
< Transfer-Encoding: chunked
< Connection: keep-alive
< Vary: Accept-Encoding
<
<span>The page you're looking for could not be found.</span>

* Connection #0 to host 172.17.4.99 left intact
```

Specifying json as expected format:

```
$ curl -v http://172.17.4.99/ -H 'Accept: application/json'
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 80 (#0)
> GET / HTTP/1.1
> Host: 172.17.4.99
> User-Agent: curl/7.43.0
> Accept: application/json
>
< HTTP/1.1 404 Not Found
< Server: nginx/1.10.0
< Date: Wed, 04 May 2016 02:54:00 GMT
< Content-Type: text/html
< Transfer-Encoding: chunked
< Connection: keep-alive
< Vary: Accept-Encoding
<
{ "message": "The page you're looking for could not be found" }

* Connection #0 to host 172.17.4.99 left intact
```

By default the Ingress controller provides support for `html`, `json` and `XML`.
