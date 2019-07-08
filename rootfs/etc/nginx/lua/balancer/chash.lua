local balancer_resty = require("balancer.resty")
local resty_chash = require("resty.chash")
local util = require("util")

local _M = balancer_resty:new({ factory = resty_chash, name = "chash" })

function _M.new(self, backend)
  local nodes = util.get_nodes(backend.endpoints)
  local o = {
    instance = self.factory:new(nodes),
    hash_by = backend["upstreamHashByConfig"]["upstream-hash-by"],
    traffic_shaping_policy = backend.trafficShapingPolicy,
    alternative_backends = backend.alternativeBackends,
  }
  setmetatable(o, self)
  self.__index = self
  return o
end

function _M.balance(self)
  local key = util.lua_ngx_var(self.hash_by)
  return self.instance:find(key)
end

return _M
