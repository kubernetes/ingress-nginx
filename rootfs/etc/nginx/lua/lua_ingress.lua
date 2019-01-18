local _M = {}

local seeds = {}
local original_randomseed = math.randomseed

local function get_seed_from_urandom()
  local seed
  local frandom = io.open("/dev/urandom", "rb")
  if frandom then
    local str = frandom:read(4)
    frandom:close()
    seed = 0
    for i = 1, 4 do
        seed = 256 * seed + str:byte(i)
    end
  end
  return seed
end

math.randomseed = function()
  local pid = ngx.worker.pid()
  local seed = seeds[pid]
  if seed then
    ngx.log(ngx.WARN,
      string.format("ignoring math.randomseed() since PRNG is already seeded for worker %d", pid))
    return
  end

  seed = get_seed_from_urandom()
  if not seed then
    seed = ngx.now() * 1000 + pid
  end
  original_randomseed(seed)
  seeds[pid] = seed
end

function _M.init_worker()
  math.randomseed()
end

return _M
