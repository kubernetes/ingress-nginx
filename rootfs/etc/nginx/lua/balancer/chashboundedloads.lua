-- Implements Consistent Hashing with Bounded Loads based on the paper [1].
-- For the specified hash-balance-factor, requests to any upstream host are capped
-- at hash_balance_factor times the average number of requests across the cluster.
-- When a request arrives for an upstream host that is currently serving at its max capacity,
-- linear probing is used to identify the next eligible host.
--
-- This is an O(N) algorithm, unlike other load balancers. Using a lower hash-balance-factor
-- results in more hosts being probed, so use a higher value if you require better performance.
--
-- [1]: https://arxiv.org/abs/1608.01350

local resty_roundrobin = require("resty.roundrobin")
local resty_chash = require("resty.chash")
local setmetatable = setmetatable
local lrucache = require("resty.lrucache")

local util = require("util")
local split = require("util.split")
local reverse_table = util.reverse_table

local string_format = string.format
local INFO = ngx.INFO
local ngx_ERR = ngx.ERR
local ngx_WARN = ngx.WARN
local ngx_log = ngx.log
local math_ceil = math.ceil
local ipairs = ipairs
local ngx = ngx

local DEFAULT_HASH_BALANCE_FACTOR = 2

local HOST_SEED = util.get_host_seed()

-- Controls how many "tenants" we'll keep track of
-- to avoid routing them to alternative_backends
-- as they were already consistently routed to some endpoint.
-- Lowering this value will increases the chances of more
-- tenants being routed to alternative_backends.
-- Similarly, increasing this value will keep more tenants
-- consistently routed to the same endpoint in the main backend.
local SEEN_LRU_SIZE = 1000

local _M = {}

local function incr_req_stats(self, endpoint)
  if not self.requests_by_endpoint[endpoint] then
    self.requests_by_endpoint[endpoint] = 1
  else
    self.requests_by_endpoint[endpoint] = self.requests_by_endpoint[endpoint] + 1
  end
  self.total_requests = self.total_requests + 1
end

local function decr_req_stats(self, endpoint)
  if self.requests_by_endpoint[endpoint] then
    self.requests_by_endpoint[endpoint] = self.requests_by_endpoint[endpoint] - 1
    if self.requests_by_endpoint[endpoint] == 0 then
      self.requests_by_endpoint[endpoint] = nil
    end
  end
  self.total_requests = self.total_requests - 1
end

local function get_hash_by_value(self)
  if not ngx.ctx.chash_hash_by_value then
    ngx.ctx.chash_hash_by_value = util.generate_var_value(self.hash_by)
  end

  local v = ngx.ctx.chash_hash_by_value
  if v == "" then
    return nil
  end
  return v
end

local function endpoint_eligible(self, endpoint)
  -- (num_requests * hash-balance-factor / num_servers)
  local allowed = math_ceil(
    (self.total_requests + 1) * self.balance_factor / self.total_endpoints)
  local current = self.requests_by_endpoint[endpoint]
  if current == nil then
    return true, 0, allowed
  else
    return current < allowed, current, allowed
  end
end

local function update_balance_factor(self, backend)
  local balance_factor = backend["upstreamHashByConfig"]["upstream-hash-by-balance-factor"]
  if balance_factor and balance_factor <= 1 then
    ngx_log(ngx_WARN,
    "upstream-hash-by-balance-factor must be > 1. Forcing it to the default value of ",
      DEFAULT_HASH_BALANCE_FACTOR)
    balance_factor = DEFAULT_HASH_BALANCE_FACTOR
  end
  self.balance_factor = balance_factor or DEFAULT_HASH_BALANCE_FACTOR
end

local function normalize_endpoints(endpoints)
  local b = {}
  for i, endpoint in ipairs(endpoints) do
    b[i] = string_format("%s:%s", endpoint.address, endpoint.port)
  end
  return b
end

local function update_endpoints(self, endpoints)
  self.endpoints = endpoints
  self.endpoints_reverse = reverse_table(endpoints)
  self.total_endpoints = #endpoints
  self.ring_seed = util.array_mod(HOST_SEED, self.total_endpoints)
end

function _M.is_affinitized(self)
  -- alternative_backends might contain a canary backend that gets a percentage of traffic.
  -- If a tenant has already been consistently routed to a endpoint, we want to stick to that
  -- to keep a higher cache ratio, rather than routing it to an alternative backend.
  -- This would mean that alternative backends (== canary) would mostly be seeing "new" tenants.

  if not self.alternative_backends or not self.alternative_backends[1] then
    return false
  end

  local hash_by_value = get_hash_by_value(self)
  if not hash_by_value then
    return false
  end

  return self.seen_hash_by_values:get(hash_by_value) ~= nil
