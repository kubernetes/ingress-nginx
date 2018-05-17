local ngx_balancer = require("ngx.balancer")
local json = require("cjson")
local configuration = require("configuration")
local util = require("util")
local lrucache = require("resty.lrucache")
local ewma = require("balancer.ewma")
local resty_balancer = require("balancer.resty")

-- measured in seconds
-- for an Nginx worker to pick up the new list of upstream peers
-- it will take <the delay until controller POSTed the backend object to the Nginx endpoint> + BACKENDS_SYNC_INTERVAL
local BACKENDS_SYNC_INTERVAL = 1

local DEFAULT_LB_ALG = "round_robin"

local _M = {}

-- TODO(elvinefendi) we can probably avoid storing all backends here. We already store them in their respective
-- load balancer implementations
local backends, backends_err = lrucache.new(1024)
if not backends then
  return error("failed to create the cache for backends: " .. (backends_err or "unknown"))
end

local function get_current_backend()
  local backend_name = ngx.var.proxy_upstream_name
  local backend = backends:get(backend_name)

  if not backend then
    -- TODO(elvinefendi) maybe force backend sync here?
    ngx.log(ngx.WARN, "no backend configuration found for " .. tostring(backend_name))
  end

  return backend
end

local function get_balancer(backend)
  if not backend then
    return nil
  end

  local lb_alg = backend["load-balance"] or DEFAULT_LB_ALG
  if resty_balancer.is_applicable(backend) then
    return resty_balancer
  elseif lb_alg ~= "ewma" then
    if lb_alg ~= DEFAULT_LB_ALG then
      ngx.log(ngx.WARN,
        string.format("%s is not supported, falling back to %s", backend["load-balance"], DEFAULT_LB_ALG))
    end
    return resty_balancer
  end

  return ewma
end

local function balance()
  local backend = get_current_backend()
  local balancer = get_balancer(backend)
  if not balancer then
    return nil, nil
  end

  local endpoint = balancer.balance(backend)
  if not endpoint then
    return nil, nil
  end

  return endpoint.address, endpoint.port
end

local function sync_backend(backend)
  backends:set(backend.name, backend)

  local balancer = get_balancer(backend)
  if not balancer then
    return
  end
  balancer.sync(backend)
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
    local backend = backends:get(new_backend.name)
    local backend_changed = true

    if backend then
      backend_changed = not util.deep_compare(backend, new_backend)
    end

    if backend_changed then
      sync_backend(new_backend)
    end
  end
end

local function after_balance()
  local backend = get_current_backend()
  local balancer = get_balancer(backend)
  if not balancer then
    return
  end

  balancer.after_balance()
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
  if phase == "log" then
    after_balance()
    return
  end
  if phase ~= "balancer" then
    return error("must be called in balancer or log, but was called in: " .. phase)
  end

  local host, port = balance()
  if not host then
    ngx.status = ngx.HTTP_SERVICE_UNAVAILABLE
    return ngx.exit(ngx.status)
  end

  ngx_balancer.set_more_tries(1)

  local ok, err = ngx_balancer.set_current_peer(host, port)
  if not ok then
    ngx.log(ngx.ERR, string.format("error while setting current upstream peer to %s", tostring(err)))
  end
end

return _M
