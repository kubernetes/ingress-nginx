local json = require "json"
local b = require "ngx.balancer"

local http_host = ngx.var.host
local request_uri = ngx.var.request_uri
local shared_memory = ngx.shared.shared_memory;
local vhosts_json = shared_memory:get("VHOSTS")
local vhosts = json.decode(json.decode(vhosts_json))

local server = vhosts.servers[http_host]
if (server == nil) then
    server = vhosts.servers["_"]
    if (server == nil) then
        ngx.status = 503
        ngx.exit(ngx.status)
    end
end

local location
local hit_length = 0
for k, v in pairs(server.locations) do
    local path_length = string.len(k)
    if string.sub(request_uri,1, path_length)==k then
        if path_length > hit_length then
            hit_length = path_length
            location = server.locations[k]
        end
    end
end

if (location == nil) then
    ngx.status = 404
    ngx.exit(ngx.status)
end

if (location.endpoints == nil) then
    ngx.status = 404
    ngx.exit(ngx.status)
end

if (location.endpoints[1] == nil) then
    ngx.status = 404
    ngx.exit(ngx.status)
end

local selected_endpoint
local endpoints_roundrobin = ngx.shared.endpoints_roundrobin;
local ep_index = endpoints_roundrobin:get(http_host)
if ep_index == nil then
    selected_endpoint = location.endpoints[1]
    endpoints_roundrobin:set(http_host, 1, 600)
else
    local new_index = ep_index + 1
    if location.endpoints[new_index] == nil then
        selected_endpoint = location.endpoints[1]
        endpoints_roundrobin:set(http_host, 1, 600)
    else
        selected_endpoint = location.endpoints[new_index]
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
