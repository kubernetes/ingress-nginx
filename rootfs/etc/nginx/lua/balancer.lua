local ngx_balancer = require("ngx.balancer")
local json = require("cjson")
local configuration = require("configuration")
local util = require("util")
local lrucache = require("resty.lrucache")
local resty_lock = require("resty.lock")

-- measured in seconds
-- for an Nginx worker to pick up the new list of upstream peers 
-- it will take <the delay until controller POSTed the backend object to the Nginx endpoint> + BACKENDS_SYNC_INTERVAL
local BACKENDS_SYNC_INTERVAL = 1

ROUND_ROBIN_LOCK_KEY = "round_robin"

local round_robin_state = ngx.shared.round_robin_state

local _M = {}

local round_robin_lock = resty_lock:new("locks", {timeout = 0, exptime = 0.1})

local backends, err = lrucache.new(1024)
if not backends then
  return error("failed to create the cache for backends: " .. (err or "unknown"))
end

local function balance()
  local backend_name = ngx.var.proxy_upstream_name
  local backend = backends:get(backend_name)
  -- lb_alg field does not exist for ingress.Backend struct for now, so lb_alg
  -- will always be round_robin
  local lb_alg = backend.lb_alg or "round_robin"

  if lb_alg == "ip_hash" then
    -- TODO(elvinefendi) implement me
    return backend.endpoints[0].address, backend.endpoints[0].port
  end

  -- Round-Robin
  round_robin_lock:lock(backend_name .. ROUND_ROBIN_LOCK_KEY)
  local index = round_robin_state:get(backend_name)
  local index, endpoint = next(backend.endpoints, index)
  if not index then
    index = 1
    endpoint = backend.endpoints[index]
  end
  round_robin_state:set(backend_name, index)
  round_robin_lock:unlock(backend_name .. ROUND_ROBIN_LOCK_KEY)

  return endpoint.address, endpoint.port
end

local function sync_backend(backend)
  backends:set(backend.name, backend)

  -- also reset the respective balancer state since backend has changed
  round_robin_state:delete(backend.name)

  ngx.log(ngx.INFO, "syncronization completed for: " .. backend.name)
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

function _M.init_worker()
  _, err = ngx.timer.every(BACKENDS_SYNC_INTERVAL, sync_backends)
  if err then
    ngx.log(ngx.ERR, "error when setting up timer.every for sync_backends: " .. tostring(err))
  end
end

function _M.call()
  ngx_balancer.set_more_tries(1)

  local host, port = balance()

  local ok, err = ngx_balancer.set_current_peer(host, port)
  if ok then
    ngx.log(ngx.INFO, "current peer is set to " .. host .. ":" .. port)
  else
    ngx.log(ngx.ERR, "error while setting current upstream peer to: " .. tostring(err))
  end
end

return _M
