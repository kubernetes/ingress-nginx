local balancer_resty = require("balancer.resty")
local resty_chash = require("resty.chash")
local util = require("util")
local ck = require("resty.cookie")
local math = require("math")

local _M = balancer_resty:new({ factory = resty_chash, name = "sticky" })
local DEFAULT_COOKIE_NAME = "route"

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

function _M.balance(self)
  local cookie, err = ck:new()
  if not cookie then
    ngx.log(ngx.ERR, "error while initializing cookie: " .. tostring(err))
    return
  end

  local key = cookie:get(self:cookie_name())
  if not key then
    key = string.format("%s.%s.%s", ngx.now(), ngx.worker.pid(), math.random(999999))

    if self.cookie_session_affinity.locations then
      local locs = self.cookie_session_affinity.locations[ngx.var.host]
      if locs ~= nil then
        for _, path in pairs(locs) do
          if ngx.var.location_path == path then
            set_cookie(self, key)
            break
          end
        end
      end
    end
  end

  return self.instance:find(key)
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

  self.cookie_session_affinity = backend.sessionAffinityConfig.cookieSessionAffinity
end

return _M
