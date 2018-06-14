local socket = ngx.socket.udp
local cjson = require('cjson')
local defer = require('defer')
local assert = assert

local _M = {}

local function send_data(jsonData)
  local s = assert(socket())
  assert(s:setpeername("127.0.0.1", 8000))
  assert(s:send(jsonData))
  assert(s:close())
end

function _M.encode_nginx_stats()
  return cjson.encode({
    host = ngx.var.host or "-",
    status = ngx.var.status or "-",
    remoteAddr = ngx.var.remote_addr or "-",
    realIpAddr = ngx.var.realip_remote_addr or "-",
    remoteUser = ngx.var.remote_user or "-",
    bytesSent = tonumber(ngx.var.bytes_sent) or -1,
    protocol = ngx.var.server_protocol or "-",
    method = ngx.var.request_method or "-",
    uri = ngx.var.uri or "-",
    requestLength = tonumber(ngx.var.request_length) or -1,
    requestTime = tonumber(ngx.var.request_time) or -1,
    upstreamName = ngx.var.proxy_upstream_name or "-",
    upstreamIP = ngx.var.upstream_addr or "-",
    upstreamResponseTime = tonumber(ngx.var.upstream_response_time) or -1,
    upstreamStatus = ngx.var.upstream_status or "-",
    namespace = ngx.var.namespace or "-",
    ingress = ngx.var.ingress_name or "-",
    service = ngx.var.service_name or "-",
  })
end

function _M.call()
  local ok, err = defer.to_timer_phase(send_data, _M.encode_nginx_stats())
  if not ok then
    ngx.log(ngx.ERR, "failed to defer send_data to timer phase: ", err)
    return
  end
end

return _M
