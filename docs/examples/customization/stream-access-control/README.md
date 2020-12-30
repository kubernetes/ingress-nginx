# Stream Access Control
## Description
[Nginx stream access module](http://nginx.org/en/docs/stream/ngx_stream_access_module.html) supports limiting access to
certain client addresses. We can set Nginx stream server `allow` and `deny` directives by setting 
`allowlist-source-range` and `denylist-source-range` in Kubernetes Service annotation.
* `allowlist-source-range` is a comma separated values allowed for 
[`allow` directive](http://nginx.org/en/docs/stream/ngx_stream_access_module.html).
* `denylist-source-range` is a comma separated values allowed for 
[`deny` directive](http://nginx.org/en/docs/stream/ngx_stream_access_module.html).

In order to support updating Nginx stream server `allow` and `deny` directives, we can:
* update `allowlist-source-range` or `denylist-source-range` Kubernetes Service annotation
* update [`main-snippet`](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/configmap/#main-snippet) 
with a valid Nginx comment line, for example: `# redis-source-range-change-timestamp TIMESTAMP`, to trigger a Nginx reload.

## Test
### Creation
* Create a Redis Deployment with [redis-deployment.yaml](./redis-deployment.yaml)
* Create a Kubernetes Service for Redis Deployment with [redis-service.yaml](./redis-service.yaml)
* Check if the `deny` and `allow` directives are set in nginx.conf
### Update
* Update redis Kubernetes Service `allowlist-source-range` annotation with `127.0.0.1`
* Update redis Kubernetes Service `denylist-source-range` annotation with `127.0.0.2`
* Check if the `deny` and `allow` directives are updated in nginx.conf