-- An affinity mode which makes sure a session is always routed to the same endpoint.
-- The advantage of this mode is that a user will never lose his session.
-- The drawback of this mode is that when scaling up a deployment, sessions will not
-- be rebalanced.
--
local balancer_sticky = require("balancer.sticky")
local util_get_nodes = require("util").get_nodes
local util_nodemap = require("util.nodemap")
local setmetatable = setmetatable

local _M = balancer_sticky:new()

function _M.new(self, backend)
  local nodes = util_get_nodes(backend.endpoints)
  local hash_salt = backend["name"]

  local o = {
    name = "sticky_persistent",
    instance = util_nodemap:new(nodes, hash_salt)
  }

  setmetatable(o, self)
  self.__index = self

  balancer_sticky.sync(o, backend)

  return o
end

function _M.pick_new_upstream(self, failed_upstreams)
  return self.instance:random_except(failed_upstreams)
end

return _M
