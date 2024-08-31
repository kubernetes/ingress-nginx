local request_uri = ngx.var.request_uri
local redirect_to = ngx.arg[1]

local luaconfig = ngx.shared.luaconfig
local use_forwarded_headers = luaconfig:get("use_forwarded_headers")


if string.sub(request_uri, -1) == "/" then
    request_uri = string.sub(request_uri, 1, -2)
end

local redirectScheme = ngx.var.scheme
local redirectPort = ngx.var.server_port

if use_forwarded_headers then
    if ngx.var.http_x_forwarded_proto then
        redirectScheme = ngx.var.http_x_forwarded_proto
    end
    if ngx.var.http_x_forwarded_port then
        redirectPort = ngx.var.http_x_forwarded_port
    end
end


return string.format("%s://%s:%s%s", redirectScheme,
    redirect_to, redirectPort, request_uri)

