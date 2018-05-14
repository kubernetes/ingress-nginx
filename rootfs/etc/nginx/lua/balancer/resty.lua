local resty_roundrobin = require("resty.roundrobin")
local resty_chash = require("resty.chash")
local util = require("util")

local _M = {}
local instances = {}

local function get_resty_balancer_nodes(endpoints)
  local nodes = {}
  local weight = 1

  for _, endpoint in pairs(endpoints) do
    local endpoint_string = endpoint.address .. ":" .. endpoint.port
    nodes[endpoint_string] = weight
  end

  return nodes
end

local function init_resty_balancer(factory, instance, endpoints)
  local nodes = get_resty_balancer_nodes(endpoints)

  if instance then
    instance:reinit(nodes)
  else
    instance = factory:new(nodes)
  end

  return instance
end

function _M.balance(backend)
  local instance = instances[backend.name]
  if not instance then
    ngx.log(ngx.ERR, "no LB algorithm instance was found")
    return nil
  end

  local endpoint_string
  if backend["upstream-hash-by"] then
    local key = util.lua_ngx_var(backend["upstream-hash-by"])
    endpoint_string = instance:find(key)
  else
    endpoint_string = instance:find()
  end

  local address, port = util.split_pair(endpoint_string, ":")
  return { address = address, port = port }
end

function _M.reinit(backend)
  local instance = instances[backend.name]
  local factory = resty_roundrobin
  if backend["upstream-hash-by"] then
    factory = resty_chash
  end

  if instance then
    local mt = getmetatable(instance)
    if mt.__index ~= factory then
      ngx.log(ngx.INFO, "LB algorithm has been changed, resetting the instance")
      instance = nil
    end
  end

  instances[backend.name] = init_resty_balancer(factory, instance, backend.endpoints)
end

return _M
