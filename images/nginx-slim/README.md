
nginx 1.9.x base image using alpine linux

nginx [engine x] is an HTTP and reverse proxy server, a mail proxy server, and a generic TCP proxy server.

This custom nginx image contains:
- http2 instead of spdy
- [lua](https://github.com/openresty/lua-nginx-module) support
- [stream](http://nginx.org/en/docs/stream/ngx_stream_core_module.html) tcp support for upstreams
- nginx stats [nginx-module-vts](https://github.com/vozlt/nginx-module-vts)


**How to use this image:**
This image does provides a default configuration file with no backend servers.

*Using docker*
```
$ docker run -v /some/nginx.con:/etc/nginx/nginx.conf:ro gcr.io/google_containers/nginx-slim:0.3
```

*Creating a replication controller*
```
$ kubectl create -f ./rc.yaml
```
