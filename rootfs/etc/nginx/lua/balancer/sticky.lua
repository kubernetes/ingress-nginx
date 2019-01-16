local balancer_resty = require("balancer.resty")
local resty_chash = require("resty.chash")
local util = require("util")
local ck = require("resty.cookie")

local _M = balancer_resty:new({ factory = resty_chash, name = "sticky" })

function _M.new(self, backend)
  local nodes = util.get_nodes(backend.endpoints)
  local digest_func = util.md5_digest
  if backend["sessionAffinityConfig"]["cookieSessionAffinity"]["hash"] == "sha1" then
    digest_func = util.sha1_digest
  end

  local o = {
    instance = self.factory:new(nodes),
    cookie_name = backend["sessionAffinityConfig"]["cookieSessionAffinity"]["name"] or "route",
    cookie_expires = backend["sessionAffinityConfig"]["cookieSessionAffinity"]["expires"],
    cookie_max_age = backend["sessionAffinityConfig"]["cookieSessionAffinity"]["maxage"],
    cookie_path = backend["sessionAffinityConfig"]["cookieSessionAffinity"]["path"],
    cookie_locations = backend["sessionAffinityConfig"]["cookieSessionAffinity"]["locations"],
    digest_func = digest_func,
    traffic_shaping_policy = backend.trafficShapingPolicy,
    alternative_backends = backend.alternativeBackends,
  }
  setmetatable(o, self)
  self.__index = self
  return o
end

local function encrypted_endpoint_string(self, endpoint_string)
  local encrypted, err = self.digest_func(endpoint_string)
  if err ~= nil then
    ngx.log(ngx.ERR, err)
  end

  return encrypted
end

local function set_cookie(self, value)
  local cookie, err = ck:new()
  if not cookie then
    ngx.log(ngx.ERR, err)
  end

  local cookie_path = self.cookie_path
  if not cookie_path then
    cookie_path = ngx.var.location_path
  end

  local cookie_data = {
    key = self.cookie_name,
    value = value,
    path = cookie_path,
    httponly = true,
    secure = ngx.var.https == "on",
  }

  if self.cookie_expires and self.cookie_expires ~= "" then
      cookie_data.expires = ngx.cookie_time(ngx.time() + tonumber(self.cookie_expires))
  end

  if self.cookie_max_age and self.cookie_max_age ~= "" then
    cookie_data.max_age = tonumber(self.cookie_max_age)
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

  local key = cookie:get(self.cookie_name)
  if not key then
    local random_str = string.format("%s.%s", ngx.now(), ngx.worker.pid())
    key = encrypted_endpoint_string(self, random_str)

    if self.cookie_locations then
      local locs = self.cookie_locations[ngx.var.host]
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

return _M
