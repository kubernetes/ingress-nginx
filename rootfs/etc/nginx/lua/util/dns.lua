local resolver = require("resty.dns.resolver")
local lrucache = require("resty.lrucache")
local configuration = require("configuration")
local util = require("util")

local _M = {}
local CACHE_SIZE = 10000
local MAXIMUM_TTL_VALUE = 2147483647 -- maximum value according to https://tools.ietf.org/html/rfc2181

local cache, err = lrucache.new(CACHE_SIZE)
if not cache then
  return error("failed to create the cache: " .. (err or "unknown"))
end

local function a_records_and_max_ttl(answers)
  local addresses = {}
  local ttl = MAXIMUM_TTL_VALUE -- maximum value according to https://tools.ietf.org/html/rfc2181

  for _, ans in ipairs(answers) do
    if ans.address then
      table.insert(addresses, ans.address)
      if ttl > ans.ttl then
        ttl = ans.ttl
      end
    end
  end

  return addresses, ttl
end

local function resolve_host(host, r, qtype)
  local answers
  answers, err = r:query(host, { qtype = qtype }, {})
  if not answers then
    return nil, -1, tostring(err)
  end

  if answers.errcode then
    return nil, -1, string.format("server returned error code: %s: %s", answers.errcode, answers.errstr)
  end

  local addresses, ttl = a_records_and_max_ttl(answers)
  if #addresses == 0 then
    return nil, -1, "no record resolved"
  end

  return addresses, ttl, nil
end

function _M.resolve(host)
  local cached_addresses = cache:get(host)
  if cached_addresses then
    local message = string.format(
      "addresses %s for host %s was resolved from cache",
      table.concat(cached_addresses, ", "), host)
    ngx.log(ngx.INFO, message)
    return cached_addresses
  end

  local r
  r, err = resolver:new{
    nameservers = util.deepcopy(configuration.nameservers),
    retrans = 5,
    timeout = 2000,  -- 2 sec
  }

  if not r then
    ngx.log(ngx.ERR, "failed to instantiate the resolver: " .. tostring(err))
    return { host }
  end

  local dns_errors = {}

  local addresses, ttl
  addresses, ttl, err = resolve_host(host, r, r.TYPE_A)
  if not addresses then
    table.insert(dns_errors, tostring(err))
  elseif #addresses > 0 then
    cache:set(host, addresses, ttl)
    return addresses
  end

  addresses, ttl, err = resolve_host(host, r, r.TYPE_AAAA)
  if not addresses then
    table.insert(dns_errors, tostring(err))
  elseif #addresses > 0 then
    cache:set(host, addresses, ttl)
    return addresses
  end

  if #dns_errors > 0 then
    ngx.log(ngx.ERR, "failed to query the DNS server:\n" .. table.concat(dns_errors, "\n"))
  end

  return { host }
end

return _M
