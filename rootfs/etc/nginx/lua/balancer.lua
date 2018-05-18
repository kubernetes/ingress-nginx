local ngx_balancer = require("ngx.balancer")
local json = require("cjson")
local configuration = require("configuration")
local round_robin = require("balancer.round_robin")
local chash = require("balancer.chash")
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
  sticky = sticky,
  ewma = ewma,
}

local _M = {}
local balancers = {}

local function get_implementation(backend)
  local name = backend["load-balance"] or DEFAULT_LB_ALG

  if backend["sessionAffinityConfig"] and backend["sessionAffinityConfig"]["name"] == "cookie" then
    name = "sticky"
  elseif backend["upstream-hash-by"] then
    name = "chash"
  end

  local implementation = IMPLEMENTATIONS[name]
  if not implementation then
    ngx.log(ngx.WARN, string.format("%s is not supported, falling back to %s", backend["load-balance"], DEFAULT_LB_ALG))
    implementation = IMPLEMENTATIONS[DEFAULT_LB_ALG]
  end

  return implementation
end

local function sync_backend(backend)
  local implementation = get_implementation(backend)
  local balancer = balancers[backend.name]

  if not balancer then
    balancers[backend.name] = implementation:new(backend)
    return
  end

  if getmetatable(balancer) ~= implementation then
    ngx.log(ngx.INFO,
      string.format("LB algorithm changed from %s to %s, resetting the instance", balancer.name, implementation.name))
    balancers[backend.name] = implementation:new(backend)
    return
  end

  balancer:sync(backend)
end

local function sync_backends()
  local backends_data = configuration.get_backends_data()
  if not backends_data then
    return
  end

  local ok, new_backends = pcall(json.decode, backends_data)
  if not ok then
    ngx.log(ngx.ERR,  "could not parse backends data: " .. tostring(new_backends))
    return
  end

  for _, new_backend in pairs(new_backends) do
    sync_backend(new_backend)
  end
end

function _M.init_worker()
  sync_backends() -- when worker starts, sync backends without delay
  local _, err = ngx.timer.every(BACKENDS_SYNC_INTERVAL, sync_backends)
  if err then
    ngx.log(ngx.ERR, string.format("error when setting up timer.every for sync_backends: %s", tostring(err)))
  end
end

function _M.call()
  local phase = ngx.get_phase()
  if phase ~= "log" and phase ~= "balancer" then
    ngx.log(ngx.ERR, "must be called in balancer or log, but was called in: " .. phase)
    return
  end

  local backend_name = ngx.var.proxy_upstream_name
  local balancer = balancers[backend_name]
  if not balancer then
    ngx.status = ngx.HTTP_SERVICE_UNAVAILABLE
    return ngx.exit(ngx.status)
  end

  if phase == "log" then
    balancer:after_balance()
    return
  end

  local host, port = balancer:balance()
  if not host then
    ngx.status = ngx.HTTP_SERVICE_UNAVAILABLE
    return ngx.exit(ngx.status)
  end

  ngx_balancer.set_more_tries(1)

  local ok, err = ngx_balancer.set_current_peer(host, port)
  if not ok then
    ngx.log(ngx.ERR, "error while setting current upstream peer to " .. tostring(err))
  end
end

if _TEST then
  _M.get_implementation = get_implementation
  _M.sync_backend = sync_backend
end

return _M
