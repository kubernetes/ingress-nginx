NGINX base image using [alpine](https://www.alpinelinux.org/)

This custom image contains:

- [nginx-http-auth-digest](https://github.com/atomx/nginx-http-auth-digest)
- [ngx_http_substitutions_filter_module](https://github.com/yaoweibin/ngx_http_substitutions_filter_module)
- [OpenTelemetry-CPP](https://github.com/open-telemetry/opentelemetry-cpp)
- [OpenTelemetry-CPP-Nginx](https://github.com/open-telemetry/opentelemetry-cpp-contrib/tree/main/instrumentation/nginx)
- [nginx-opentracing](https://github.com/opentracing-contrib/nginx-opentracing)
- [opentracing-cpp](https://github.com/opentracing/opentracing-cpp)
- [zipkin-cpp-opentracing](https://github.com/rnburn/zipkin-cpp-opentracing)
- [dd-opentracing-cpp](https://github.com/DataDog/dd-opentracing-cpp)
- [ModSecurity-nginx](https://github.com/SpiderLabs/ModSecurity-nginx) (only supported in x86_64)
- [brotli](https://github.com/google/brotli)
- [geoip2](https://github.com/leev/ngx_http_geoip2_module)

**How to use this image:**
This image provides a default configuration file with no backend servers.

_Using docker_

```console
docker run -v /some/nginx.conf:/etc/nginx/nginx.conf:ro k8s.gcr.io/ingress-nginx/nginx:81c2afd975a6f9a9847184472286044d7d5296f6@sha256:a71ac64dd8cfd68341ba47dbdc4d8c2cb91325fce669875193ea0319118201b5
```

