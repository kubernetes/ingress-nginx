local ngx_balancer = require("ngx.balancer")
local cjson = require("cjson.safe")
local util = require("util")
local dns_lookup = require("util.dns").lookup
local configuration = require("tcp_udp_configuration")
local round_robin = require("balancer.round_robin")

local ngx = ngx
local table = table
local ipairs = ipairs
local pairs = pairs
local tostring = tostring
local string = string
local getmetatable = getmetatable

-- measured in seconds
-- for an Nginx worker to pick up the new list of upstream peers
-- it will take <the delay until controller POSTed the backend object
-- to the Nginx endpoint> + BACKENDS_SYNC_INTERVAL
local BACKENDS_SYNC_INTERVAL = 1
local BACKENDS_FORCE_SYNC_INTERVAL = 30

local DEFAULT_LB_ALG = "round_robin"
local IMPLEMENTATIONS = {
  round_robin = round_robin
}

local PROHIBITED_LOCALHOST_PORT = configuration.prohibited_localhost_port or '10246'
local PROHIBITED_PEER_PATTERN = "^127.*:" .. PROHIBITED_LOCALHOST_PORT .. "$"

local _M = {}
local balancers = {}
local backends_with_external_name = {}
local backends_last_synced_at = 0

local function get_implementation(backend)
  local name = backend["load-balance"] or DEFAULT_LB_ALG

  local implementation = IMPLEMENTATIONS[name]
  if not implementation then
    ngx.log(ngx.WARN, string.format("%s is not supported, falling back to %s",
      backend["load-balance"], DEFAULT_LB_ALG))
    implementation = IMPLEMENTATIONS[DEFAULT_LB_ALG]
  end

  return implementation
end

local function resolve_external_names(original_backend)
  local backend = util.deepcopy(original_backend)
  local endpoints = {}
  for _, endpoint in ipairs(backend.endpoints) do
    local ips = dns_lookup(endpoint.address)
    for _, ip in ipairs(ips) do
      table.insert(endpoints, {address = ip, port = endpoint.port})
    end
  end
  backend.endpoints = endpoints
  return backend
end

local function format_ipv6_endpoints(endpoints)
  local formatted_endpoints = {}
  for _, endpoint in ipairs(endpoints) do
    local formatted_endpoint = endpoint
    if not endpoint.address:match("^%d+.%d+.%d+.%d+$") then
      formatted_endpoint.address = string.format("[%s]", endpoint.address)
    end
    table.insert(formatted_endpoints, formatted_endpoint)
  end
  return formatted_endpoints
end

local function is_backend_with_external_name(backend)
  local serv_type = backend.service and backend.service.spec
                      and backend.service.spec["type"]
  return serv_type == "ExternalName"
end

local function sync_backend(backend)
  if not backend.endpoints or #backend.endpoints == 0 then
    return
  end

  ngx.log(ngx.INFO, "sync tcp/udp backend: ", backend.name)
  local implementation = get_implementation(backend)
  local balancer = balancers[backend.name]

  if not balancer then
    balancers[backend.name] = implementation:new(backend)
    return
  end

  -- every implementation is the metatable of its instances (see .new(...) functions)
  -- here we check if `balancer` is the instance of `implementation`
  -- if it is not then we deduce LB algorithm has changed for the backend
  if getmetatable(balancer) ~= implementation then
    ngx.log(ngx.INFO, string.format("LB algorithm changed from %s to %s, "
      .. "resetting the instance", balancer.name, implementation.name))
    balancers[backend.name] = implementation:new(backend)
    return
  end

  if is_backend_with_external_name(backend) then
    backend = resolve_external_names(backend)
  end

  backend.endpoints = format_ipv6_endpoints(backend.endpoints)

  balancer:sync(backend)
end

local function sync_backends()
  local raw_backends_last_synced_at = configuration.get_raw_backends_last_synced_at()
  ngx.update_time()
  local current_timestamp = ngx.time()
  if current_timestamp - backends_last_synced_at < BACKENDS_FORCE_SYNC_INTERVAL
      and raw_backends_last_synced_at <= backends_last_synced_at then
    for _, backend_with_external_name in pairs(backends_with_external_name) do
      sync_backend(backend_with_external_name)
    end
    return
  end

  local backends_data = configuration.get_backends_data()
  if not backends_data then
    balancers = {}
    return
  end

  local new_backends, err = cjson.decode(backends_data)
  if not new_backends then
    ngx.log(ngx.ERR, "could not parse backends data: ", err)
    return
  end

  local balancers_to_keep = {}
  for _, new_backend in ipairs(new_backends) do
    sync_backend(new_backend)
    balancers_to_keep[new_backend.name] = balancers[new_backend.name]
    if is_backend_with_external_name(new_backend) then
      local backend_with_external_name = util.deepcopy(new_backend)
      backends_with_external_name[backend_with_external_name.name] = backend_with_external_name
    end
  end

  for backend_name, _ in pairs(balancers) do
    if not balancers_to_keep[backend_name] then
      balancers[backend_name] = nil
      backends_with_external_name[backend_name] = nil
    end
  end
  backends_last_synced_at = raw_backends_last_synced_at
end

local function get_balancer()
  local backend_name = ngx.var.proxy_upstream_name
  local balancer = balancers[backend_name]
  if not balancer then
    return
  end

  return balancer
end

function _M.init_worker()
  sync_backends() -- when worker starts, sync backends without delay
  local _, err = ngx.timer.every(BACKENDS_SYNC_INTERVAL, sync_backends)
  if err then
    ngx.log(ngx.ERR, string.format("error when setting up timer.every "
      .. "for sync_backends: %s", tostring(err)))
  end
end

function _M.balance()
  local balancer = get_balancer()
  if not balancer then
    return
  end

  local peer = balancer:balance()
  if not peer then
    ngx.log(ngx.WARN, "no peer was returned, balancer: " .. balancer.name)
    return
  end

  if peer:match(PROHIBITED_PEER_PATTERN) then
    ngx.log(ngx.ERR, "attempted to proxy to self, balancer: ", balancer.name, ", peer: ", peer)
    return
  end

  ngx_balancer.set_more_tries(1)

  local ok, err = ngx_balancer.set_current_peer(peer)
  if not ok then
    ngx.log(ngx.ERR, string.format("error while setting current upstream peer %s: %s", peer, err))
  end
end

function _M.log()
  local balancer = get_balancer()
  if not balancer then
    return
  end

  if not balancer.after_balance then
    return
  end

  balancer:after_balance()
end

setmetatable(_M, {__index = {
  get_implementation = get_implementation,
  sync_backend = sync_backend,
}})

return _M
