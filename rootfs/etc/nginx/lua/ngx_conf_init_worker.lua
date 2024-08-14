local function initialize_worker(enablemetrics, monitorbatchsize)
  local lua_ingress = require("lua_ingress")
  local balancer = require("balancer")
  local plugins = require("plugins")
  local monitor = require("monitor")
  lua_ingress.init_worker()
  balancer.init_worker()
  if enablemetrics then
    monitor.init_worker(monitorbatchsize)
  end
  plugins.run()
end

return { initialize_worker = initialize_worker }