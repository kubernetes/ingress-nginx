local balancer = require("balancer")
local monitor = require("monitor")

local luaconfig = ngx.shared.luaconfig
local enablemetrics = luaconfig:get("enablemetrics")

balancer.log()

if enablemetrics then
    monitor.call()
end