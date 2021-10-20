# Custom DH parameters for perfect forward secrecy

This example aims to demonstrate the deployment of an nginx ingress controller and
use a ConfigMap to configure a custom Diffie-Hellman parameters file to help with
"Perfect Forward Secrecy".

## Custom configuration

```console
$ cat configmap.yaml
apiVersion: v1
data:
  ssl-dh-param: "ingress-nginx/lb-dhparam"
kind: ConfigMap
metadata:
  name: ingress-nginx-controller
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
```

```console
$ kubectl create -f configmap.yaml
```

## Custom DH parameters secret

```console
$ openssl dhparam 4096 2> /dev/null | base64
LS0tLS1CRUdJTiBESCBQQVJBTUVURVJ...
```

```console
$ cat ssl-dh-param.yaml
apiVersion: v1
data:
  dhparam.pem: "LS0tLS1CRUdJTiBESCBQQVJBTUVURVJ..."
kind: Secret
metadata:
  name: lb-dhparam
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
```

```console
$ kubectl create -f ssl-dh-param.yaml
```

## Test

Check the contents of the configmap is present in the nginx.conf file using:
```console
$ kubectl exec ingress-nginx-controller-873061567-4n3k2 -n kube-system -- cat /etc/nginx/nginx.conf
```
