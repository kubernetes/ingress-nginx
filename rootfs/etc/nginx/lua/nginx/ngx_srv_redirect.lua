local request_uri = ngx.var.request_uri
local redirect_to = ngx.arg[1]

local luaconfig = ngx.shared.luaconfig
local use_forwarded_headers = luaconfig:get("use_forwarded_headers")
local listen_https_ports = luaconfig:get("listen_https_ports")


if string.sub(request_uri, -1) == "/" then
    request_uri = string.sub(request_uri, 1, -2)
end

local redirectScheme

if use_forwarded_headers then
    if not ngx.var.http_x_forwarded_proto then
        redirectScheme = ngx.var.scheme
    else
        redirectScheme = ngx.var.http_x_forwarded_proto
    end
else
    redirectScheme = ngx.var.scheme
end

if listen_https_ports == '443' then
    return string.format("%s://%s%s", redirectScheme, redirect_to, request_uri)
else
    return string.format("%s://%s:%s%s", redirectScheme,
        redirect_to, listen_https_ports, request_uri)
end
