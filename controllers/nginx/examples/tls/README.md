This is an example to use a TLS Ingress rule to use SSL in NGINX

# TLS certificate termination

This examples uses 2 different certificates to terminate SSL for 2 hostnames.

1. Deploy the controller by creating the rc in the parent dir
2. Create tls secret for foo.bar.com
3. Create rc-ssl.yaml

*Next create a SSL certificate for `foo.bar.com` host:*

```
openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout /tmp/tls.key -out /tmp/tls.crt -subj "/CN=foo.bar.com"
```

*Now store the SSL certificate in a secret:*

```
echo "
apiVersion: v1
kind: Secret
metadata:
  name: foo-secret
data:
  tls.crt: `base64 /tmp/tls.crt`
  tls.key: `base64 /tmp/tls.key`
" | kubectl create -f -
```

*Finally create a tls Ingress rule:*

```
echo "
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: foo
  namespace: default
spec:
  tls:
  - hosts:
    - foo.bar.com
    secretName: foo-secret
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          serviceName: echoheaders-x
          servicePort: 80
        path: /
" | kubectl create -f -
```

You should be able to reach your nginx service or echoheaders service using a hostname:
```
$ kubectl get ing
NAME      RULE          BACKEND   ADDRESS
foo       -                       10.4.0.3
          foo.bar.com
          /             echoheaders-x:80
```

```
$ curl https://10.4.0.3 -H 'Host:foo.bar.com' -k
old-mbp:contrib aledbf$ curl https://10.4.0.3 -H 'Host:foo.bar.com' -k
CLIENT VALUES:
client_address=10.2.48.4
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://foo.bar.com:8080/

SERVER VALUES:
server_version=nginx: 1.9.7 - lua: 9019

HEADERS RECEIVED:
accept=*/*
connection=close
host=foo.bar.com
user-agent=curl/7.43.0
x-forwarded-for=10.2.48.1
x-forwarded-host=foo.bar.com
x-forwarded-proto=https
x-real-ip=10.2.48.1
BODY:
-no body in request-
```
