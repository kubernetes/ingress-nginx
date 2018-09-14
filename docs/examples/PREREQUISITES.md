# Prerequisites

Many of the examples in this directory have common prerequisites.

## TLS certificates

Unless otherwise mentioned, the TLS secret used in examples is a 2048 bit RSA
key/cert pair with an arbitrarily chosen hostname, created as follows

```console
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout tls.key -out tls.crt -subj "/CN=nginxsvc/O=nginxsvc"
Generating a 2048 bit RSA private key
................+++
................+++
writing new private key to 'tls.key'
-----

$ kubectl create secret tls tls-secret --key tls.key --cert tls.crt
secret "tls-secret" created
```

## CA Authentication

You can act as your very own CA, or use an existing one. As an exercise / learning, we're going to generate our
own CA, and also generate a client certificate.

These instructions are based on CoreOS OpenSSL. [See live doc.](https://coreos.com/kubernetes/docs/latest/openssl.html)

### Generating a CA

First of all, you've to generate a CA. This is going to be the one who will sign your client certificates.
In real production world, you may face CAs with intermediate certificates, as the following:

```console
$ openssl s_client -connect www.google.com:443
[...]
---
Certificate chain
 0 s:/C=US/ST=California/L=Mountain View/O=Google Inc/CN=www.google.com
   i:/C=US/O=Google Inc/CN=Google Internet Authority G2
 1 s:/C=US/O=Google Inc/CN=Google Internet Authority G2
   i:/C=US/O=GeoTrust Inc./CN=GeoTrust Global CA
 2 s:/C=US/O=GeoTrust Inc./CN=GeoTrust Global CA
   i:/C=US/O=Equifax/OU=Equifax Secure Certificate Authority

```

To generate our CA Certificate, we've to run the following commands:

```console
$ openssl genrsa -out ca.key 2048
$ openssl req -x509 -new -nodes -key ca.key -days 10000 -out ca.crt -subj "/CN=example-ca"
```

This will generate two files: A private key (ca.key) and a public key (ca.crt). This CA is valid for 10000 days.
The ca.crt can be used later in the step of creation of CA authentication secret.

### Generating the client certificate

The following steps generate a client certificate signed by the CA generated above. This client can be
used to authenticate in a tls-auth configured ingress.

First, we need to generate an 'openssl.cnf' file that will be used while signing the keys:

```console
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
```

Then, a user generates his very own private key (that he needs to keep secret)
and a CSR (Certificate Signing Request) that will be sent to the CA to sign and generate a certificate.

```console
$ openssl genrsa -out client1.key 2048
$ openssl req -new -key client1.key -out client1.csr -subj "/CN=client1" -config openssl.cnf
```

As the CA receives the generated 'client1.csr' file, it signs it and generates a client.crt certificate:

```console
$ openssl x509 -req -in client1.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client1.crt -days 365 -extensions v3_req -extfile openssl.cnf
```

Then, you'll have 3 files: the client.key (user's private key), client.crt (user's public key) and client.csr (disposable CSR).

### Creating the CA Authentication secret

If you're using the CA Authentication feature, you need to generate a secret containing 
all the authorized CAs. You must download them from your CA site in PEM format (like the following):

```
-----BEGIN CERTIFICATE-----
[....]
-----END CERTIFICATE-----
``` 

You can have as many certificates as you want. If they're in the binary DER format, 
you can convert them as the following:

```console
$ openssl x509 -in certificate.der -inform der -out certificate.crt -outform pem
```

Then, you've to concatenate them all in only one file, named 'ca.crt' as the following:

```console
$ cat certificate1.crt certificate2.crt certificate3.crt >> ca.crt
```

The final step is to create a secret with the content of this file. This secret is going to be used in 
the TLS Auth directive:

```console
$ kubectl create secret generic caingress --namespace=default --from-file=ca.crt=<ca.crt>
```

__Note:__ You can also generate the CA Authentication Secret along with the TLS Secret by using:
```console
$ kubectl create secret generic caingress --namespace=default --from-file=ca.crt=<ca.crt> --from-file=tls.crt=<tls.crt> --from-file=tls.key=<tls.key>
```

## Test HTTP Service

All examples that require a test HTTP Service use the standard http-svc pod,
which you can deploy as follows

```console
$ kubectl create -f http-svc.yaml
service "http-svc" created
replicationcontroller "http-svc" created

$ kubectl get po
NAME             READY     STATUS    RESTARTS   AGE
http-svc-p1t3t   1/1       Running   0          1d

$ kubectl get svc
NAME             CLUSTER-IP     EXTERNAL-IP   PORT(S)            AGE
http-svc         10.0.122.116   <pending>     80:30301/TCP       1d
```

You can test that the HTTP Service works by exposing it temporarily

```console
$ kubectl patch svc http-svc -p '{"spec":{"type": "LoadBalancer"}}'
"http-svc" patched

$ kubectl get svc http-svc
NAME             CLUSTER-IP     EXTERNAL-IP   PORT(S)            AGE
http-svc         10.0.122.116   <pending>     80:30301/TCP       1d

$ kubectl describe svc http-svc
Name:				    http-svc
Namespace:			    default
Labels:			        app=http-svc
Selector:		        app=http-svc
Type:			        LoadBalancer
IP:			            10.0.122.116
LoadBalancer Ingress:	108.59.87.136
Port:			        http	80/TCP
NodePort:		        http	30301/TCP
Endpoints:		        10.180.1.6:8080
Session Affinity:	    None
Events:
  FirstSeen	LastSeen	Count	From			SubObjectPath	Type		Reason			Message
  ---------	--------	-----	----			-------------	--------	------			-------
  1m		1m		1	{service-controller }			Normal		Type			ClusterIP -> LoadBalancer
  1m		1m		1	{service-controller }			Normal		CreatingLoadBalancer	Creating load balancer
  16s		16s		1	{service-controller }			Normal		CreatedLoadBalancer	Created load balancer

$ curl 108.59.87.136
CLIENT VALUES:
client_address=10.240.0.3
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://108.59.87.136:8080/

SERVER VALUES:
server_version=nginx: 1.9.11 - lua: 10001

HEADERS RECEIVED:
accept=*/*
host=108.59.87.136
user-agent=curl/7.46.0
BODY:
-no body in request-

$ kubectl patch svc http-svc -p '{"spec":{"type": "NodePort"}}'
"http-svc" patched
```
