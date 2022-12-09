local balancer_resty = require("balancer.resty")
local resty_chash = require("resty.chash")
local util = require("util")
local ngx_log = ngx.log
local ngx_ERR = ngx.ERR
local setmetatable = setmetatable

local _M = balancer_resty:new({ factory = resty_chash, name = "chash" })

local function build_node_map(backend)
  local node_map = {}
  local nodes = {}
  local use_hostname = backend["upstreamHashByConfig"] and backend["upstreamHashByConfig"]["upstream-hash-by-use-hostname"]

  for _, endpoint in pairs(backend.endpoints) do
    local id = endpoint.address .. ":" .. endpoint.port
    if use_hostname then
      id = endpoint.hostname
    end
    nodes[id] = endpoint
    node_map[id] = 1
  end

  return node_map, nodes
end

function _M.new(self, backend)
  local node_map, nodes = build_node_map(backend)
  local complex_val, err =
    util.parse_complex_value(backend["upstreamHashByConfig"]["upstream-hash-by"])
  if err ~= nil then
    ngx_log(ngx_ERR, "could not parse the value of the upstream-hash-by: ", err)
  end

  local o = {
    instance = self.factory:new(node_map),
    hash_by = complex_val,
    nodes = nodes,
    current_endpoints = backend.endpoints,
    traffic_shaping_policy = backend.trafficShapingPolicy,
    alternative_backends = backend.alternativeBackends,
  }
  setmetatable(o, self)
  self.__index = self
  return o
end

function _M.balance(self)
  local key = util.generate_var_value(self.hash_by)
  local id = self.instance:find(key)
  local endpoint = self.nodes[id]
  return endpoint.address .. ":" .. endpoint.port
end

function _M.sync(self, backend)
  local node_map

  local changed = not util.deep_compare(self.current_endpoints, backend.endpoints)
  if not changed then
    return
  end

  self.current_endpoints = backend.endpoints

  node_map, self.nodes = build_node_map(backend)

  self.instance:reinit(node_map)
end

return _M
