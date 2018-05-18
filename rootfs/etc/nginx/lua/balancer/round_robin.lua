local balancer_resty = require("balancer.resty")
local resty_roundrobin = require("resty.roundrobin")
local util = require("util")

local _M = balancer_resty:new({ factory = resty_roundrobin, name = "round_robin" })

function _M.new(self, backend)
  local nodes = util.get_nodes(backend.endpoints)
  local o = { instance = self.factory:new(nodes) }
  setmetatable(o, self)
  self.__index = self
  return o
end

function _M.balance(self)
  local endpoint_string = self.instance:find()
  return util.split_pair(endpoint_string, ":")
end

return _M