end

function _M.new(self, backend)
  local nodes = util.get_nodes(backend.endpoints)

  local complex_val, err =
    util.parse_complex_value(backend["upstreamHashByConfig"]["upstream-hash-by"])
  if err ~= nil then
    ngx_log(ngx_ERR, "could not parse the value of the upstream-hash-by: ", err)
  end

  local o = {
    name = "chashboundedloads",

    chash = resty_chash:new(nodes),
    roundrobin = resty_roundrobin:new(nodes),
    alternative_backends = backend.alternativeBackends,
    hash_by = complex_val,

    requests_by_endpoint = {},
    total_requests = 0,
    seen_hash_by_values = lrucache.new(SEEN_LRU_SIZE)
  }

  update_endpoints(o, normalize_endpoints(backend.endpoints))
  update_balance_factor(o, backend)

  setmetatable(o, self)
  self.__index = self
  return o
end

function _M.sync(self, backend)
  self.alternative_backends = backend.alternativeBackends

  update_balance_factor(self, backend)

  local new_endpoints = normalize_endpoints(backend.endpoints)

  if util.deep_compare(self.endpoints, new_endpoints) then
    ngx_log(INFO, "endpoints did not change for backend", backend.name)
    return
  end

  ngx_log(INFO, string_format("[%s] endpoints have changed for backend %s",
    self.name, backend.name))

  update_endpoints(self, new_endpoints)

  local nodes = util.get_nodes(backend.endpoints)
  self.chash:reinit(nodes)
  self.roundrobin:reinit(nodes)

  self.seen_hash_by_values = lrucache.new(SEEN_LRU_SIZE)

  ngx_log(INFO, string_format("[%s] nodes have changed for backend %s", self.name, backend.name))
end

function _M.balance(self)
  local hash_by_value = get_hash_by_value(self)

  -- Tenant key not available, falling back to round-robin
  if not hash_by_value then
    local endpoint = self.roundrobin:find()
    ngx.var.chashbl_debug = "fallback_round_robin"
    return endpoint
  end

  self.seen_hash_by_values:set(hash_by_value, true)

  local tried_endpoints
  if not ngx.ctx.balancer_chashbl_tried_endpoints then
    tried_endpoints = {}
    ngx.ctx.balancer_chashbl_tried_endpoints = tried_endpoints
  else
    tried_endpoints = ngx.ctx.balancer_chashbl_tried_endpoints
  end

  local first_endpoint = self.chash:find(hash_by_value)
  local index = self.endpoints_reverse[first_endpoint]

  -- By design, resty.chash always points to the same element of the ring,
  -- regardless of the environment. In this algorithm, we want the consistency
  -- to be "seeded" based on the host where it's running.
  -- That's how both Envoy and Haproxy implement this.
  -- For convenience, we keep resty.chash but manually introduce the seed.
  index = util.array_mod(index + self.ring_seed, self.total_endpoints)

  for i=0, self.total_endpoints-1 do
    local j = util.array_mod(index + i, self.total_endpoints)
    local endpoint = self.endpoints[j]

    if not tried_endpoints[endpoint] then
      local eligible, current, allowed = endpoint_eligible(self, endpoint)

      if eligible then
        ngx.var.chashbl_debug = string_format(
          "attempt=%d score=%d allowed=%d total_requests=%d hash_by_value=%s",
          i, current, allowed, self.total_requests, hash_by_value)

        incr_req_stats(self, endpoint)
        tried_endpoints[endpoint] = true
        return endpoint
      end
    end
  end

  -- Normally, this case should never be reach out because with balance_factor > 1
  -- there should always be an eligible endpoint.
  -- This would get reached only if the number of endpoints is less or equal
  -- than max Nginx retries and tried_endpoints contains all endpoints.
  incr_req_stats(self, first_endpoint)
  ngx.var.chashbl_debug = "fallback_first_endpoint"
  return first_endpoint
end

function _M.after_balance(self)
  local tried_upstreams = split.split_upstream_var(ngx.var.upstream_addr)
  if (not tried_upstreams) or (not get_hash_by_value(self)) then
    return
  end

  for _, addr in ipairs(tried_upstreams) do
    decr_req_stats(self, addr)
  end
end

return _M
