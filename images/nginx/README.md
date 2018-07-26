
nginx 1.15.x base image using [debian-base](https://github.com/kubernetes/kubernetes/tree/master/build/debian-base)

nginx [engine x] is an HTTP and reverse proxy server, a mail proxy server, and a generic TCP proxy server.

This custom nginx image contains:

- [stream](http://nginx.org/en/docs/stream/ngx_stream_core_module.html) tcp support for upstreams
- [Dynamic TLS record sizing](https://blog.cloudflare.com/optimizing-tls-over-tcp-to-reduce-latency/)
- [ngx_devel_kit](https://github.com/simpl/ngx_devel_kit)
- [set-misc-nginx-module](https://github.com/openresty/set-misc-nginx-module)
- [headers-more-nginx-module](https://github.com/openresty/headers-more-nginx-module)
- [nginx-sticky-module-ng](https://bitbucket.org/nginx-goodies/nginx-sticky-module-ng)
- [nginx-http-auth-digest](https://github.com/atomx/nginx-http-auth-digest)
- [ngx_http_substitutions_filter_module](https://github.com/yaoweibin/ngx_http_substitutions_filter_module)
- [nginx-opentracing](https://github.com/opentracing-contrib/nginx-opentracing)
- [opentracing-cpp](https://github.com/opentracing/opentracing-cpp)
- [zipkin-cpp-opentracing](https://github.com/rnburn/zipkin-cpp-opentracing)
- [ModSecurity-nginx](https://github.com/SpiderLabs/ModSecurity-nginx) (only supported in x86_64)
- [brotli](https://github.com/google/brotli) (not supported in s390x)

**How to use this image:**
This image provides a default configuration file with no backend servers.

*Using docker*

```console
docker run -v /some/nginx.con:/etc/nginx/nginx.conf:ro quay.io/kubernetes-ingress-controller/nginx:0.56
```

*Creating a replication controller*

```console
kubectl create -f ./rc.yaml
```
