local json = require "json"
local b = require "ngx.balancer"

local function compare(t, fn)
    if #t == 0 then return nil, nil end
    local key, value = 1, t[1]
    for i = 2, #t do
        if fn(value, t[i]) then
            key, value = i, t[i]
        end
    end
    return key, value
end

local http_host = ngx.var.http_host
local remote_addr = ngx.var.remote_addr
local proxy_upstream_name = ngx.var.proxy_upstream_name
local shared_memory = ngx.shared.shared_memory;
local cfgs_json = shared_memory:get("CFGS")
if cfgs_json == nil then
    ngx.status = 503
    ngx.exit(ngx.status)
end
local cfgs = json.decode(cfgs_json)
local loadbalancealgorithm = cfgs.cfg.loadbalancealgorithm
local backend = cfgs.backends[proxy_upstream_name]
if backend == nil then
    ngx.status = 503
    ngx.exit(ngx.status)
end
local selected_endpoint
if backend.sessionaffinity.affinitytype == "cookie" then
    local backend_cookie_name = backend.sessionaffinity.cookiesessionaffinity.name
    local backend_cookie_hash = backend.sessionaffinity.cookiesessionaffinity.hash
    local var_name = "cookie_SESSION_" .. backend_cookie_name .. backend_cookie_hash
    
    local cookie_value = ngx.var[var_name]

    if cookie_value == nil then
        local endpoints_roundrobin = ngx.shared.endpoints_roundrobin
        local ep_index = endpoints_roundrobin:get(http_host)
        if ep_index == nil then
            selected_endpoint = backend.endpoints[1]
            endpoints_roundrobin:set(http_host, 1, 600)
        else
            local new_index = ep_index + 1
            if backend.endpoints[new_index] == nil then
                selected_endpoint = backend.endpoints[1]
                endpoints_roundrobin:set(http_host, 1, 600)
            else
                selected_endpoint = backend.endpoints[new_index]
                endpoints_roundrobin:set(http_host, new_index, 600)
            end
        end
        ngx.header['Set-Cookie'] = var_name .. '=' .. tostring(ep_index)
    else
        selected_endpoint = backend.endpoints[tonumber(cookie_value)]
    end
else
    if loadbalancealgorithm == "" or loadbalancealgorithm == "round_robin" then
        local endpoints_roundrobin = ngx.shared.endpoints_roundrobin
        local ep_index = endpoints_roundrobin:get(http_host)
        if ep_index == nil then
            selected_endpoint = backend.endpoints[1]
            endpoints_roundrobin:set(http_host, 1, 600)
        else
            local new_index = ep_index + 1
            if backend.endpoints[new_index] == nil then
                selected_endpoint = backend.endpoints[1]
                endpoints_roundrobin:set(http_host, 1, 600)
            else
                selected_endpoint = backend.endpoints[new_index]
                endpoints_roundrobin:set(http_host, new_index, 600)
            end
        end
    elseif loadbalancealgorithm == "ip_hash" or loadbalancealgorithm == "least_conn" then
        local endpoints_leastconn = ngx.shared.endpoints_leastconn
        local ep_index = endpoints_leastconn:get(remote_addr)
        if ep_index == nil then
            local counter = {}
            for k, v in pairs(backend.endpoints) do
                counter[k] = 0
            end
            local leastconn_rows = endpoints_leastconn:get_keys(0)
            for _, v in pairs(leastconn_rows) do
                counter[v] = counter[v] + 1
            end
            local free_endpoint_id, _ = compare(counter, function(a,b) return a > b end)
            selected_endpoint = backend.endpoints[free_endpoint_id]
            endpoints_leastconn:set(remote_addr, free_endpoint_id, 600)
        else
            selected_endpoint = backend.endpoints[ep_index]
            endpoints_leastconn:set(remote_addr, ep_index, 600)
        end
    else
        ngx.status = 503
        ngx.exit(ngx.status)
    end
end

local max_retries = 20
if selected_endpoint.maxfails ~= 0 then
    max_retries = selected_endpoint.maxfails
end

assert(b.set_current_peer(selected_endpoint.hostname, selected_endpoint.port))
if (selected_endpoint.failtimeout ~= 0) then
    assert(b.set_timeouts(selected_endpoint.failtimeout, selected_endpoint.failtimeout, selected_endpoint.failtimeout))
end
if (selected_endpoint.maxfails ~= 0) then
    assert(b.set_more_tries(selected_endpoint.maxfails))
end
