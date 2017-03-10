This example shows how is possible to create a custom configuration for a particular upstream associated with an Ingress rule.

```
echo "
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: echoheaders
  annotations:
    ingress.kubernetes.io/upstream-fail-timeout: "30"
spec:
  rules:
  - host: foo.bar.com
    http:
      paths:
      - path: /
        backend:
          serviceName: echoheaders
          servicePort: 80
" | kubectl create -f -
```

Check the annotation is present in the Ingress rule:
```
kubectl get ingress echoheaders -o yaml
```

Check the NGINX configuration is updated using kubectl or the status page:

```
$ kubectl exec nginx-ingress-controller-v1ppm cat /etc/nginx/nginx.conf
```

```
....
    upstream default-echoheaders-x-80 {
        least_conn;
        server 10.2.92.2:8080 max_fails=5 fail_timeout=30;

    }
....
```


![nginx-module-vts](custom-upstream.png "screenshot with custom configuration")
