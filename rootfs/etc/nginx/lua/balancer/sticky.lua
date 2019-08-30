local affinity_balanced = require("affinity.balanced")
local affinity_persistent = require("affinity.persistent")
local balancer_resty = require("balancer.resty")
local util = require("util")
local ck = require("resty.cookie")
local ngx_balancer = require("ngx.balancer")
local split = require("util.split")

local string_format = string.format
local ngx_log = ngx.log
local INFO = ngx.INFO

local _M = balancer_resty:new({ name = "sticky" })
local DEFAULT_COOKIE_NAME = "route"


function _M.cookie_name(self)
  return self.cookie_session_affinity.name or DEFAULT_COOKIE_NAME
end

local function init_affinity_mode(self, backend)
  local mode = backend["sessionAffinityConfig"]["mode"] or 'balanced'

  -- set default mode to 'balanced' for backwards compatibility
  if mode == nil or mode == '' then
    mode = 'balanced'
  end

  self.affinity_mode = mode

  if mode == 'persistent' then
    return affinity_persistent:new(self, backend)
  end

  -- default is 'balanced' for backwards compatibility
  if mode ~= 'balanced' then
    ngx.log(ngx.WARN, string.format("Invalid affinity mode '%s'! Using 'balanced' as a default.", mode))
  end

  return affinity_balanced:new(self, backend)
end

function _M.new(self, backend)
  local o = {
    instance = nil,
    affinity_mode = nil,
    traffic_shaping_policy = backend.trafficShapingPolicy,
    alternative_backends = backend.alternativeBackends,
    cookie_session_affinity = backend["sessionAffinityConfig"]["cookieSessionAffinity"]
  }
  setmetatable(o, self)
  self.__index = self
  
  return init_affinity_mode(o, backend)
end

function _M.get_cookie(self)
  local cookie, err = ck:new()
  if not cookie then
    ngx.log(ngx.ERR, err)
  end

  return cookie:get(self:cookie_name())
end

function _M.set_cookie(self, value)
  local cookie, err = ck:new()
  if not cookie then
    ngx.log(ngx.ERR, err)
  end

  local cookie_path = self.cookie_session_affinity.path
  if not cookie_path then
    cookie_path = ngx.var.location_path
  end

  local cookie_data = {
    key = self:cookie_name(),
    value = value,
    path = cookie_path,
    httponly = true,
    secure = ngx.var.https == "on",
  }

  if self.cookie_session_affinity.expires and self.cookie_session_affinity.expires ~= "" then
      cookie_data.expires = ngx.cookie_time(ngx.time() + tonumber(self.cookie_session_affinity.expires))
  end

  if self.cookie_session_affinity.maxage and self.cookie_session_affinity.maxage ~= "" then
    cookie_data.max_age = tonumber(self.cookie_session_affinity.maxage)
  end

  local ok
  ok, err = cookie:set(cookie_data)
  if not ok then
    ngx.log(ngx.ERR, err)
  end
end

function _M.get_last_failure()
  return ngx_balancer.get_last_failure()
end

local function get_failed_upstreams()
  local indexed_upstream_addrs = {}
  local upstream_addrs = split.split_upstream_var(ngx.var.upstream_addr) or {}

  for _, addr in ipairs(upstream_addrs) do
    indexed_upstream_addrs[addr] = true
  end

  return indexed_upstream_addrs
end

--- get_routing_key gets the current routing key from the cookie
-- @treturn string, string The routing key and an error message if an error occured.
function _M.get_routing_key(self)
  -- interface method to get the routing key from the cookie
  -- has to be overridden by an affinity mode
  ngx.log(ngx.ERR, "[BUG] Failed to get routing key as no implementation has been provided!")
  return nil, nil
end

--- set_routing_key sets the current routing key on the cookie
-- @tparam string key The routing key to set on the cookie.
function _M.set_routing_key(self, key)
  -- interface method to set the routing key on the cookie
  -- has to be overridden by an affinity mode
  ngx.log(ngx.ERR, "[BUG] Failed to set routing key as no implementation has been provided!")
end

--- pick_new_upstream picks a new upstream while ignoring the given failed upstreams.
-- @tparam {[string]=boolean} A table of upstreams to ignore where the key is the endpoint and the value a boolean.
-- @treturn string, string The endpoint and its key.
function _M.pick_new_upstream(self, failed_upstreams)
  -- interface method to get a new upstream
  -- has to be overridden by an affinity mode
  ngx.log(ngx.ERR, "[BUG] Failed to pick new upstream as no implementation has been provided!")
  return nil, nil
end

local function should_set_cookie(self)
  if self.cookie_session_affinity.locations and ngx.var.host then
    local locs = self.cookie_session_affinity.locations[ngx.var.host]
    if locs == nil then
      -- Based off of wildcard hostname in ../certificate.lua
      local wildcard_host, _, err = ngx.re.sub(ngx.var.host, "^[^\\.]+\\.", "*.", "jo")
      if err then
        ngx.log(ngx.ERR, "error: ", err);
      elseif wildcard_host then
        locs = self.cookie_session_affinity.locations[wildcard_host]
      end
    end

    if locs ~= nil then
      for _, path in pairs(locs) do
        if ngx.var.location_path == path then
          return true
        end
      end
    end
  end

  return false
end

function _M.balance(self)
  local upstream_from_cookie

  local key = self:get_routing_key()
  if key then
    upstream_from_cookie = self.instance:find(key)
  end

  local last_failure = self.get_last_failure()
  local should_pick_new_upstream = last_failure ~= nil and self.cookie_session_affinity.change_on_failure or
    upstream_from_cookie == nil

  if not should_pick_new_upstream then
    return upstream_from_cookie
  end

  local new_upstream

  new_upstream, key = self:pick_new_upstream(get_failed_upstreams())
  if not new_upstream then
    ngx.log(ngx.WARN, string.format("failed to get new upstream; using upstream %s", new_upstream))
  elseif should_set_cookie(self) then
    self:set_routing_key(key)
  end

  return new_upstream
end

function _M.sync(self, backend)
  local changed = false

  -- check and reinit affinity mode before syncing the balancer which will reinit the nodes
  if self.affinity_mode ~= backend.sessionAffinityConfig.mode then
    changed = true
    init_affinity_mode(self, backend)
  end

  -- reload balancer nodes
  balancer_resty.sync(self, backend)

  -- Reload the balancer if any of the annotations have changed.
  changed = changed or not util.deep_compare(
    self.cookie_session_affinity,
    backend.sessionAffinityConfig.cookieSessionAffinity
  )

  if not changed then
    return
  end

  ngx_log(INFO, string_format("[%s] nodes have changed for backend %s", self.name, backend.name))

  self.cookie_session_affinity = backend.sessionAffinityConfig.cookieSessionAffinity
end

return _M
