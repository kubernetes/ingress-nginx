-- An affinity mode which makes sure connections are rebalanced when a deployment is scaled.
-- The advantage of this mode is that the load on the pods will be redistributed.
-- The drawback of this mode is that, when scaling up a deployment, roughly (n-c)/n users
-- will lose their session, where c is the current number of pods and n is the new number of
-- pods.
--
local balancer_sticky = require("balancer.sticky")
local math_random = require("math").random
local resty_chash = require("resty.chash")
local util_get_nodes = require("util").get_nodes

local ngx = ngx
local string = string
local setmetatable = setmetatable

local _M = balancer_sticky:new()

-- Consider the situation of N upstreams one of which is failing.
-- Then the probability to obtain failing upstream after M iterations would be close to (1/N)**M.
-- For the worst case (2 upstreams; 20 iterations) it would be ~10**(-6)
-- which is much better then ~10**(-3) for 10 iterations.
local MAX_UPSTREAM_CHECKS_COUNT = 20

function _M.new(self, backend)
  local nodes = util_get_nodes(backend.endpoints)

  local o = {
    name = "sticky_balanced",
    instance = resty_chash:new(nodes)
  }

  setmetatable(o, self)
  self.__index = self

  balancer_sticky.sync(o, backend)

  return o
end

function _M.pick_new_upstream(self, failed_upstreams)
  for i = 1, MAX_UPSTREAM_CHECKS_COUNT do
    local key = string.format("%s.%s.%s", ngx.now() + i, ngx.worker.pid(), math_random(999999))
    local new_upstream = self.instance:find(key)

    if not failed_upstreams[new_upstream] then
      return new_upstream, key
    end
  end

  return nil, nil
end

return _M
