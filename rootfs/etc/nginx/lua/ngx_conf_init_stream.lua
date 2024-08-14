local function initialize_stream(statusport)
  collectgarbage("collect")

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
    tcp_udp_configuration.prohibited_localhost_port = statusport

  end

  ok, res = pcall(require, "tcp_udp_balancer")
  if not ok then
    error("require failed: " .. tostring(res))
  else
    tcp_udp_balancer = res
  end
end

return { initialize_stream = initialize_stream }