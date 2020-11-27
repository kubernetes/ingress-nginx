# Exposing TCP and UDP services

Ingress does not support TCP or UDP services. For this reason this Ingress controller uses the flags `--tcp-services-configmap` and `--udp-services-configmap` to point to an existing config map where the key is the external port to use and the value indicates the service to expose using the format:
`[PROTOCOL://]<namespace/service name>:<service port>:[PROXY]:[PROXY][?option1=val1&option2=val2]`

It is also possible to use a number or the name of the port. The two last fields are optional.
Adding `PROXY` in either or both of the two last fields we can use [Proxy Protocol](https://www.nginx.com/resources/admin-guide/proxy-protocol) decoding (listen) and/or encoding (proxy_pass) in a TCP service

Adding `options` like url query params to specify extra functionalities, current support options:
* [Custom NGINX upstream hashing](./nginx-configuration/annotations.md#custom-nginx-upstream-hashing), hash key can use:
    * [NGINX stream module Embedded Variables](http://nginx.org/en/docs/stream/ngx_stream_core_module.html#variables), like `remote_port` will be very useful in balancer decision
    * Specific PROTOCOL variable, now support MQTT protocol, see [MQTT Variable](#MQTT)

The next example shows how to expose the service `example-go` running in the namespace `default` in the port `8080` using the port `9000`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tcp-services
  namespace: ingress-nginx
data:
  9000: "default/example-go:8080"
```

The next example shows how to expose the service `mqtt-service` running in the namespace `default` in the port `1883` using the port `51883` with upstream hash by client id

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: tcp-services
  namespace: ingress-nginx
data:
  "51883": "MQTT://default/example-go:1883?upstream-hash-by=$mqtt_client_id"
```

Since 1.9.13 NGINX provides [UDP Load Balancing](https://www.nginx.com/blog/announcing-udp-load-balancing/).
The next example shows how to expose the service `kube-dns` running in the namespace `kube-system` in the port `53` using the port `53`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: udp-services
  namespace: ingress-nginx
data:
  53: "kube-system/kube-dns:53"
```

If TCP/UDP proxy support is used, then those ports need to be exposed in the Service defined for the Ingress.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: ingress-nginx
  namespace: ingress-nginx
  labels:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
spec:
  type: LoadBalancer
  ports:
    - name: http
      port: 80
      targetPort: 80
      protocol: TCP
    - name: https
      port: 443
      targetPort: 443
      protocol: TCP
    - name: proxied-tcp-9000
      port: 9000
      targetPort: 9000
      protocol: TCP
  selector:
    app.kubernetes.io/name: ingress-nginx
    app.kubernetes.io/part-of: ingress-nginx
```

# Specific Protocol Variable

## MQTT
Support Version: 3.1, 3.1.1
* `mqtt_protocol_name`
* `mqtt_protocol_version`
* `mqtt_client_id`
* `mqtt_user_name`
* `mqtt_password`
