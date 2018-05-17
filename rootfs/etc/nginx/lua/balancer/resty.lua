local resty_roundrobin = require("resty.roundrobin")
local resty_chash = require("resty.chash")
local util = require("util")
local ck = require("resty.cookie")

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

local function is_sticky(backend)
  return backend["sessionAffinityConfig"] and backend["sessionAffinityConfig"]["name"] == "cookie"
end

local function cookie_name(backend)
  return backend["sessionAffinityConfig"]["cookieSessionAffinity"]["name"] or "route"
end

local function encrypted_endpoint_string(backend, endpoint_string)
  local encrypted, err
  if backend["sessionAffinityConfig"]["cookieSessionAffinity"]["hash"] == "sha1" then
    encrypted, err = util.sha1_digest(endpoint_string)
  else
    encrypted, err = util.md5_digest(endpoint_string)
  end
  if err ~= nil then
    ngx.log(ngx.ERR, err)
  end

  return encrypted
end

local function set_cookie(backend, value)
  local cookie, err = ck:new()
  if not cookie then
    ngx.log(ngx.ERR, err)
  end

  local ok
  ok, err = cookie:set({
    key = cookie_name(backend),
    value = value,
    path = "/",
    domain = ngx.var.host,
    httponly = true,
  })
  if not ok then
    ngx.log(ngx.ERR, err)
  end
end

local function pick_random(instance)
  local index = math.random(instance.npoints)
  return instance:next(index)
end

local function sticky_endpoint_string(instance, backend)
  local cookie, err = ck:new()
  if not cookie then
    ngx.log(ngx.ERR, err)
    return pick_random(instance)
  end

  local key = cookie:get(cookie_name(backend))
  if not key then
    local tmp_endpoint_string = pick_random(instance)
    key = encrypted_endpoint_string(backend, tmp_endpoint_string)
    set_cookie(backend, key)
  end

  return instance:find(key)
end

function _M.is_applicable(backend)
  return is_sticky(backend) or backend["upstream-hash-by"] or backend["load-balance"] == "round_robin"
end

function _M.balance(backend)
  local instance = instances[backend.name]
  if not instance then
    ngx.log(ngx.ERR, "no LB algorithm instance was found")
    return nil
  end

  local endpoint_string
  if is_sticky(backend) then
    endpoint_string = sticky_endpoint_string(instance, backend)
  elseif backend["upstream-hash-by"] then
    local key = util.lua_ngx_var(backend["upstream-hash-by"])
    endpoint_string = instance:find(key)
  else
    endpoint_string = instance:find()
  end

  local address, port = util.split_pair(endpoint_string, ":")
  return { address = address, port = port }
end

function _M.sync(backend)
  local instance = instances[backend.name]
  local factory = resty_roundrobin
  if is_sticky(backend) or backend["upstream-hash-by"] then
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

function _M.after_balance()
end

return _M
