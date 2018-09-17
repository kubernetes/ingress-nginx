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
    digest_func = digest_func,
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

  local ok
  ok, err = cookie:set({
    key = self.cookie_name,
    value = value,
    path = ngx.var.location_path,
    domain = ngx.var.host,
    httponly = true,
  })
  if not ok then
    ngx.log(ngx.ERR, err)
  end
end

local function pick_random(instance)
  local index = math.random(instance.npoints)
  return instance:next(index)
end

function _M.balance(self)
  local cookie, err = ck:new()
  if not cookie then
    ngx.log(ngx.ERR, err)
    return pick_random(self.instance)
  end

  local key = cookie:get(self.cookie_name)
  if not key then
    local tmp_endpoint_string = pick_random(self.instance)
    key = encrypted_endpoint_string(self, tmp_endpoint_string)
    set_cookie(self, key)
  end

  return self.instance:find(key)
end

return _M
