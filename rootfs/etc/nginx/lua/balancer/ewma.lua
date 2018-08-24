-- Original Authors: Shiv Nagarajan & Scott Francis
-- Accessed: March 12, 2018
-- Inspiration drawn from:
-- https://github.com/twitter/finagle/blob/1bc837c4feafc0096e43c0e98516a8e1c50c4421
--   /finagle-core/src/main/scala/com/twitter/finagle/loadbalancer/PeakEwma.scala


local resty_lock = require("resty.lock")
local util = require("util")
local split = require("util.split")

local DECAY_TIME = 10 -- this value is in seconds
local LOCK_KEY = ":ewma_key"
local PICK_SET_SIZE = 2

local ewma_lock = resty_lock:new("locks", {timeout = 0, exptime = 0.1})

local _M = { name = "ewma" }

local function lock(upstream)
  local _, err = ewma_lock:lock(upstream .. LOCK_KEY)
  if err then
    if err ~= "timeout" then
      ngx.log(ngx.ERR, string.format("EWMA Balancer failed to lock: %s", tostring(err)))
    end
  end

  return err
end

local function unlock()
  local ok, err = ewma_lock:unlock()
  if not ok then
    ngx.log(ngx.ERR, string.format("EWMA Balancer failed to unlock: %s", tostring(err)))
  end

  return err
end

local function decay_ewma(ewma, last_touched_at, rtt, now)
  local td = now - last_touched_at
  td = (td > 0) and td or 0
  local weight = math.exp(-td/DECAY_TIME)

  ewma = ewma * weight + rtt * (1.0 - weight)
  return ewma
end

local function get_or_update_ewma(upstream, rtt, update)
  local lock_err = nil
  if update then
    lock_err = lock(upstream)
  end
  local ewma = ngx.shared.balancer_ewma:get(upstream) or 0
  if lock_err ~= nil then
    return ewma, lock_err
  end

  local now = ngx.now()
  local last_touched_at = ngx.shared.balancer_ewma_last_touched_at:get(upstream) or 0
  ewma = decay_ewma(ewma, last_touched_at, rtt, now)

  if not update then
    return ewma, nil
  end

  local success, err, forcible = ngx.shared.balancer_ewma_last_touched_at:set(upstream, now)
  if not success then
    ngx.log(ngx.WARN, "balancer_ewma_last_touched_at:set failed " .. err)
  end
  if forcible then
    ngx.log(ngx.WARN, "balancer_ewma_last_touched_at:set valid items forcibly overwritten")
  end

  success, err, forcible = ngx.shared.balancer_ewma:set(upstream, ewma)
  if not success then
    ngx.log(ngx.WARN, "balancer_ewma:set failed " .. err)
  end
  if forcible then
    ngx.log(ngx.WARN, "balancer_ewma:set valid items forcibly overwritten")
  end

  unlock()
  return ewma, nil
end


local function score(upstream)
  -- Original implementation used names
  -- Endpoints don't have names, so passing in IP:Port as key instead
  local upstream_name = upstream.address .. ":" .. upstream.port
  return get_or_update_ewma(upstream_name, 0, false)
end

-- implementation similar to https://en.wikipedia.org/wiki/Fisher%E2%80%93Yates_shuffle
-- or https://en.wikipedia.org/wiki/Random_permutation
-- loop from 1 .. k
-- pick a random value r from the remaining set of unpicked values (i .. n)
-- swap the value at position i with the value at position r
local function shuffle_peers(peers, k)
  for i=1, k do
    local rand_index = math.random(i,#peers)
    peers[i], peers[rand_index] = peers[rand_index], peers[i]
  end
  -- peers[1 .. k] will now contain a randomly selected k from #peers
end

local function pick_and_score(peers, k)
  shuffle_peers(peers, k)
  local lowest_score_index = 1
  local lowest_score = score(peers[lowest_score_index])
  for i = 2, k do
    local new_score = score(peers[i])
    if new_score < lowest_score then
      lowest_score_index, lowest_score = i, new_score
    end
  end
  return peers[lowest_score_index]
end

function _M.balance(self)
  local peers = self.peers
  local endpoint = peers[1]

  if #peers > 1 then
    local k = (#peers < PICK_SET_SIZE) and #peers or PICK_SET_SIZE
    local peer_copy = util.deepcopy(peers)
    endpoint = pick_and_score(peer_copy, k)
  end

  -- TODO(elvinefendi) move this processing to _M.sync
  return endpoint.address .. ":" .. endpoint.port
end

function _M.after_balance(_)
  local response_time = tonumber(split.get_first_value(ngx.var.upstream_response_time)) or 0
  local connect_time = tonumber(split.get_first_value(ngx.var.upstream_connect_time)) or 0
  local rtt = connect_time + response_time
  local upstream = split.get_first_value(ngx.var.upstream_addr)

  if util.is_blank(upstream) then
    return
  end
  get_or_update_ewma(upstream, rtt, true)
end

function _M.sync(self, backend)
  local changed = not util.deep_compare(self.peers, backend.endpoints)
  if not changed then
    return
  end

  self.peers = backend.endpoints

  -- TODO: Reset state of EWMA per backend
  ngx.shared.balancer_ewma:flush_all()
  ngx.shared.balancer_ewma_last_touched_at:flush_all()
end

function _M.new(self, backend)
  local o = { peers = backend.endpoints }
  setmetatable(o, self)
  self.__index = self
  return o
end

return _M
