
nginx 1.13.x base image using [ubuntu-slim](https://github.com/kubernetes/ingress-nginx/tree/master/images/ubuntu-slim)

nginx [engine x] is an HTTP and reverse proxy server, a mail proxy server, and a generic TCP proxy server.

This custom nginx image contains:

- [stream](http://nginx.org/en/docs/stream/ngx_stream_core_module.html) tcp support for upstreams
- nginx stats [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)
- [Dynamic TLS record sizing](https://blog.cloudflare.com/optimizing-tls-over-tcp-to-reduce-latency/)

**How to use this image:**
This image does provides a default configuration file with no backend servers.

*Using docker*

```console
docker run -v /some/nginx.con:/etc/nginx/nginx.conf:ro quay.io/kubernetes-ingress-controller/nginx-slim:0.28
```

*Creating a replication controller*

```console
kubectl create -f ./rc.yaml
```
