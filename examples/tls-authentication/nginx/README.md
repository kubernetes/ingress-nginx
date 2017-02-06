# TLS termination

This example demonstrates how to enable the TLS Authentication through the nginx Ingress controller.

## Prerequisites

You need a valid CA File, composed of a group of valid enabled CAs. This MUST be in PEM Format.
Also the Ingress must terminate TLS, otherwise this makes no sense ;)

## Deployment

The following command instructs the controller to enable the TLS Authentication using
the secret containing the valid CA chains.

```console
$ kubectl create -f nginx-tls-auth.yaml
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
  tls-secret terminates ingress.test.com
Rules:
  Host	Path	Backends
  ----	----	--------
  *
    	 	http-svc:80 (<none>)
Annotations:
  auth-tls-secret:	default/caingress

Events:
  FirstSeen	LastSeen	Count	From				SubObjectPath	Type		Reason	Message
  ---------	--------	-----	----				-------------	--------	------	-------
  7s		7s		1	{nginx-ingress-controller }			Normal		CREATE	default/nginx-test
  7s		7s		1	{nginx-ingress-controller }			Normal		UPDATE	default/nginx-test
  7s		7s		1	{nginx-ingress-controller }			Normal		CREATE	ip: 104.198.183.6
  7s		7s		1	{nginx-ingress-controller }			Warning		MAPPING	Ingress rule 'default/nginx-test' contains no path definition. Assuming /


$ curl -k https://ingress.test.com
HTTP/1.1 400 Bad Request
Server: nginx/1.11.9

$ curl -I -k --key ~/user.key --cert ~/user.cer https://ingress.test.com 
HTTP/1.1 200 OK
Server: nginx/1.11.9

```
