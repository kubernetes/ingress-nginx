local util = require("util")

local timer_started = false
local queue = {}
local MAX_QUEUE_SIZE = 10000

local _M = {}

local function flush_queue(premature)
  -- TODO Investigate if we should actually still flush the queue when we're
  -- shutting down.
  if premature then return end

  local current_queue = queue
  queue = {}
  timer_started = false

  for _,v in ipairs(current_queue) do
    v.func(unpack(v.args))
  end
end

-- `to_timer_phase` will enqueue a function that will be executed in a timer
-- context, at a later point in time. The purpose is that some APIs (such as
-- sockets) are not available during some nginx request phases (such as the
-- logging phase), but are available for use in timers. There are no ordering
-- guarantees for when a function will be executed.
function _M.to_timer_phase(func, ...)
  if ngx.get_phase() == "timer" then
    func(...)
    return true
  end

  if #queue >= MAX_QUEUE_SIZE then
    ngx.log(ngx.ERR, "deferred timer queue full")
    return nil, "deferred timer queue full"
  end

  table.insert(queue, { func = func, args = {...} })
  if not timer_started then
    local ok, err = ngx.timer.at(0, flush_queue)
    if ok then
      -- unfortunately this is to deal with tests - when running unit tests, we
      -- dont actually run the timer, we call the function inline
      if util.tablelength(queue) > 0 then
        timer_started = true
      end
    else
      local msg = "failed to create timer: " .. tostring(err)
      ngx.log(ngx.ERR, msg)
      return nil, msg
    end
  end
  return true
end

return _M
