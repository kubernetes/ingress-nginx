local statsd = require("plugins.statsd_monitor.statsd")
local defer_to_timer = require("plugins.statsd_monitor.defer_to_timer")
local split = require("util.split")

local _M = {}

local function send_response_data(upstream_state, client_state)
  local status_class
  if upstream_state.status then
    for i, status in ipairs(upstream_state.status) do
      -- TODO: link this with the zones when we use openresty-upstream
      if status == '-' then
        status = 'ngx_error'
        status_class = 'ngx_error'
      else
        status_class = string.sub(status, 0, 1) .. "xx"
      end

      statsd.increment('nginx.upstream.response', 1, {
        status=status,
        status_class=status_class,
        upstream_name=client_state.upstream_name
      })

      statsd.histogram('nginx.upstream.response_time',
        upstream_state.response_time[i], {
          upstream_name=client_state.upstream_name
      })
    end
  end

  status_class = string.sub(client_state.status, 0, 1) .. "xx"
  statsd.increment('nginx.client.response', 1, {
    status=client_state.status,
    status_class=status_class,
    upstream_name=client_state.upstream_name
  })

  statsd.histogram('nginx.client.request_time', client_state.request_time, {
    upstream_name=client_state.upstream_name
  })
end

function _M.init_worker()
  defer_to_timer.init_worker()
end

function _M.log()
  local status, status_err = split.split_upstream_var(ngx.var.upstream_status)
  if status_err then
    return nil, status_err
  end

  local addrs, addrs_err = split.split_upstream_addr(ngx.var.upstream_addr)
  if addrs_err then
    return nil, addrs_err
  end

  local response_time, rt_err = split.split_upstream_var(ngx.var.upstream_response_time)
  if rt_err then
    return nil, rt_err
  end

  local err = defer_to_timer.enqueue(send_response_data, {
      status=status,
      addr=addrs,
      response_time=response_time
    }, {
      status=ngx.var.status,
      request_time=ngx.var.request_time,
      upstream_name=ngx.var.proxy_upstream_name
    })

  if err then
    local msg = "failed to send response data: " .. tostring(err)
    ngx.log(ngx.ERR,  msg)
    return nil, msg
  end
end

return _M
