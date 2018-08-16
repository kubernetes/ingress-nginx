local balancer_resty = require("balancer.resty")
local resty_chash = require("resty.chash")
local util = require("util")
local split = require("util.split")

local _M = balancer_resty:new({ factory = resty_chash, name = "chash" })

function _M.new(self, backend)
  local nodes = util.get_nodes(backend.endpoints)
  local o = { instance = self.factory:new(nodes), hash_by = backend["upstream-hash-by"] }
  setmetatable(o, self)
  self.__index = self
  return o
end

function _M.balance(self)
  local key = util.lua_ngx_var(self.hash_by)
  local endpoint_string = self.instance:find(key)
  return split.split_pair(endpoint_string, ":")
end

return _M
