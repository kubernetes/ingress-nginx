local balancer = require("balancer")
local monitor = require("monitor")
local plugins = require("plugins")

local luaconfig = ngx.shared.luaconfig
local enablemetrics = luaconfig:get("enablemetrics")


balancer.log()

if enablemetrics then
    monitor.call()
end
plugins.run()