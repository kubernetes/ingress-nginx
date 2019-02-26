local _M = {}

local seeds = {}
local original_randomseed = math.randomseed

local function get_seed_from_urandom()
  local seed
  local frandom, err = io.open("/dev/urandom", "rb")
  if not frandom then
    ngx.log(ngx.WARN, 'failed to open /dev/urandom: ', err)
    return nil
  end

  local str = frandom:read(4)
  frandom:close()
  if not str then
    ngx.log(ngx.WARN, 'failed to read data from /dev/urandom')
    return nil
  end

  seed = 0
  for i = 1, 4 do
      seed = 256 * seed + str:byte(i)
  end

  return seed
end

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
  local seed = get_seed_from_urandom()
  if not seed then
    ngx.log(ngx.WARN, 'failed to get seed from urandom')
    seed = ngx.now() * 1000 + ngx.worker.pid()
  end
  math.randomseed(seed)
end

function _M.init_worker()
  randomseed()
end

return _M
