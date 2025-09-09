local cjson = require("cjson.safe")

local f = io.open("/etc/nginx/lua/cfg.json", "r")
local content = f:read("*a")
f:close()
local configfile = cjson.decode(content)

local lua_ingress = require("lua_ingress")
local balancer = require("balancer")
local monitor = require("monitor")
lua_ingress.init_worker()
balancer.init_worker()
if configfile.enable_metrics and configfile.monitor_batch_max_size then
  monitor.init_worker(configfile.monitor_batch_max_size)
end

if configfile.enable_arxignis then
  local worker = require "resty.arxignis.worker"
  ngx.log(ngx.DEBUG, "Starting flush timers " .. ngx.worker.id())
    worker.start_flush_timers({
        ARXIGNIS_API_URL = os.getenv("ARXIGNIS_API_URL"),
        ARXIGNIS_API_KEY = os.getenv("ARXIGNIS_API_KEY")
    })
end
