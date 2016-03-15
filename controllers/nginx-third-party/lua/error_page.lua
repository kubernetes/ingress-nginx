http = require "resty.http"

function openURL(status, page)
    local httpc = http.new()

    local res, err = httpc:request_uri(page, {
        path = "/",
        method = "GET"
    })

    if not res then
        ngx.log(ngx.ERR, err)
        ngx.exit(500)
    end

    ngx.status = tonumber(status)
    ngx.header["Content-Type"] = ngx.var.httpReturnType or "text/plain"
    if ngx.var.http_cookie then
        ngx.header["Cookie"] = ngx.var.http_cookie
    end
    
    ngx.say(res.body)
end
