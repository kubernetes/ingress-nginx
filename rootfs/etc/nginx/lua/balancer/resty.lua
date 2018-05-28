local util = require("util")

local _M = {}

function _M.new(self, o)
  o = o or {}
  setmetatable(o, self)
  self.__index = self
  return o
end

function _M.sync(self, backend)
  local nodes = util.get_nodes(backend.endpoints)
  local changed = not util.deep_compare(self.instance.nodes, nodes)
  if not changed then
    return
  end

  self.instance:reinit(nodes)
end

return _M
