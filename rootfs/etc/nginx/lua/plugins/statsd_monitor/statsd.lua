local pairs = pairs
local string_format = string.format
local string_len = string.len
local udp = ngx.socket.udp

local util = require("util")
local defer_to_timer = require("plugins.statsd_monitor.defer_to_timer")

local util_tablelength = util.tablelength

local _M = {}
local default_tag_string = "|#"

local METRIC_COUNTER   = "c"
local METRIC_GAUGE     = "g"
local METRIC_HISTOGRAM = "h"
local METRIC_SET       = "s"
local MICROSECONDS     = 1000000

local ENV_TAGS = {
  kube_namespace = os.getenv("POD_NAMESPACE"),
  deploy_stage = os.getenv("DEPLOY_STAGE"),
}

local function create_udp_socket(host, port)
  local sock, sock_err = udp()
  if not sock then
    return nil, sock_err
  end

  local ok, peer_err  = sock:setpeername(host, port)
  if not ok then
    return nil, peer_err
  end

  return sock, nil
end

local function get_udp_socket(host, port)
  local id = string_format("%s:%d", host, port)
  local ctx = ngx.ctx.statsd_sockets
  local err

  if not ctx then
    ctx = {}
    ngx.ctx.statsd_sockets = ctx
  end

  local sock = ctx[id]
  if not sock then
    sock, err = create_udp_socket(host, port)
    if not sock then
      return nil, err
    end

    ctx[id] = sock
  end

  return sock, nil
end

-- gets called once after statsd config is parsed
-- the function expands default tags with environment specific ones
local function expand_default_tags()
  if not _M.config then
    return
  end

  if not _M.config.tags then
    _M.config.tags = {}
  end

  for k, v  in pairs(ENV_TAGS) do
    if v then
      _M.config.tags[k] = v
    end
  end
end

local function generate_tag_string(tags)
  if not tags or util_tablelength(tags) == 0 then
    return ""
  end

  local tag_str = default_tag_string
  for k,v in pairs(tags) do
    if string_len(tag_str) > 2 then
      tag_str = tag_str .. ","
    end
    tag_str = tag_str .. k .. ":" .. v
  end

  return tag_str
end

local function generate_packet(metric, key, value, tags, sampling_rate)
  if sampling_rate == 1 then
    sampling_rate = ""
  else
    sampling_rate = string_format("|@%g", sampling_rate)
  end

  return string_format("%s:%s|%s%s%s", key, tostring(value), metric, sampling_rate, generate_tag_string(tags))
end

local function metric(metric_type, key, value, tags, sample_rate)
  if not value then
    return nil, "no value passed"
  end
  if value == '-' then
    return nil, nil -- don't pass an error to avoid logging to error log
  end

  if not _M.config then
    return true, nil
  end

  local sampling_rate = sample_rate or _M.config.sampling_rate

  if sampling_rate ~= 1 and math.random() > sampling_rate then
    return nil, nil -- don't pass an error to avoid logging to error log
  end

  local packet = generate_packet(metric_type, key, value, tags, sampling_rate)

  local sock, err = get_udp_socket(_M.config.host, _M.config.port)
  if not sock then
    return nil, err
  end

  return sock:send(packet)
end

local function send_metrics(...)
  local ok, err = metric(...)
  if not ok and err then
    ngx.log(ngx.WARN, "failed logging to statsd: " .. tostring(err))
  end
  return ok, err
end

-- to avoid logging everywhere in #metric
local function log_metric(...)
  local err = defer_to_timer.enqueue(send_metrics, ...)
  if err then
    local msg = "failed to log metric: " .. tostring(err)
    ngx.log(ngx.ERR,  msg)
    return nil, msg
  end
  return true
end

-- Statsd module level convenince functions

function _M.increment(key, value, tags, ...)
  return log_metric(METRIC_COUNTER, key, value or 1, tags, ...)
end

function _M.gauge(key, value, tags, ...)
  return log_metric(METRIC_GAUGE, key, value, tags, ...)
end

function _M.histogram(key, value, tags, ...)
  return log_metric(METRIC_HISTOGRAM, key, value, tags, ...)
end

function _M.set(key, value, tags, ...)
  return log_metric(METRIC_SET, key, value, tags, ...)
end

function _M.time(f)
  local start_time = ngx.now()
  local ret = { f() }
  return ret, (ngx.now() - start_time) * MICROSECONDS
end

function _M.measure(key, f, tags)
  local ret, time = _M.time(f)
  _M.histogram(key, time, tags or {})
  return unpack(ret)
end

_M.config = {
  host = os.getenv("STATSD_HOST"),
  port = os.getenv("STATSD_PORT"),
  sampling_rate = 0.1,
  tags = {},
}

if not _M.config.host or not _M.config.port then
  error("STATSD_HOST and STATSD_PORT env variables must be set")
end

expand_default_tags()

if _M.config.tags and util_tablelength(_M.config.tags) > 0 then
  default_tag_string = generate_tag_string(_M.config.tags)
end

return _M
