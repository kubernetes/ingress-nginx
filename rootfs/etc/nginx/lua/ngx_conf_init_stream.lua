local cjson = require("cjson.safe")
collectgarbage("collect")
local f = io.open("/etc/nginx/lua/cfg.json", "r")
local content = f:read("*a")
f:close()
local configfile = cjson.decode(content)
-- init modules
local ok, res
ok, res = pcall(require, "configuration")
if not ok then
  error("require failed: " .. tostring(res))
else
  configuration = res
end
ok, res = pcall(require, "tcp_udp_configuration")
if not ok then
  error("require failed: " .. tostring(res))
else
  tcp_udp_configuration = res
  if not configfile.listen_ports.status_port then
    error("required status port not found")
  end
  tcp_udp_configuration.prohibited_localhost_port = configfile.listen_ports.status_port
end
ok, res = pcall(require, "tcp_udp_balancer")
if not ok then
  error("require failed: " .. tostring(res))
else
  tcp_udp_balancer = res
end
