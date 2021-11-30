# modify monitor.lua
all metrics in monitor.lua

```
local function metrics()
  return {
    host = ngx.var.host or "-",
    namespace = ngx.var.namespace or "-",
    ingress = ngx.var.ingress_name or "-",
    service = ngx.var.service_name or "-",
    canary = ngx.var.proxy_alternative_upstream_name or "-",
    path = ngx.var.location_path or "-",

    method = ngx.var.request_method or "-",
    status = ngx.var.status or "-",
    requestLength = tonumber(ngx.var.request_length) or -1,
    requestTime = tonumber(ngx.var.request_time) or -1,
    responseLength = tonumber(ngx.var.bytes_sent) or -1,

    upstreamLatency = tonumber(ngx.var.upstream_connect_time) or -1,
    upstreamResponseTime = tonumber(ngx.var.upstream_response_time) or -1,
    upstreamResponseLength = tonumber(ngx.var.upstream_response_length) or -1,
    --upstreamStatus = ngx.var.upstream_status or "-",
  }
end
```
for example, some times we do not need upstreamLatency,
we can change "upstreamLatency = tonumber(ngx.var.upstream_connect_time) or -1," to "--upstreamLatency = -1,"

# create a configmap

get monitor.lua from repository,create a configmap

```
kubectl create configmap monitor-lua --from-file=monitor.lua -n ingress-nginx
```

# mount configmap

modify your deployment of ingress-nginx as below!

```
          volumeMounts:
            - name: webhook-cert
              mountPath: /usr/local/certificates/
              readOnly: true
            - name: monitor-lua-volume
              mountPath: /etc/nginx/lua/monitor.lua
              readOnly: true
              subPath: monitor.lua
      nodeSelector:
        kubernetes.io/os: linux
        ingressNginx: ingressNginx
      serviceAccountName: ingress-nginx
      terminationGracePeriodSeconds: 300
      volumes:
        - name: webhook-cert
          secret:
            secretName: ingress-nginx-admission
        - name: monitor-lua-volume
          configMap:
            name: monitor-lua
            items:
            - key: monitor.lua
              path: monitor.lua
```

## restart ingress nginx controller
