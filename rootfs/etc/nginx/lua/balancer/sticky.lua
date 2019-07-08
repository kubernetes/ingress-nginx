local balancer_resty = require("balancer.resty")
local resty_chash = require("resty.chash")
local util = require("util")
local ck = require("resty.cookie")
local math = require("math")
local ngx_balancer = require("ngx.balancer")
local split = require("util.split")

local string_format = string.format
local ngx_log = ngx.log
local INFO = ngx.INFO

local _M = balancer_resty:new({ factory = resty_chash, name = "sticky" })
local DEFAULT_COOKIE_NAME = "route"

-- Consider the situation of N upstreams one of which is failing.
-- Then the probability to obtain failing upstream after M iterations would be close to (1/N)**M.
-- For the worst case (2 upstreams; 20 iterations) it would be ~10**(-6)
-- which is much better then ~10**(-3) for 10 iterations.
local MAX_UPSTREAM_CHECKS_COUNT = 20

function _M.cookie_name(self)
  return self.cookie_session_affinity.name or DEFAULT_COOKIE_NAME
end

function _M.new(self, backend)
  local nodes = util.get_nodes(backend.endpoints)

  local o = {
    instance = self.factory:new(nodes),
    traffic_shaping_policy = backend.trafficShapingPolicy,
    alternative_backends = backend.alternativeBackends,
    cookie_session_affinity = backend["sessionAffinityConfig"]["cookieSessionAffinity"]
  }
  setmetatable(o, self)
  self.__index = self
  return o
end

local function set_cookie(self, value)
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

local function pick_new_upstream(self)
  local failed_upstreams = get_failed_upstreams()

  for i = 1, MAX_UPSTREAM_CHECKS_COUNT do
    local key = string.format("%s.%s.%s", ngx.now() + i, ngx.worker.pid(), math.random(999999))

    local new_upstream = self.instance:find(key)

    if not failed_upstreams[new_upstream] then
      return new_upstream, key
    end
  end

  return nil, nil
end

local function should_set_cookie(self)
  if self.cookie_session_affinity.locations then
    local locs = self.cookie_session_affinity.locations[ngx.var.host]
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
  local cookie, err = ck:new()
  if not cookie then
    ngx.log(ngx.ERR, "error while initializing cookie: " .. tostring(err))
    return
  end

  local upstream_from_cookie

  local key = cookie:get(self:cookie_name())
  if key then
    upstream_from_cookie = self.instance:find(key)
  end

  local last_failure = self.get_last_failure()
  local should_pick_new_upstream = last_failure ~= nil and self.cookie_session_affinity.change_on_failure or upstream_from_cookie == nil

  if not should_pick_new_upstream then
    return upstream_from_cookie
  end

  local new_upstream, key = pick_new_upstream(self)
  if not new_upstream then
    ngx.log(ngx.WARN, string.format("failed to get new upstream; using upstream %s", new_upstream))
  elseif should_set_cookie(self) then
    set_cookie(self, key)
  end

  return new_upstream
end

function _M.sync(self, backend)
  balancer_resty.sync(self, backend)

  -- Reload the balancer if any of the annotations have changed.
  local changed = not util.deep_compare(
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
