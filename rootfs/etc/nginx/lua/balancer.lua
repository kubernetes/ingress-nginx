local ngx_balancer = require("ngx.balancer")
local cjson = require("cjson.safe")
local util = require("util")
local dns_lookup = require("util.dns").lookup
local configuration = require("configuration")
local round_robin = require("balancer.round_robin")
local chash = require("balancer.chash")
local chashsubset = require("balancer.chashsubset")
local sticky_balanced = require("balancer.sticky_balanced")
local sticky_persistent = require("balancer.sticky_persistent")
local ewma = require("balancer.ewma")
local string = string
local ipairs = ipairs
local table = table
local getmetatable = getmetatable
local tostring = tostring
local pairs = pairs
local math = math
local ngx = ngx
local unpack = unpack

-- measured in seconds
-- for an Nginx worker to pick up the new list of upstream peers
-- it will take <the delay until controller POSTed the backend object to the
-- Nginx endpoint> + BACKENDS_SYNC_INTERVAL
local BACKENDS_SYNC_INTERVAL = 1

local DEFAULT_LB_ALG = "round_robin"
local IMPLEMENTATIONS = {
  round_robin = round_robin,
  chash = chash,
  chashsubset = chashsubset,
  sticky_balanced = sticky_balanced,
  sticky_persistent = sticky_persistent,
  ewma = ewma,
}

local PROHIBITED_LOCALHOST_PORT = configuration.prohibited_localhost_port or '10246'
local PROHIBITED_PEER_PATTERN = "^127.*:" .. PROHIBITED_LOCALHOST_PORT .. "$"

local _M = {}
local balancers = {}
local backends_with_external_name = {}
local backends_last_synced_at = 0

local function get_implementation(backend)
  local name = backend["load-balance"] or DEFAULT_LB_ALG

  if backend["sessionAffinityConfig"] and
     backend["sessionAffinityConfig"]["name"] == "cookie" then
    if backend["sessionAffinityConfig"]["mode"] == "persistent" then
      name = "sticky_persistent"
    else
      name = "sticky_balanced"
    end

  elseif backend["upstreamHashByConfig"] and
         backend["upstreamHashByConfig"]["upstream-hash-by"] then
    if backend["upstreamHashByConfig"]["upstream-hash-by-subset"] then
      name = "chashsubset"
    else
      name = "chash"
    end
  end

  local implementation = IMPLEMENTATIONS[name]
  if not implementation then
    ngx.log(ngx.WARN, backend["load-balance"], " is not supported, ",
            "falling back to ", DEFAULT_LB_ALG)
    implementation = IMPLEMENTATIONS[DEFAULT_LB_ALG]
  end

  return implementation
end

local function resolve_external_names(original_backend)
  if not original_backend.endpoints or
     #original_backend.endpoints == 0 then
    return original_backend
  end
  local backend = util.deepcopy(original_backend)
  local endpoints = {}
  for _, endpoint in ipairs(backend.endpoints) do
    local ips = dns_lookup(endpoint.address)
    if #ips ~= 1 or ips[1] ~= endpoint.address then
      for _, ip in ipairs(ips) do
        table.insert(endpoints, { address = ip, port = endpoint.port })
      end
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
  -- We resolve external names before checking if the endpoints are empty
  -- because the behavior for resolve_external_names when the name was not
  -- resolved is to return an empty table so we set the balancer to nil below
  -- see https://github.com/kubernetes/ingress-nginx/pull/10989
  if is_backend_with_external_name(backend) then
    backend = resolve_external_names(backend)
  end

  if not backend.endpoints or #backend.endpoints == 0 then
    balancers[backend.name] = nil
    return
  end

  backend.endpoints = format_ipv6_endpoints(backend.endpoints)

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
        string.format("LB algorithm changed from %s to %s, resetting the instance",
                      balancer.name, implementation.name))
    balancers[backend.name] = implementation:new(backend)
    return
  end

  balancer:sync(backend)
end

local function sync_backends_with_external_name()
  for _, backend_with_external_name in pairs(backends_with_external_name) do
    sync_backend(backend_with_external_name)
  end
end

local function sync_backends()
  local raw_backends_last_synced_at = configuration.get_raw_backends_last_synced_at()
  if raw_backends_last_synced_at <= backends_last_synced_at then
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
    if is_backend_with_external_name(new_backend) then
      local backend_with_external_name = util.deepcopy(new_backend)
      backends_with_external_name[backend_with_external_name.name] = backend_with_external_name
    else
      sync_backend(new_backend)
    end
    balancers_to_keep[new_backend.name] = true
  end

  for backend_name, _ in pairs(balancers) do
    if not balancers_to_keep[backend_name] then
      balancers[backend_name] = nil
      backends_with_external_name[backend_name] = nil
    end
  end
  backends_last_synced_at = raw_backends_last_synced_at
end

