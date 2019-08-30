-- An affinity mode which makes sure connections are rebalanced when a deployment is scaled.
-- The advantage of this mode is that the load on the pods will be redistributed.
-- The drawback of this mode is that, when scaling up a deployment, roughly (n-c)/n users 
-- will lose their session, where c is the current number of pods and n is the new number of 
-- pods.
--
-- This class extends/implements the abstract class balancer.sticky.
--
local math = require("math")
local resty_chash = require("resty.chash")
local util = require("util")

local _M = {}

-- Consider the situation of N upstreams one of which is failing.
-- Then the probability to obtain failing upstream after M iterations would be close to (1/N)**M.
-- For the worst case (2 upstreams; 20 iterations) it would be ~10**(-6)
-- which is much better then ~10**(-3) for 10 iterations.
local MAX_UPSTREAM_CHECKS_COUNT = 20

local function get_routing_key(self)
  return self:get_cookie(), nil
end

local function set_routing_key(self, key)
	self:set_cookie(key)
end

local function pick_new_upstream(self, failed_upstreams)
  for i = 1, MAX_UPSTREAM_CHECKS_COUNT do
    local key = string.format("%s.%s.%s", ngx.now() + i, ngx.worker.pid(), math.random(999999))

    local new_upstream = self.instance:find(key)

    if not failed_upstreams[new_upstream] then
      return new_upstream, key
    end
  end

  return nil, nil
end

function _M.new(self, sticky_balancer, backend)
  local o = sticky_balancer or {}
  
  local nodes = util.get_nodes(backend.endpoints)

  -- override sticky.balancer methods
  o.instance = resty_chash:new(nodes)
  o.get_routing_key = get_routing_key
  o.set_routing_key = set_routing_key
  o.pick_new_upstream = pick_new_upstream

  return sticky_balancer
end

return _M
