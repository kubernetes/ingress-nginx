local balancer_resty = require("balancer.resty")
local resty_chash = require("resty.chash")
local util = require("util")

local _M = balancer_resty:new({ factory = resty_chash, name = "chash" })

local function get_key(hash_by)
  local it, err = ngx.re.gmatch(hash_by, "([a-z0-9_]+)", "ij")
  if not it then
    ngx.log(ngx.ERR, "error: ", err)
    return
  end

  local key = ""

  while true do
    local m, err = it()
    if err then
      ngx.log(ngx.ERR, "error: ", err)
      return
    end

    if not m then
      -- no match found (any more)
      break
    end

    local value = util.lua_ngx_var(m[0])
    key = key .. tostring(value)
  end

  return key
end

function _M.new(self, backend)
  local nodes = util.get_nodes(backend.endpoints)
  local o = {
    instance = self.factory:new(nodes),
    hash_by = backend["upstream-hash-by"],
    traffic_shaping_policy = backend.trafficShapingPolicy,
    alternative_backends = backend.alternativeBackends,
  }
  setmetatable(o, self)
  self.__index = self
  return o
end

function _M.balance(self)
  local key = get_key(self.hash_by)
  ngx.log(ngx.INFO, "chash key to pick upstream peer: " .. tostring(key))

  return self.instance:find(key)
end

return _M
