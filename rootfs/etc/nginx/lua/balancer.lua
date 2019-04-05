local ngx_balancer = require("ngx.balancer")
local cjson = require("cjson.safe")
local util = require("util")
local dns_util = require("util.dns")
local configuration = require("configuration")
local round_robin = require("balancer.round_robin")
local chash = require("balancer.chash")
local chashsubset = require("balancer.chashsubset")
local sticky = require("balancer.sticky")
local ewma = require("balancer.ewma")

-- measured in seconds
-- for an Nginx worker to pick up the new list of upstream peers
-- it will take <the delay until controller POSTed the backend object to the Nginx endpoint> + BACKENDS_SYNC_INTERVAL
local BACKENDS_SYNC_INTERVAL = 1

local DEFAULT_LB_ALG = "round_robin"
local IMPLEMENTATIONS = {
  round_robin = round_robin,
  chash = chash,
  chashsubset = chashsubset,
  sticky = sticky,
  ewma = ewma,
}

local _M = {}
local balancers = {}

local function get_implementation(backend)
  local name = backend["load-balance"] or DEFAULT_LB_ALG

  if backend["sessionAffinityConfig"] and backend["sessionAffinityConfig"]["name"] == "cookie" then
    name = "sticky"
  elseif backend["upstreamHashByConfig"] and backend["upstreamHashByConfig"]["upstream-hash-by"] then
    if backend["upstreamHashByConfig"]["upstream-hash-by-subset"] then
      name = "chashsubset"
    else
      name = "chash"
    end
  end

  local implementation = IMPLEMENTATIONS[name]
  if not implementation then
    ngx.log(ngx.WARN, string.format("%s is not supported, falling back to %s", backend["load-balance"], DEFAULT_LB_ALG))
    implementation = IMPLEMENTATIONS[DEFAULT_LB_ALG]
  end

  return implementation
end

local function resolve_external_names(original_backend)
  local backend = util.deepcopy(original_backend)
  local endpoints = {}
  for _, endpoint in ipairs(backend.endpoints) do
    local ips = dns_util.resolve(endpoint.address)
    for _, ip in ipairs(ips) do
      table.insert(endpoints, { address = ip, port = endpoint.port })
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

local function sync_backend(backend)
  if not backend.endpoints or #backend.endpoints == 0 then
    ngx.log(ngx.INFO, string.format("there is no endpoint for backend %s. Removing...", backend.name))
    balancers[backend.name] = nil
    return
  end

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
    ngx.log(ngx.INFO,
      string.format("LB algorithm changed from %s to %s, resetting the instance", balancer.name, implementation.name))
    balancers[backend.name] = implementation:new(backend)
    return
  end

  local service_type = backend.service and backend.service.spec and backend.service.spec["type"]
  if service_type == "ExternalName" then
    backend = resolve_external_names(backend)
  end

  backend.endpoints = format_ipv6_endpoints(backend.endpoints)

  balancer:sync(backend)
end

local function sync_backends()
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
  end

  for backend_name, _ in pairs(balancers) do
    if not balancers_to_keep[backend_name] then
      balancers[backend_name] = nil
    end
  end
end

local function route_to_alternative_balancer(balancer)
  if not balancer.alternative_backends then
    return false
  end

  -- TODO: support traffic shaping for n > 1 alternative backends
  local backend_name = balancer.alternative_backends[1]
  if not backend_name then
    ngx.log(ngx.ERR, "empty alternative backend")
    return false
  end

  local alternative_balancer = balancers[backend_name]
  if not alternative_balancer then
    ngx.log(ngx.ERR, "no alternative balancer for backend: " .. tostring(backend_name))
    return false
  end

  local traffic_shaping_policy =  alternative_balancer.traffic_shaping_policy
  if not traffic_shaping_policy then
    ngx.log(ngx.ERR, "traffic shaping policy is not set for balanacer of backend: " .. tostring(backend_name))
    return false
  end

  local target_header = util.replace_special_char(traffic_shaping_policy.header, "-", "_")
  local header = ngx.var["http_" .. target_header]
  if header then
    if traffic_shaping_policy.headerValue and #traffic_shaping_policy.headerValue > 0 then
      if traffic_shaping_policy.headerValue == header then
        return true
      end
    elseif header == "always" then
      return true
    elseif header == "never" then
      return false
    end
  end

  local target_cookie = traffic_shaping_policy.cookie
  local cookie = ngx.var["cookie_" .. target_cookie]
  if cookie then
    if cookie == "always" then
      return true
    elseif cookie == "never" then
      return false
    end
  end

  if math.random(100) <= traffic_shaping_policy.weight then
    return true
  end

  return false
end

local function get_balancer()
  local backend_name = ngx.var.proxy_upstream_name

  local balancer = balancers[backend_name]
  if not balancer then
    return
  end

  if route_to_alternative_balancer(balancer) then
    local alternative_balancer = balancers[balancer.alternative_backends[1]]
    return alternative_balancer
  end

  return balancer
end

function _M.init_worker()
  sync_backends() -- when worker starts, sync backends without delay
  local _, err = ngx.timer.every(BACKENDS_SYNC_INTERVAL, sync_backends)
  if err then
    ngx.log(ngx.ERR, string.format("error when setting up timer.every for sync_backends: %s", tostring(err)))
  end
end

function _M.rewrite()
  local balancer = get_balancer()
  if not balancer then
    ngx.status = ngx.HTTP_SERVICE_UNAVAILABLE
    return ngx.exit(ngx.status)
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

if _TEST then
  _M.get_implementation = get_implementation
  _M.sync_backend = sync_backend
end

return _M
