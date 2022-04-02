# TLS termination

This example demonstrates how to terminate TLS through the nginx Ingress controller.

## Prerequisites

You need a [TLS cert](../PREREQUISITES.md#tls-certificates) and a [test HTTP service](../PREREQUISITES.md#test-http-service) for this example.

## Deployment

Create a `ingress.yaml` file.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx-test
spec:
  tls:
    - hosts:
      - foo.bar.com
      # This assumes tls-secret exists and the SSL
      # certificate contains a CN for foo.bar.com
      secretName: tls-secret
  ingressClassName: nginx
  rules:
    - host: foo.bar.com
      http:
        paths:
        - path: /
          pathType: Prefix
          backend:
            # This assumes http-svc exists and routes to healthy endpoints
            service:
              name: http-svc
              port:
                number: 80
```

The following command instructs the controller to terminate traffic using the provided
TLS cert, and forward un-encrypted HTTP traffic to the test HTTP service.

```console
kubectl apply -f ingress.yaml
```

## Validation

You can confirm that the Ingress works.

```console
$ kubectl describe ing nginx-test
Name:			nginx-test
Namespace:		default
Address:		104.198.183.6
Default backend:	default-http-backend:80 (10.180.0.4:8080,10.240.0.2:8080)
TLS:
  tls-secret terminates
Rules:
  Host	Path	Backends
  ----	----	--------
  *
    	 	http-svc:80 (<none>)
Annotations:
Events:
  FirstSeen	LastSeen	Count	From				SubObjectPath	Type		Reason	Message
  ---------	--------	-----	----				-------------	--------	------	-------
  7s		7s		1	{ingress-nginx-controller }			Normal		CREATE	default/nginx-test
  7s		7s		1	{ingress-nginx-controller }			Normal		UPDATE	default/nginx-test
  7s		7s		1	{ingress-nginx-controller }			Normal		CREATE	ip: 104.198.183.6
  7s		7s		1	{ingress-nginx-controller }			Warning		MAPPING	Ingress rule 'default/nginx-test' contains no path definition. Assuming /

$ curl 104.198.183.6 -L
curl: (60) SSL certificate problem: self signed certificate
More details here: http://curl.haxx.se/docs/sslcerts.html

$ curl 104.198.183.6 -Lk
CLIENT VALUES:
client_address=10.240.0.4
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://35.186.221.137:8080/

SERVER VALUES:
server_version=nginx: 1.9.11 - lua: 10001

HEADERS RECEIVED:
accept=*/*
connection=Keep-Alive
host=35.186.221.137
user-agent=curl/7.46.0
via=1.1 google
x-cloud-trace-context=f708ea7e369d4514fc90d51d7e27e91d/13322322294276298106
x-forwarded-for=104.132.0.80, 35.186.221.137
x-forwarded-proto=https
BODY:

```
