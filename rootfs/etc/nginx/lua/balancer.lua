local json = require "json"
local b = require "ngx.balancer"

local http_host = ngx.var.http_host
local proxy_upstream_name = ngx.var.proxy_upstream_name
local shared_memory = ngx.shared.shared_memory;
local backends_json = shared_memory:get("BACKENDS")
local backends = json.decode(json.decode(backends_json))

local backend = backends[proxy_upstream_name]
local selected_endpoint
local endpoints_roundrobin = ngx.shared.endpoints_roundrobin;
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

local max_retries = 20
if selected_endpoint.maxfails ~= 0 then
    max_retries = selected_endpoint.maxfails
end

assert(b.set_current_peer(selected_endpoint.hostname, selected_endpoint.port))
if (selected_endpoint.failtimeout ~= 0) then
    assert(b.set_timeouts(selected_endpoint.failtimeout, selected_endpoint.failtimeout, selected_endpoint.failtimeout))
end
