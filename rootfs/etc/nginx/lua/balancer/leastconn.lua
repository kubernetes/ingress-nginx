local util = require("util")
local split = require("util.split")
require("resty.core")

local ngx = ngx
local ipairs = ipairs
local tostring = tostring
local string = string
local tonumber = tonumber
local setmetatable = setmetatable
local string_format = string.format
local ngx_log = ngx.log
local INFO = ngx.INFO
local WARN = ngx.WARN


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
        local lowestconns = 9999
        -- find the lowest connection count
        for _, peer in pairs(peers) do
            local conns = endpoints:get(get_upstream_name(peer))
            if conns == nil then
              endpoints:set(get_upstream_name(peer),0,600)
              conns = 0
            end
            ngx_log(WARN, "Found ", conns, " conns for peer ", get_upstream_name(peer))
            if conns <= lowestconns then
                lowestconns = conns
            end
        end

        -- get peers with lowest connections
        for _, peer in pairs(peers) do
            local conns = endpoints:get(get_upstream_name(peer))
            if conns ~= nil and conns == lowestconns then
                feasible_endpoints[#feasible_endpoints+1] = peer
            end
        end
        ngx_log(WARN, "got ", #feasible_endpoints, " feasible endpoints")

        endpoint = feasible_endpoints[math.random(1,#feasible_endpoints)]
    end


    ngx_log(WARN, "chose endpoint ", get_upstream_name(endpoint))
    -- Update the endpoint connection count with a TTL of 10 minutes
    endpoints:incr(get_upstream_name(endpoint),1,1,600)

    return get_upstream_name(endpoint)
end

function _M.after_balance(_)
    local endpoints = ngx.shared.balancer_leastconn
    local upstream = split.get_last_value(ngx.var.upstream_addr)

    ngx_log(WARN, "decrement conn count for upstream ", upstream)

    if util.is_blank(upstream) then
        return
    end
    ngx_log(WARN, "decrement endpoints", upstream)
    ngx_log(WARN, endpoints:incr(upstream,-1,0,600))
end

function _M.sync(self, backend)
    local normalized_endpoints_added, normalized_endpoints_removed =
        util.diff_endpoints(self.peers, backend.endpoints)

    if #normalized_endpoints_added == 0 and #normalized_endpoints_removed == 0 then
        ngx_log(WARN, "endpoints did not change for backend " .. tostring(backend.name))
        return
    end

    ngx_log(WARN, string_format("[%s] peers have changed for backend %s", self.name, backend.name))

    self.peers = backend.endpoints

    for _, endpoint_string in ipairs(normalized_endpoints_removed) do
        ngx.shared.balancer_leastconn:delete(endpoint_string)
    end

    for _, endpoint_string in ipairs(normalized_endpoints_added) do
        ngx.shared.balancer_leastconn:set(endpoint_string,0,600)
    end

end

return _M