local socket = require('socket')
local cjson = require('cjson')

local _M = {}

function _M.call()
  local current_phase = ngx.get_phase()
  if current_phase == "log" then

    -- Initialize UDP Socket --
    udp = assert(socket.udp())
    asssert(udp:setpeername("127.0.0.1", 8000))

    -- Create JSON Metrics Payload  --
    local json = cjson.encode({
      host                 = ngx.var.host,
      status               = ngx.var.status,
      time                 = ngx.localtime(),
      remoteAddr           = ngx.var.realip_remote_addr,
      remoteUser           = ngx.var.remote_user,
      bytesSent            = ngx.var.bytes_sent,
      protocol             = ngx.var.server_protocol,
      method               = ngx.var.request_method,
      path                 = ngx.var.uri,
      requestTime          = ngx.var.request_time,
      requestLength        = ngx.var.request_length,
      duration             = ngx.var.request_time,
      upstreamName         = ngx.var.upstream,
      upstreamIP           = ngx.var.ngx.var.upstream_addr,
      upstreamResponseTime = ngx.var.upstream_response_time,
      upstreamStatus       = ngx.var.upstream_status,
      namespace__          = ngx.var.namespace,
      ingress              = ngx.var.ingress_name,
      service              = ngx.var.service_name
    })

    assert(udp:send(json))

    -- Close UDP Socket --
    assert(udp:close())
  end

end

return _M
