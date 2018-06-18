local socket = ngx.socket.tcp
local cjson = require('cjson')
local assert = assert

local metrics_batch = {}
-- if an Nginx worker processes more than (MAX_BATCH_SIZE/FLUSH_INTERVAL) RPS then it will start dropping metrics
local MAX_BATCH_SIZE = 10000
local FLUSH_INTERVAL = 1 -- second

local _M = {}

local function send(payload)
  local s = assert(socket())
  assert(s:connect("unix:/tmp/prometheus-nginx.socket"))
  assert(s:send(payload))
  assert(s:close())
end

local function metrics()
  return {
    host = ngx.var.host or "-",
    namespace = ngx.var.namespace or "-",
    ingress = ngx.var.ingress_name or "-",
    service = ngx.var.service_name or "-",
    path = ngx.var.location_path or "-",

    method = ngx.var.request_method or "-",
    status = ngx.var.status or "-",
    requestLength = tonumber(ngx.var.request_length) or -1,
    requestTime = tonumber(ngx.var.request_time) or -1,
    responseLength = tonumber(ngx.var.bytes_sent) or -1,

    endpoint = ngx.var.upstream_addr or "-",
    upstreamLatency = tonumber(ngx.var.upstream_connect_time) or -1,
    upstreamResponseTime = tonumber(ngx.var.upstream_response_time) or -1,
    upstreamResponseLength = tonumber(ngx.var.upstream_response_length) or -1,
    upstreamStatus = ngx.var.upstream_status or "-",
  }
end

local function flush(premature)
  if premature then
    return
  end

  if #metrics_batch == 0 then
    return
  end

  local current_metrics_batch = metrics_batch
  metrics_batch = {}

  local ok, payload = pcall(cjson.encode, current_metrics_batch)
  if not ok then
    ngx.log(ngx.ERR, "error while encoding metrics: " .. tostring(payload))
    return
  end

  send(payload)
end

function _M.init_worker()
  local _, err = ngx.timer.every(FLUSH_INTERVAL, flush)
  if err then
    ngx.log(ngx.ERR, string.format("error when setting up timer.every: %s", tostring(err)))
  end
end

function _M.call()
  if #metrics_batch >= MAX_BATCH_SIZE then
    ngx.log(ngx.WARN, "omitting metrics for the request, current batch is full")
    return
  end

  table.insert(metrics_batch, metrics())
end

if _TEST then
  _M.flush = flush
  _M.get_metrics_batch = function() return metrics_batch end
end

return _M
