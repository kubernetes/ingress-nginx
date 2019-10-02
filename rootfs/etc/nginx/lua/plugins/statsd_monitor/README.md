# ingress-nginx Lua StatsD based monitoring plugin

When enabled this plugin will send following metrics to the configured StatsD endpoint:

```
nginx.upstream.response
nginx.upstream.response_time
nginx.client.response
nginx.client.request_time
```

Befor enabling the plugin make sure you have `STATSD_HOST` and `STATSD_PORT` environment variables set.
