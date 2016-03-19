http = require "resty.http"
def_backend = "http://upstream-default-backend"

function openURL(status)
    local httpc = http.new()

    local res, err = httpc:request_uri(def_backend, {
        path = "/",
        method = "GET",
        headers = {
          ["Content-Type"] = ngx.var.httpAccept or "html",
        }
    })

    if not res then
        ngx.log(ngx.ERR, err)
        ngx.exit(500)
    end

    ngx.status = tonumber(status)
    if ngx.var.http_cookie then
        ngx.header["Cookie"] = ngx.var.http_cookie
    end
    
    ngx.say(res.body)
end
