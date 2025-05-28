# Miscellaneous

## Source IP address

By default NGINX uses the content of the header `X-Forwarded-For` as the source of truth to get information about the client IP address. This works without issues in L7 **if we configure the setting `proxy-real-ip-cidr`** with the correct information of the IP/network address of trusted external load balancer.

This setting can be enabled/disabled by setting [`use-forwarded-headers`](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#use-forwarded-headers).

If the ingress controller is running in AWS we need to use the VPC IPv4 CIDR.

Another option is to enable the **PROXY protocol** using [`use-proxy-protocol: "true"`](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#use-proxy-protocol).

In this mode, NGINX uses the PROXY protocol TCP header to retrieve the source IP address of the connection.

This works in most cases, but if you have a Layer 7 proxy (e.g., Cloudflare) in front of a TCP load balancer, it may not work correctly. The HTTP proxy IP address might appear as the client IP address. In this case, you should also enable the `use-forwarded-headers` setting in addition to enabling `use-proxy-protocol`, and properly configure `proxy-real-ip-cidr` to trust all intermediate proxies (both within the private network and any external proxies).

Example configmap for setups with multiple proxies:

```yaml
use-proxy-protocol: "true"
use-forwarded-headers: "true"
proxy-real-ip-cidr: "10.0.0.0/8,131.0.72.0/22,172.64.0.0/13,104.24.0.0/14,104.16.0.0/13,162.158.0.0/15,198.41.128.0/17"
```

**Note:** Be sure to use real CIDRs that match your exact environment.

## Path types

Each path in an Ingress is required to have a corresponding path type. Paths that do not include an explicit pathType will fail validation.
By default NGINX path type is Prefix to not break existing definitions

## Proxy Protocol

If you are using a L4 proxy to forward the traffic to the Ingress NGINX pods and terminate HTTP/HTTPS there, you will lose the remote endpoint's IP address. To prevent this you could use the [PROXY Protocol](http://www.haproxy.org/download/1.5/doc/proxy-protocol.txt) for forwarding traffic, this will send the connection details before forwarding the actual TCP connection itself.

Amongst others [ELBs in AWS](http://docs.aws.amazon.com/ElasticLoadBalancing/latest/DeveloperGuide/enable-proxy-protocol.html) and [HAProxy](http://www.haproxy.org/) support Proxy Protocol.

## Websockets

Support for websockets is provided by NGINX out of the box. No special configuration required.

The only requirement to avoid the close of connections is the increase of the values of `proxy-read-timeout` and `proxy-send-timeout`.

The default value of these settings is `60 seconds`.

A more adequate value to support websockets is a value higher than one hour (`3600`).

!!! Important
    If the Ingress-Nginx Controller is exposed with a service `type=LoadBalancer` make sure the protocol between the loadbalancer and NGINX is TCP.

## Optimizing TLS Time To First Byte (TTTFB)

NGINX provides the configuration option [ssl_buffer_size](https://nginx.org/en/docs/http/ngx_http_ssl_module.html#ssl_buffer_size) to allow the optimization of the TLS record size.

This improves the [TLS Time To First Byte](https://www.igvita.com/2013/12/16/optimizing-nginx-tls-time-to-first-byte/) (TTTFB).
The default value in the Ingress controller is `4k` (NGINX default is `16k`).

## Retries in non-idempotent methods

Since 1.9.13 NGINX will not retry non-idempotent requests (POST, LOCK, PATCH) in case of an error.
The previous behavior can be restored using `retry-non-idempotent=true` in the configuration ConfigMap.

## Limitations

- Ingress rules for TLS require the definition of the field `host`

## Why endpoints and not services

The Ingress-Nginx Controller does not use [Services](http://kubernetes.io/docs/user-guide/services) to route traffic to the pods. Instead it uses the Endpoints API in order to bypass [kube-proxy](http://kubernetes.io/docs/admin/kube-proxy/) to allow NGINX features like session affinity and custom load balancing algorithms. It also removes some overhead, such as conntrack entries for iptables DNAT.
