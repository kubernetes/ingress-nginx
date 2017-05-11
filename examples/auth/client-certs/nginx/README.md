# TLS authentication 

This example demonstrates how to enable the TLS Authentication through the nginx Ingress controller.

## Terminology

* CA: Certificate authority signing the client cert, in this example we will play the role of a CA. 
You can generate a CA cert as show in this doc.

* CA Certificate(s) - Certificate Authority public key. Client certs must chain back to this cert, 
meaning the Issuer field of some certificate in the chain leading up to the client cert must contain 
the name of this CA. For purposes of this example, this is a self signed certificate.

* CA chains: A chain of certificates where the parent has a Subject field matching the Issuer field of 
the child, except for the root, which has Issuer == Subject.

* Client Cert: Certificate used by the clients to authenticate themselves with the loadbalancer/backends.


## Prerequisites

You need a valid CA File, composed of a group of valid enabled CAs. This MUST be in PEM Format.
The instructions are described [here](../../../PREREQUISITES.md#ca-authentication)

Also your ingress must be configured as a HTTPs/TLS Ingress.

## Deployment

Certificate Authentication is achieved through 2 annotations on the Ingress, as shown in the [example](nginx-tls-auth.yaml).

|Name|Description|Values|
| --- | --- | --- |
|ingress.kubernetes.io/auth-tls-secret|Sets the secret that contains the authorized CA Chain|string|
|ingress.kubernetes.io/auth-tls-verify-depth|The verification depth Certificate Authentication will make|number (default to 1)|


The following command instructs the controller to enable TLS authentication using the secret from the ``ingress.kubernetes.io/auth-tls-secret``
annotation on the Ingress. Clients must present this cert to the loadbalancer, or they will receive a HTTP 400 response

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
  auth-tls-verify-depth: 3

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

You must use the full DNS name while testing, as NGINX relies on the Server Name (SNI) to select the correct Ingress to be used.

The curl version used here was ``curl 7.47.0``

## Which certificate was used for authentication?

In your backend application you might want to know which certificate was used for authentication. For this purpose, we pass the full certificate in PEM format to the backend in the `ssl-client-cert` header.