local function get_alternative_or_original_balancer(balancer)
  -- checks whether an alternative backend should be used
  -- the first backend matching the request context is returned
  -- if no suitable alterntative backends are found, the original is returned
  if balancer.is_affinitized(balancer) then
    -- If request is already affinitized to a primary balancer, keep the primary balancer.
    return balancer
  end

  if not balancer.alternative_backends then
    return balancer
  end

  local available_now_balancers = {}
  for _, backend_name in ipairs(balancer.alternative_backends) do
    if not backend_name then
      ngx.log(ngx.ERR, "empty alternative backend")
    else
      local alternative_balancer = balancers[backend_name]
      if not alternative_balancer then
        ngx.log(ngx.ERR, "no alternative balancer for backend: ",
                tostring(backend_name))
      else
        if alternative_balancer.is_affinitized(alternative_balancer) then
          -- If request is affinitized to an alternative balancer, instruct caller to
          -- switch to alternative.
          return alternative_balancer
        elseif not alternative_balancer.traffic_shaping_policy then
          -- If alternative_balancer has no traffic_shaping_policy, do not save it and log error
          ngx.log(ngx.ERR, "traffic shaping policy is not set for balancer ",
                  "of backend: ", tostring(backend_name))
        else
          -- Save alternative_balancers with traffic shaping policy, if request didn't have affinity set.
          table.insert(available_now_balancers, alternative_balancer)
        end
      end
    end
  end

  local never_id_list = {}
  -- Using by-header traffic_shaping_policy to find a suitable alternative_balancer
  for i, alternative_balancer in ipairs(available_now_balancers) do
    local traffic_shaping_policy =  alternative_balancer.traffic_shaping_policy
    local target_header = util.replace_special_char(traffic_shaping_policy.header,
                                                    "-", "_")
    local header = ngx.var["http_" .. target_header]
    if header then
      if traffic_shaping_policy.headerValue
             and #traffic_shaping_policy.headerValue > 0 then
        if traffic_shaping_policy.headerValue == header then
          return alternative_balancer
        end
      elseif traffic_shaping_policy.headerPattern
         and #traffic_shaping_policy.headerPattern > 0 then
        -- Check headerPattern if it was specified
        local m, err = ngx.re.match(header, traffic_shaping_policy.headerPattern)
        if m then
          return alternative_balancer
        elseif err then
            ngx.log(ngx.ERR, "error when matching canary-by-header-pattern: '",
                    traffic_shaping_policy.headerPattern, "', error: ", err)
        -- if headerPattern is broken, we remove this alternative_backend from further processing
        table.insert(never_id_list, i)
        end
      -- If header was specified, but headerValue or headerPattern was not, check the value of header by "always"/"never"
      elseif header == "always" then
        return alternative_balancer
      elseif header == "never" then
        -- This alternative_balancer will not be used for further checks
        -- Saving the alternative_balancer's ID wich we will not use for further checks
        table.insert(never_id_list, i)
      end
    end
  end

  if #never_id_list > 0 then
    table.remove(available_now_balancers, unpack(never_id_list))
  end
  never_id_list = {}

  -- Using by-header traffic_shaping_policy to find a suitable alternative_balancer
  for i, alternative_balancer in ipairs(available_now_balancers) do
    local traffic_shaping_policy =  alternative_balancer.traffic_shaping_policy
    local target_cookie = traffic_shaping_policy.cookie
    local cookie = ngx.var["cookie_" .. target_cookie]

    if cookie then
      if cookie == "always" then
        return alternative_balancer
      elseif cookie == "never" then
        table.insert(never_id_list, i)
      end
    end
  end

  if #never_id_list > 0 then
    table.remove(available_now_balancers, unpack(never_id_list))
  end

  for _, alternative_balancer in ipairs(available_now_balancers) do
    local traffic_shaping_policy =  alternative_balancer.traffic_shaping_policy

    if traffic_shaping_policy.weight ~= nil and traffic_shaping_policy.weight > 0 then
      local weightTotal = 100
      if traffic_shaping_policy.weightTotal ~= nil and traffic_shaping_policy.weightTotal > 100 then
        weightTotal = traffic_shaping_policy.weightTotal
      end
      if math.random(weightTotal) <= traffic_shaping_policy.weight then
        return alternative_balancer
      end
    end
  end

  return balancer
end

local function get_balancer_by_upstream_name(upstream_name)
  return balancers[upstream_name]
end

local function get_balancer()
  if ngx.ctx.balancer then
    return ngx.ctx.balancer
  end

  local backend_name = ngx.var.proxy_upstream_name

  local balancer = balancers[backend_name]
  if not balancer then
    return nil
  end

  ngx.ctx.balancer = get_alternative_or_original_balancer(balancer)

  return ngx.ctx.balancer
end

function _M.init_worker()
  -- when worker starts, sync non ExternalName backends without delay
  sync_backends()
  -- we call sync_backends_with_external_name in timer because for endpoints that require
  -- DNS resolution it needs to use socket which is not available in
  -- init_worker phase
  local ok, err = ngx.timer.at(0, sync_backends_with_external_name)
  if not ok then
    ngx.log(ngx.ERR, "failed to create timer: ", err)
  end

  ok, err = ngx.timer.every(BACKENDS_SYNC_INTERVAL, sync_backends)
  if not ok then
    ngx.log(ngx.ERR, "error when setting up timer.every for sync_backends: ", err)
  end
  ok, err = ngx.timer.every(BACKENDS_SYNC_INTERVAL, sync_backends_with_external_name)
  if not ok then
    ngx.log(ngx.ERR, "error when setting up timer.every for sync_backends_with_external_name: ",
            err)
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

  if peer:match(PROHIBITED_PEER_PATTERN) then
    ngx.log(ngx.ERR, "attempted to proxy to self, balancer: ", balancer.name, ", peer: ", peer)
    return
  end

  ngx_balancer.set_more_tries(1)

  local ok, err = ngx_balancer.set_current_peer(peer)
  if not ok then
    ngx.log(ngx.ERR, "error while setting current upstream peer ", peer,
            ": ", err)
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
  get_alternative_or_original_balancer = get_alternative_or_original_balancer,
  get_balancer = get_balancer,
  get_balancer_by_upstream_name = get_balancer_by_upstream_name,
}})

return _M
