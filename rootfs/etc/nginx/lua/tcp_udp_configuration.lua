-- this is the Lua representation of TCP/UDP Configuration
local tcp_udp_configuration_data = ngx.shared.tcp_udp_configuration_data

local _M = {
  nameservers = {}
}

function _M.get_backends_data()
  return tcp_udp_configuration_data:get("backends")
end

function _M.call()
  local sock, err = ngx.req.socket(true)
  if not sock then
    ngx.log(ngx.ERR, "failed to get raw req socket: ", err)
    ngx.say("error: ", err)
    return
  end

  local reader = sock:receiveuntil("\r\n")
  local backends, err_read = reader()
  if not backends then
    ngx.log(ngx.ERR, "failed TCP/UDP dynamic-configuration:", err_read)
    ngx.say("error: ", err_read)
    return
  end

  if backends == nil or backends == "" then
    return
  end

  local success, err_conf = tcp_udp_configuration_data:set("backends", backends)
  if not success then
    ngx.log(ngx.ERR, "dynamic-configuration: error updating configuration: " .. tostring(err_conf))
    ngx.say("error: ", err_conf)
    return
  end
end

return _M
