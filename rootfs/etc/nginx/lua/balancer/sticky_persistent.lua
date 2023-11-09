-- An affinity mode which makes sure a session is always routed to the same endpoint.
-- The advantage of this mode is that a user will never lose his session.
-- The drawback of this mode is that when scaling up a deployment, sessions will not
-- be rebalanced.
--
local balancer_sticky = require("balancer.sticky")
local balancer_round_robin = require("balancer.round_robin")
local balancer_chash = require("balancer.chash")
local balancer_chashsubset = require("balancer.chashsubset")
local balancer_ewma = require("balancer.ewma")
local util_get_nodes = require("util").get_nodes
local util_nodemap = require("util.nodemap")
local setmetatable = setmetatable

local function get_secondary_balancer(backend)
  local name = backend["load-balance"]

  if not name then return nil
  elseif name == "chash" then return balancer_chash:new(backend)
  elseif name == "chashsubset" then return balancer_chashsubset:new(backend)
  elseif name == "round_robin" then return balancer_round_robin:new(backend)
  elseif name == "ewma" then return balancer_ewma:new(backend)
  end
end

local _M = balancer_sticky:new()

function _M.new(self, backend)
  local nodes = util_get_nodes(backend.endpoints)
  local hash_salt = backend["name"]
  local secondary_balancer = get_secondary_balancer(backend)

  local o = {
    name = "sticky_persistent",
    secondary_balancer = secondary_balancer,
    instance = util_nodemap:new(nodes, hash_salt)
  }

  setmetatable(o, self)
  self.__index = self

  balancer_sticky.sync(o, backend)

  return o
end

function _M.pick_new_upstream(self, failed_upstreams)
  if self.secondary_balancer then
    local endpoint = self.secondary_balancer:balance()
    local key = self.instance:key_from_endpoint(endpoint)
    if endpoint and key then
      return endpoint, key
    end
  end

  return self.instance:random_except(failed_upstreams)
end

function _M.sync(self, backend)
  -- sync inherited balancer
  balancer_sticky.sync(self, backend)
  
  -- note this may be inefficient
  -- perhaps better to only update if name changes?
  self.secondary_balancer = get_secondary_balancer(backend)

  -- sync secondary_balancer as well
  if self.secondary_balancer then
    self.secondary_balancer:sync(backend)
  end
end

return _M
