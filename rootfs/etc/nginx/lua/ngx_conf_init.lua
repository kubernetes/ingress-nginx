local cjson = require("cjson.safe")

collectgarbage("collect")
local f = io.open("/etc/nginx/lua/cfg.json", "r")
local content = f:read("*a")
f:close()
local configfile = cjson.decode(content)

local luaconfig = ngx.shared.luaconfig
luaconfig:set("enablemetrics", configfile.enable_metrics)
luaconfig:set("use_forwarded_headers", configfile.use_forwarded_headers)
-- init modules
local ok, res
ok, res = pcall(require, "lua_ingress")
if not ok then
  error("require failed: " .. tostring(res))
else
  lua_ingress = res
  lua_ingress.set_config(configfile)
end
ok, res = pcall(require, "configuration")
if not ok then
  error("require failed: " .. tostring(res))
else
  configuration = res
  if not configfile.listen_ports.status_port then
    error("required status port not found")
  end
  configuration.prohibited_localhost_port = configfile.listen_ports.status_port
end
ok, res = pcall(require, "balancer")
if not ok then
  error("require failed: " .. tostring(res))
else
  balancer = res
end
if configfile.enable_metrics then
    ok, res = pcall(require, "monitor")
    if not ok then
        error("require failed: " .. tostring(res))
    else
        monitor = res
    end
end
ok, res = pcall(require, "certificate")
if not ok then
  error("require failed: " .. tostring(res))
else
  certificate = res
  if configfile.enable_ocsp then
    certificate.is_ocsp_stapling_enabled = configfile.enable_ocsp
  end
end

if configfile.enable_arxignis then
  local mlcache = require "resty.mlcache"
  local arxignis_cache, err = mlcache.new("arxignis_cache", "arxignis_cache", {
    lru_size = 50000,
    ttl = 800,
    neg_ttl = 10,
    ipc_shm = "arxignis_cache"
    })
    if err then
      error("Failed to create arxignis cache: " .. tostring(err))
    end
    _G.arxignis_cache = arxignis_cache
end
