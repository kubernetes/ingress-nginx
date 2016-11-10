The next command shows the defaults:
```
$ ./nginx-third-party-lb --dump-nginx—configuration
Example of ConfigMap to customize NGINX configuration:
data:
  body-size: 1m
  error-log-level: info
  gzip-types: application/atom+xml application/javascript application/json application/rss+xml
    application/vnd.ms-fontobject application/x-font-ttf application/x-web-app-manifest+json
    application/xhtml+xml application/xml font/opentype image/svg+xml image/x-icon
    text/css text/plain text/x-component
  hts-include-subdomains: "true"
  hts-max-age: "15724800"
  keep-alive: "75"
  max-worker-connections: "16384"
  proxy-connect-timeout: "30"
  proxy-read-timeout: "30"
  proxy-real-ip-cidr: 0.0.0.0/0
  proxy-send-timeout: "30"
  server-name-hash-bucket-size: "64"
  server-name-hash-max-size: "512"
  ssl-buffer-size: 4k
  ssl-ciphers: ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:DHE-DSS-AES128-GCM-SHA256:kEDH+AESGCM:ECDHE-RSA-AES128-SHA256:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA:ECDHE-ECDSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES128-SHA:DHE-DSS-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-DSS-AES256-SHA:DHE-RSA-AES256-SHA:AES128-GCM-SHA256:AES256-GCM-SHA384:AES128-SHA256:AES256-SHA256:AES128-SHA:AES256-SHA:AES:CAMELLIA:DES-CBC3-SHA:!aNULL:!eNULL:!EXPORT:!DES:!RC4:!MD5:!PSK:!aECDH:!EDH-DSS-DES-CBC3-SHA:!EDH-RSA-DES-CBC3-SHA:!KRB5-DES-CBC3-SHA
  ssl-protocols: TLSv1 TLSv1.1 TLSv1.2
  ssl-session-cache: "true"
  ssl-session-cache-size: 10m
  ssl-session-tickets: "true"
  ssl-session-timeout: 10m
  use-gzip: "true"
  use-hts: "true"
  worker-processes: "8"
metadata:
  name: custom-name
  namespace: a-valid-namespace
```

For instance, if we want to change the timeouts we need to create a ConfigMap:
```
$ cat nginx-load-balancer-conf.yaml
apiVersion: v1
data:
  proxy-connect-timeout: "10"
  proxy-read-timeout: "120"
  proxy-send-imeout: "120"
kind: ConfigMap
metadata:
  name: nginx-load-balancer-conf

```

```
$ kubectl create -f nginx-load-balancer-conf.yaml
```

Please check the example `rc-custom-configuration.yaml`

If the Configmap it is updated, NGINX will be reloaded with the new configuration
