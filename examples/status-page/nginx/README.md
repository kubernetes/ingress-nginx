Example configuration for configuring enabling Status Page on the NGINX Ingress Controller  
=========

# Create replica set and service definition for the default http backend
```
kubectl create -f default-backend-rc.yaml -f default-backend-svc.yaml
```

# Create NGINX Ingress controller with the desired vts status page enabled option
```
kubectl create -f status-page-configmap.yaml -f status-page-rc.yaml
```

# Expose ports externally
The following example service configuration can be used to make services addressable outside of the Kubernetes cluster

* http endpoint: http://${NODE_IP}:32080
* https endpoint: https://${NODE_IP}:32443
* status endpoint: http://${NODE_IP}:32081

```
kubectl create -f status-page-svc.yaml
```

# Testing

IP 172.17.4.99 used here as an example

## Default endpoint 
### Browser

* http://172.17.4.99:32080/ will give 404 with "default backend - 404" content
* http:///172.17.4.99:32080/healthz will give 200 with "ok" content

### CLI

```
curl -v http://172.17.4.99:32080
* Rebuilt URL to: http://172.17.4.99:32080/
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 32080 (#0)
> GET / HTTP/1.1
> Host: 172.17.4.99:32080
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 404 Not Found
< Server: nginx/1.11.3
< Date: Thu, 22 Dec 2016 10:32:23 GMT
< Content-Type: text/plain; charset=utf-8
< Content-Length: 21
< Connection: keep-alive
< Strict-Transport-Security: max-age=15724800; includeSubDomains; preload
<
* Connection #0 to host 172.17.4.99 left intact

curl -v http://172.17.4.99:32080/healthz
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 32080 (#0)
> GET /healthz HTTP/1.1
> Host: 172.17.4.99:32080
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.11.3
< Date: Thu, 22 Dec 2016 10:31:36 GMT
< Content-Type: text/plain; charset=utf-8
< Content-Length: 2
< Connection: keep-alive
< Strict-Transport-Security: max-age=15724800; includeSubDomains; preload
<
* Connection #0 to host 172.17.4.99 left intact
```

## Status page
### Browser

http://172.17.4.99:32081/nginx_status/

### CLI

```
curl -v http://172.17.4.99:32081/nginx_status/format/json
*   Trying 172.17.4.99...
* Connected to 172.17.4.99 (172.17.4.99) port 32081 (#0)
> GET /nginx_status/format/json HTTP/1.1
> Host: 172.17.4.99:32081
> User-Agent: curl/7.43.0
> Accept: */*
>
< HTTP/1.1 200 OK
< Server: nginx/1.11.3
< Date: Thu, 22 Dec 2016 10:32:57 GMT
< Content-Type: application/json
< Content-Length: 2303
< Connection: keep-alive
<
* Connection #0 to host 172.17.4.99 left intact
{"nginxVersion":"1.11.3","loadMsec":1482401683372,"nowMsec":1482402243310,"connections":{"active":4,"reading":0,"writing":1,"waiting":3,"accepted":19,"handled":19,"requests":629},"serverZones":{"_":{"requestCounter":621,"inBytes":486167,"outBytes":597318,"responses":{"1xx":0,"2xx":620,"3xx":0,"4xx":1,"5xx":0,"miss":0,"bypass":0,"expired":0,"stale":0,"updating":0,"revalidated":0,"hit":0,"scarce":0},"overCounts":{"maxIntegerSize":18446744073709551615,"requestCounter":0,"inBytes":0,"outBytes":0,"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0,"miss":0,"bypass":0,"expired":0,"stale":0,"updating":0,"revalidated":0,"hit":0,"scarce":0}},"testurl":{"requestCounter":7,"inBytes":3558,"outBytes":426201,"responses":{"1xx":0,"2xx":7,"3xx":0,"4xx":0,"5xx":0,"miss":0,"bypass":0,"expired":0,"stale":0,"updating":0,"revalidated":0,"hit":0,"scarce":0},"overCounts":{"maxIntegerSize":18446744073709551615,"requestCounter":0,"inBytes":0,"outBytes":0,"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0,"miss":0,"bypass":0,"expired":0,"stale":0,"updating":0,"revalidated":0,"hit":0,"scarce":0}},"*":{"requestCounter":628,"inBytes":489725,"outBytes":1023519,"responses":{"1xx":0,"2xx":627,"3xx":0,"4xx":1,"5xx":0,"miss":0,"bypass":0,"expired":0,"stale":0,"updating":0,"revalidated":0,"hit":0,"scarce":0},"overCounts":{"maxIntegerSize":18446744073709551615,"requestCounter":0,"inBytes":0,"outBytes":0,"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0,"miss":0,"bypass":0,"expired":0,"stale":0,"updating":0,"revalidated":0,"hit":0,"scarce":0}}},"upstreamZones":{"default-test-ui-8443":[{"server":"10.2.97.6:8443","requestCounter":7,"inBytes":3558,"outBytes":426201,"responses":{"1xx":0,"2xx":7,"3xx":0,"4xx":0,"5xx":0},"responseMsec":297,"weight":1,"maxFails":0,"failTimeout":0,"backup":false,"down":false,"overCounts":{"maxIntegerSize":18446744073709551615,"requestCounter":0,"inBytes":0,"outBytes":0,"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0}}],"upstream-default-backend":[{"server":"10.2.90.238:8080","requestCounter":0,"inBytes":0,"outBytes":0,"responses":{"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0},"responseMsec":0,"weight":1,"maxFails":0,"failTimeout":0,"backup":false,"down":false,"overCounts":{"maxIntegerSize":18446744073709551615,"requestCounter":0,"inBytes":0,"outBytes":0,"1xx":0,"2xx":0,"3xx":0,"4xx":0,"5xx":0}}]}}

```

