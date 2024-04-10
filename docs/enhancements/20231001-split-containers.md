# Proposal to split containers

* All the NGINX files should live on one container
  * No file other than NGINX files should exist on this container
  * This includes not mounting the service account
* All the controller files should live on a different container
  * Controller container should have bare minimum to work (just go program)
  * ServiceAccount should be mounted just on controller

* Inside nginx container, there should be a really small http listener just able 
to start, stop and reload NGINX

## Roadmap (what needs to be done)
* Map what needs to be done to mount the SA just on controller container
* Map all the required files for NGINX to work
* Map all the required network calls between controller and NGINX
  * eg.: Dynamic lua reconfiguration
* Map problematic features that will need attention
  * SSLPassthrough today happens on controller process and needs to happen on NGINX

### Ports and endpoints on NGINX container
* Public HTTP/HTTPs port - 80 and 443
* Lua configuration port - 10246 (HTTP) and 10247 (Stream)
* 3333 (temp) - Dataplane controller http server
  * /reload - (POST) Reloads the configuration.
    * "config" argument is the location of temporary file that should be used / moved to nginx.conf
  * /test - (POST) Test the configuration of a given file location
    * "config" argument is the location of temporary file that should be tested

### Mounting empty SA on controller container

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: test
spec:
  containers:
  - name: nginx
    image: nginx:latest
    ports:
    - containerPort: 80
  - name: othernginx
    image: alpine:latest
    command: ["/bin/sh"]
    args: ["-c", "while true; do date; sleep 3; done"]
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: emptysecret
  volumes:
  - name: emptysecret
    emptyDir:
      sizeLimit: 1Mi
```

### Mapped folders on NGINX configuration
**WARNING** We need to be aware of inter mount containers and inode problems. If we 
mount a file instead of a directory, it may take time to reflect the file value on 
the target container

*  "/etc/nginx/lua/?.lua;/etc/nginx/lua/vendor/?.lua;;"; - Lua scripts
* "/var/log/nginx" - NGINX logs
* "/tmp/nginx (nginx.pid)" - NGINX pid directory / file, fcgi socket, etc
* " /etc/nginx/geoip" - GeoIP database directory - OK - /etc/ingress-controller/geoip
* /etc/nginx/mime.types - Mime types
* /etc/ingress-controller/ssl - SSL directory (fake cert, auth cert)
* /etc/ingress-controller/auth - Authentication files
* /etc/nginx/modsecurity - Modsecurity configuration
* /etc/nginx/owasp-modsecurity-crs - Modsecurity rules
* /etc/nginx/tickets.key - SSL tickets - OK - /etc/ingress-controller/tickets.key
* /etc/nginx/opentelemetry.toml - OTEL config - OK - /etc/ingress-controller/telemetry
* /etc/nginx/opentracing.json - Opentracing config - OK - /etc/ingress-controller/telemetry
* /etc/nginx/modules - NGINX modules
* /etc/nginx/fastcgi_params (maybe) - fcgi params
* /etc/nginx/template - Template, may be used by controller only

##### List of modules
```
ngx_http_auth_digest_module.so    ngx_http_modsecurity_module.so
ngx_http_brotli_filter_module.so  ngx_http_opentracing_module.so
ngx_http_brotli_static_module.so  ngx_stream_geoip2_module.so
ngx_http_geoip2_module.so
```

##### List of files that may be removed
```
-rw-r--r--    1 www-data www-data      1077 Jun 23 19:44 fastcgi.conf
-rw-r--r--    1 www-data www-data      1077 Jun 23 19:44 fastcgi.conf.default
-rw-r--r--    1 www-data www-data      1007 Jun 23 19:44 fastcgi_params
-rw-r--r--    1 www-data www-data      1007 Jun 23 19:44 fastcgi_params.default
drwxr-xr-x    2 www-data www-data      4096 Jun 23 19:34 geoip
-rw-r--r--    1 www-data www-data      2837 Jun 23 19:44 koi-utf
-rw-r--r--    1 www-data www-data      2223 Jun 23 19:44 koi-win
drwxr-xr-x    6 www-data www-data      4096 Sep 19 14:13 lua
-rw-r--r--    1 www-data www-data      5349 Jun 23 19:44 mime.types
-rw-r--r--    1 www-data www-data      5349 Jun 23 19:44 mime.types.default
drwxr-xr-x    2 www-data www-data      4096 Jun 23 19:44 modsecurity
drwxr-xr-x    2 www-data www-data      4096 Jun 23 19:44 modules
-rw-r--r--    1 www-data www-data     18275 Oct  1 21:28 nginx.conf
-rw-r--r--    1 www-data www-data      2656 Jun 23 19:44 nginx.conf.default
-rwx------    1 www-data www-data       420 Oct  1 21:28 opentelemetry.toml
-rw-r--r--    1 www-data www-data         2 Oct  1 21:28 opentracing.json
drwxr-xr-x    7 www-data www-data      4096 Jun 23 19:44 owasp-modsecurity-crs
-rw-r--r--    1 www-data www-data       636 Jun 23 19:44 scgi_params
-rw-r--r--    1 www-data www-data       636 Jun 23 19:44 scgi_params.default
drwxr-xr-x    2 www-data www-data      4096 Sep 19 14:13 template
-rw-r--r--    1 www-data www-data       664 Jun 23 19:44 uwsgi_params
-rw-r--r--    1 www-data www-data       664 Jun 23 19:44 uwsgi_params.default
-rw-r--r--    1 www-data www-data      3610 Jun 23 19:44 win-utf
```
