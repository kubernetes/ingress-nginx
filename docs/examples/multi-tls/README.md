# Multi TLS certificate termination

This example uses 2 different certificates to terminate SSL for 2 hostnames.

1. Deploy the controller by creating the rc in the parent dir
2. Create tls secrets for foo.bar.com and bar.baz.com as indicated in the yaml
3. Create [multi-tls.yaml](multi-tls.yaml)

This should generate a segment like:
```console
$ kubectl exec -it nginx-ingress-controller-6vwd1 -- cat /etc/nginx/nginx.conf | grep "foo.bar.com" -B 7 -A 35
    server {
        listen 80;
        listen 443 ssl http2;
        ssl_certificate /etc/nginx-ssl/default-foobar.pem;
        ssl_certificate_key /etc/nginx-ssl/default-foobar.pem;


        server_name foo.bar.com;


        if ($scheme = http) {
            return 301 https://$host$request_uri;
        }



        location / {
            proxy_set_header Host                   $host;

            # Pass Real IP
            proxy_set_header X-Real-IP              $remote_addr;

            # Allow websocket connections
            proxy_set_header                        Upgrade           $http_upgrade;
            proxy_set_header                        Connection        $connection_upgrade;

            proxy_set_header X-Forwarded-For        $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Host       $host;
            proxy_set_header X-Forwarded-Proto      $pass_access_scheme;

            proxy_connect_timeout                   5s;
            proxy_send_timeout                      60s;
            proxy_read_timeout                      60s;

            proxy_redirect                          off;
            proxy_buffering                         off;

            proxy_http_version                      1.1;

            proxy_pass http://default-http-svc-80;
        }
```

And you should be able to reach your nginx service or http-svc service using a hostname switch:
```console
$  kubectl get ing
NAME      RULE          BACKEND   ADDRESS                         AGE
foo-tls   -                       104.154.30.67                   13m
          foo.bar.com
          /             http-svc:80
          bar.baz.com
          /             nginx:80

$ curl https://104.154.30.67 -H 'Host:foo.bar.com' -k
CLIENT VALUES:
client_address=10.245.0.6
command=GET
real path=/
query=nil
request_version=1.1
request_uri=http://foo.bar.com:8080/

SERVER VALUES:
server_version=nginx: 1.9.11 - lua: 10001

HEADERS RECEIVED:
accept=*/*
connection=close
host=foo.bar.com
user-agent=curl/7.35.0
x-forwarded-for=10.245.0.1
x-forwarded-host=foo.bar.com
x-forwarded-proto=https

$ curl https://104.154.30.67 -H 'Host:bar.baz.com' -k
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx on Debian!</title>

$ curl 104.154.30.67
default backend - 404
```
