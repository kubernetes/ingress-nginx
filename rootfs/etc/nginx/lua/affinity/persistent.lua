-- An affinity mode which makes sure a session is always routed to the same endpoint.
-- The advantage of this mode is that a user will never lose his session.
-- The drawback of this mode is that when scaling up a deployment, sessions will not
-- be rebalanced.
--
local util = require("util")
local util_nodemap = require("util.nodemap")

local _M = {}

local function get_routing_key(self)
  local cookie_value = self:get_cookie()

  if cookie_value then
    -- format <timestamp>.<workder-pid>.<routing-key>
    local routing_key = string.match(cookie_value, '[^\\.]+$')

    if routing_key == nil then
      local err = string.format("Failed to extract routing key from cookie '%s'!", cookie_value)
      return nil, err
    end

    return routing_key, nil
  end

  return nil, nil
end

local function set_routing_key(self, key)
  local value = string.format("%s.%s.%s", ngx.now(), ngx.worker.pid(), key)
  self:set_cookie(value);
end

local function pick_new_upstream(self, failed_upstreams)
  return self.instance:random_except(failed_upstreams)
end

function _M.new(self, sticky_balancer, backend)
  local o = sticky_balancer or {}
  
  local nodes = util.get_nodes(backend.endpoints)
  local hash_salt = backend["name"]

  -- override sticky.balancer methods
  o.instance = util_nodemap:new(nodes, hash_salt)
  o.get_routing_key = get_routing_key
  o.set_routing_key = set_routing_key
  o.pick_new_upstream = pick_new_upstream

  return sticky_balancer
end

return _M
