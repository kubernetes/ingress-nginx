# Log format

The default configuration uses a custom logging format to add additional information about upstreams, response time and status.

```
log_format upstreaminfo
    '$remote_addr - $remote_user [$time_local] "$request" '
    '$status $body_bytes_sent "$http_referer" "$http_user_agent" '
    '$request_length $request_time [$proxy_upstream_name] [$proxy_alternative_upstream_name] $upstream_addr '
    '$upstream_response_length $upstream_response_time $upstream_status $req_id';
```

| Placeholder | Description |
|-------------|-------------|
| `$proxy_protocol_addr` | remote address if proxy protocol is enabled |
| `$remote_addr` | the source IP address of the client |
| `$remote_user` | user name supplied with the Basic authentication |
| `$time_local` | local time in the Common Log Format |
| `$request` | full original request line |
| `$status` | response status |
| `$body_bytes_sent` | number of bytes sent to a client, not counting the response header |
| `$http_referer` | value of the Referer header |
| `$http_user_agent` | value of User-Agent header |
| `$request_length` | request length (including request line, header, and request body) |
| `$request_time` | time elapsed since the first bytes were read from the client |
| `$proxy_upstream_name` | name of the upstream. The format is `upstream-<namespace>-<service name>-<service port>` |
| `$proxy_alternative_upstream_name` | name of the alternative upstream. The format is `upstream-<namespace>-<service name>-<service port>` |
| `$upstream_addr` | the IP address and port (or the path to the domain socket) of the upstream server. If several servers were contacted during request processing, their addresses are separated by commas. |
| `$upstream_response_length` | the length of the response obtained from the upstream server |
| `$upstream_response_time` | time spent on receiving the response from the upstream server as seconds with millisecond resolution |
| `$upstream_status` | status code of the response obtained from the upstream server |
| `$req_id` | value of the `X-Request-ID` HTTP header. If the header is not set, a randomly generated ID. |

Additional available variables:

| Placeholder | Description |
|-------------|-------------|
| `$namespace` |  namespace of the ingress |
| `$ingress_name` | name of the ingress |
| `$service_name` | name of the service |
| `$service_port` | port of the service |


Sources:

- [Upstream variables](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#variables)
- [Embedded variables](http://nginx.org/en/docs/http/ngx_http_core_module.html#variables)
