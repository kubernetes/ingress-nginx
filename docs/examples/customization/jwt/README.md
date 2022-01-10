# Accommodation for JWT

JWT (short for Json Web Token) is an authentication method widely used. Basically an authentication server generates
a JWT and you then use this token in every request you make to a backend service. The JWT can be quite big and is
present in every http headers. This means you may have to adapt the max-header size of your nginx-ingress in order
to support it.

## Symptoms

If you use JWT and you get http 502 error from your ingress, it may be a sign that the buffer size is not big enough.

To be 100% sure look at the logs of the `ingress-nginx-controller` pod, you should see something like this:

```
upstream sent too big header while reading response header from upstream...
```


## Increase buffer size for headers

In nginx, we want to modify the property `proxy-buffer-size`. The size is arbitrary. It depends on your needs. Be aware
that a high value can lower the performance of your ingress proxy. In general a value of 16k should get you covered.

### Using helm
If you're using helm you can simply use the [`config` properties](https://github.com/kubernetes/ingress-nginx/blob/main/charts/ingress-nginx/values.yaml#L37).
```yaml
 # -- Will add custom configuration options to Nginx https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/
  config: 
    proxy-buffer-size: 16k
```

## Manually in kubernetes config files

If you use an already generated config from for a provider, you will have to change the `controller-configmap.yaml`

```yaml
---
# Source: ingress-nginx/templates/controller-configmap.yaml
apiVersion: v1
kind: ConfigMap
# ...
data:
  #...
  proxy-buffer-size: "16k"
```

References:
 * [Custom Configuration](../custom-configuration/)