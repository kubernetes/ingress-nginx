This is an example to use a TLS Ingress rule to use SSL in NGINX

*First expose the `echoheaders` service:*

```
kubectl run echoheaders --image=gcr.io/google_containers/echoserver:1.3 --replicas=1 --port=8080
kubectl expose deployment echoheaders --port=80 --target-port=8080 --name=echoheaders-x
```

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

```
TODO: 
- show logs
- curl
```


##### Another example:

This shows a more complex example that creates the servers `foo.bar.com` and `bar.baz.com` where only `foo.bar.com` uses SSL

```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: complex-foo
  namespace: default
spec:
  tls:
  - hosts:
    - foo.bar.com
    secretName: foo-tls
  - hosts:
    - bar.baz.com
    secretName: foo-tls
  rules:
  - host: foo.bar.com
    http:
      paths:
      - backend:
          serviceName: echoheaders-x
          servicePort: 80
        path: /
  - host: bar.baz.com
    http:
      paths:
      - backend:
          serviceName: echoheaders-y
          servicePort: 80
        path: /
```


```
TODO: 
- show logs
- curl
```
