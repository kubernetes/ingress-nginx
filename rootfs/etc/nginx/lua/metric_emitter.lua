local socket = ngx.socket.udp
local cjson = require('cjson')
local assert = assert

local _M = {
  queue = {}
}

local function flush_queue()
  ngx.log(ngx.INFO, "******** Flushing queue: " .. ngx.get_phase())
  socket = assert(socket())
  assert(socket:setpeername("127.0.0.1", 8000))
  for _, v in ipairs(_M.queue) do 
    assert(socket:send(v))
  end
  assert(socket:close())

end

function _M.call()
  ngx.log(ngx.INFO, "******** Reaches call ")
  local current_phase = ngx.get_phase()
  if current_phase == "log" then
    ngx.log(ngx.INFO, "******** Reaches log phase ")

    -- Create JSON Metrics Payload  --
    local rjson = cjson.encode({
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
      upstreamIP           = ngx.var.upstream_addr,
      upstreamResponseTime = ngx.var.upstream_response_time,
      upstreamStatus       = ngx.var.upstream_status,
      namespace            = ngx.var.namespace,
      ingress              = ngx.var.ingress_name,
      service              = ngx.var.service_name
    })

    ngx.log(ngx.INFO, "******** JSON structure " .. rjson)
    table.insert(_M.queue, rjson)

    local ok, err = ngx.timer.at(0, flush_queue)
    if not ok then
      ngx.log(ngx.ERR, "failed to create timer: ", err)
      return
    end
    ngx.log(ngx.INFO, "******** Sent to queue ")

  end

end

return _M
