local auth_path = ngx.var.auth_path
local auth_keepalive_share_vars = ngx.var.auth_keepalive_share_vars == "true" and true or false
local auth_response_headers = ngx.var.auth_response_headers
local ngx_re_split = require("ngx.re").split
local ipairs = ipairs
local ngx_log = ngx.log
local ngx_ERR = ngx.ERR

local res = ngx.location.capture(auth_path, {
    method = ngx.HTTP_GET, body = '',
    share_all_vars = auth_keepalive_share_vars })

if res.status == ngx.HTTP_OK then
    local header_parts, err = ngx_re_split(auth_response_headers, ",")
    if err then
       ngx_log(ngx_ERR, err)
       return
    end
    ngx.var.auth_cookie = res.header['Set-Cookie']
    for i, header_name in ipairs(header_parts) do
        local varname = "authHeader" .. tostring(i)
        ngx.var[varname] = res.header[header_name]
    end
    return
end

if res.status == ngx.HTTP_UNAUTHORIZED or res.status == ngx.HTTP_FORBIDDEN then
    ngx.exit(res.status)
end
ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
