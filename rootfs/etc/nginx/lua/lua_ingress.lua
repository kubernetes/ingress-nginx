local _M = {}

local seeds = {}
local original_randomseed = math.randomseed
math.randomseed = function(seed)
  local pid = ngx.worker.pid()

  if seeds[pid] then
    ngx.log(ngx.WARN,
      string.format("ignoring math.randomseed(%d) since PRNG is already seeded for worker %d", seed, pid))
    return
  end

  original_randomseed(seed)
  seeds[pid] = seed
end

local function randomseed()
  math.randomseed(ngx.time() + ngx.worker.pid())
end

function _M.init_worker()
  randomseed()
end

return _M
