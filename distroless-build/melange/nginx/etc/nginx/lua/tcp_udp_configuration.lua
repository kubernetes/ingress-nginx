local ngx = ngx
local tostring = tostring
-- this is the Lua representation of TCP/UDP Configuration
local tcp_udp_configuration_data = ngx.shared.tcp_udp_configuration_data

local _M = {}

function _M.get_backends_data()
  return tcp_udp_configuration_data:get("backends")
end

function _M.get_raw_backends_last_synced_at()
  local raw_backends_last_synced_at = tcp_udp_configuration_data:get("raw_backends_last_synced_at")
  if raw_backends_last_synced_at == nil then
    raw_backends_last_synced_at = 1
  end
  return raw_backends_last_synced_at
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

  ngx.update_time()
  local raw_backends_last_synced_at = ngx.time()
  success, err = tcp_udp_configuration_data:set("raw_backends_last_synced_at",
                      raw_backends_last_synced_at)
  if not success then
    ngx.log(ngx.ERR, "dynamic-configuration: error updating when backends sync, " ..
                     "new upstream peers waiting for force syncing: " .. tostring(err))
    ngx.status = ngx.HTTP_BAD_REQUEST
    return
  end
end

return _M
