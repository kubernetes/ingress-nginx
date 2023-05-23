local util = require("util")
local split = require("util.split")
require("resty.core")

local ngx = ngx
local ipairs = ipairs
local setmetatable = setmetatable
local string_format = string.format
local ngx_log = ngx.log
local DEBUG = ngx.DEBUG


local _M = { name = "leastconn" }

function _M.new(self, backend)
    local o = {
        peers = backend.endpoints
    }
    setmetatable(o, self)
    self.__index = self
    return o
end

function _M.is_affinitized()
    return false
end


local function get_upstream_name(upstream)
    return upstream.address .. ":" .. upstream.port
end


function _M.balance(self)
    local peers = self.peers
    local endpoint = peers[1]
    local endpoints = ngx.shared.balancer_leastconn
    local feasible_endpoints = {}

    if #peers ~= 1 then
        local lowestconns = 2147483647
        -- find the lowest connection count
        for _, peer in pairs(peers) do
            local peer_name = get_upstream_name(peer)
            local peer_conn_count = endpoints:get(peer_name)
            if peer_conn_count  == nil then
                -- Peer has never been recorded as having connections - add it to the connection
                -- tracking table and the list of feasible peers
                endpoints:set(peer_name,0,0)
                lowestconns = 0
                feasible_endpoints[#feasible_endpoints+1] = peer
            elseif peer_conn_count < lowestconns then
                -- Peer has fewer connections than any other peer evaluated so far - add it as the
                -- only feasible endpoint for now
                feasible_endpoints = {peer}
                lowestconns = peer_conn_count
            elseif peer_conn_count == lowestconns then
                -- Peer has equal fewest connections as other peers - add it to the list of
                -- feasible peers
                feasible_endpoints[#feasible_endpoints+1] = peer
            end
        end
        ngx_log(DEBUG, "select from ", #feasible_endpoints, " feasible endpoints out of ", #peers)
        endpoint = feasible_endpoints[math.random(1,#feasible_endpoints)]
    end

    local selected_endpoint = get_upstream_name(endpoint)
    ngx_log(DEBUG, "selected endpoint ", selected_endpoint)

    -- Update the endpoint connection count
    endpoints:incr(selected_endpoint,1,1,0)

    return selected_endpoint
end

function _M.after_balance(_)
    local endpoints = ngx.shared.balancer_leastconn
    local upstream = split.get_last_value(ngx.var.upstream_addr)

    if util.is_blank(upstream) then
        return
    end
    endpoints:incr(upstream,-1,0,0)
end

function _M.sync(self, backend)
    local normalized_endpoints_added, normalized_endpoints_removed =
        util.diff_endpoints(self.peers, backend.endpoints)

    if #normalized_endpoints_added == 0 and #normalized_endpoints_removed == 0 then
        return
    end

    ngx_log(DEBUG, string_format("[%s] peers have changed for backend %s", self.name, backend.name))

    self.peers = backend.endpoints

    for _, endpoint_string in ipairs(normalized_endpoints_removed) do
        ngx.shared.balancer_leastconn:delete(endpoint_string)
    end

    for _, endpoint_string in ipairs(normalized_endpoints_added) do
        ngx.shared.balancer_leastconn:set(endpoint_string,0,0)
    end

end

return _M
