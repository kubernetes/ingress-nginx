local ngx = ngx
local setmetatable = setmetatable

local _M = {}

function _M.rewrite()
  local ua = ngx.var.http_user_agent

  if ua == "hello" then
    ngx.req.set_header("x-hello-world", "1")
  end
end

function _M.balancer_implementation()
  -- An example balancer implementation that always returns the first endpoint.
  -- Used for demonstration and testing purposes only.
  return {
    name = "hello_world",
    new = function(self, backend)
      local o = {
        endpoints = backend.endpoints,
        traffic_shaping_policy = backend.trafficShapingPolicy,
        alternative_backends = backend.alternativeBackends,
      }
      setmetatable(o, self)
      self.__index = self
      return o
    end,
    is_affinitized = function(_) return false end,
    after_balance = function(_) end,
    sync = function(self, backend)
      self.endpoints = backend.endpoints
      self.traffic_shaping_policy = backend.trafficShapingPolicy
      self.alternative_backends = backend.alternativeBackends
    end,
    balance = function(self)
      local endpoint = self.endpoints[1]
      return endpoint.address .. ":" .. endpoint.port
    end,
  }
end

return _M
