
Using a [ConfigMap](https://kubernetes.io/docs/user-guide/configmap/) is possible to customize the NGINX configuration

For example, if we want to change the timeouts we need to create a ConfigMap:

```
$ cat nginx-load-balancer-conf.yaml
apiVersion: v1
data:
  proxy-connect-timeout: "10"
  proxy-read-timeout: "120"
  proxy-send-timeout: "120"
kind: ConfigMap
metadata:
  name: nginx-load-balancer-conf
```

```
$ kubectl create -f nginx-load-balancer-conf.yaml
```

Please check the example `nginx-custom-configuration.yaml`

If the Configmap it is updated, NGINX will be reloaded with the new configuration.
