http = require "resty.http"
def_backend = "upstream-default-backend"

local concat = table.concat
local upstream = require "ngx.upstream"
local get_servers = upstream.get_servers
local get_upstreams = upstream.get_upstreams
local random = math.random
local us = get_upstreams()

function openURL(original_headers, status)
    local httpc = http.new()

    original_headers["X-Code"] = status or "404"
    original_headers["X-Format"] = original_headers["Accept"] or "text/html"

    local random_backend = get_destination()
    local res, err = httpc:request_uri(random_backend, {
        path = "/",
        method = "GET",
        headers = original_headers,
    })

    if not res then
        ngx.log(ngx.ERR, err)
        ngx.exit(500)
    end

    for k,v in pairs(res.headers) do
        ngx.header[k] = v
    end

    ngx.status = tonumber(status)
    ngx.say(res.body)
end

function get_destination()
    for _, u in ipairs(us) do
        if u == def_backend then
            local srvs, err = get_servers(u)
            local us_table = {}
            if not srvs then
                return "http://127.0.0.1:8181"
            else
                for _, srv in ipairs(srvs) do
                    us_table[srv["name"]] = srv["weight"]
                end
            end
            local destination = random_weight(us_table)
            return "http://"..destination
        end
    end
end

function random_weight(tbl)
    local total = 0
    for k, v in pairs(tbl) do
        total = total + v
    end
    local offset = random(0, total - 1)
    for k1, v1 in pairs(tbl) do
        if offset < v1 then
            return k1
        end
        offset = offset - v1
    end
end
