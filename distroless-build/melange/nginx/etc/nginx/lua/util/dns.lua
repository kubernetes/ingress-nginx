local resolver = require("resty.dns.resolver")
local lrucache = require("resty.lrucache")
local resolv_conf = require("util.resolv_conf")

local ngx_log = ngx.log
local ngx_INFO = ngx.INFO
local ngx_ERR = ngx.ERR
local string_format = string.format
local table_concat = table.concat
local table_insert = table.insert
local ipairs = ipairs
local tostring = tostring

local _M = {}
local CACHE_SIZE = 10000
-- maximum value according to https://tools.ietf.org/html/rfc2181
local MAXIMUM_TTL_VALUE = 2147483647
-- for every host we will try two queries for the following types with the order set here
local QTYPES_TO_CHECK = { resolver.TYPE_A, resolver.TYPE_AAAA }

local cache
do
  local err
  cache, err = lrucache.new(CACHE_SIZE)
  if not cache then
    return error("failed to create the cache: " .. (err or "unknown"))
  end
end

local function cache_set(host, addresses, ttl)
  cache:set(host, addresses, ttl)
  ngx_log(ngx_INFO, string_format("cache set for '%s' with value of [%s] and ttl of %s.",
    host, table_concat(addresses, ", "), ttl))
end

local function is_fully_qualified(host)
  return host:sub(-1) == "."
end

local function a_records_and_min_ttl(answers)
  local addresses = {}
  local ttl = MAXIMUM_TTL_VALUE -- maximum value according to https://tools.ietf.org/html/rfc2181

  for _, ans in ipairs(answers) do
    if ans.address then
      table_insert(addresses, ans.address)
      if ans.ttl < ttl then
        ttl = ans.ttl
      end
    end
  end

  return addresses, ttl
end

local function resolve_host_for_qtype(r, host, qtype)
  local answers, err = r:query(host, { qtype = qtype }, {})
  if not answers then
    return nil, -1, err
  end

  if answers.errcode then
    return nil, -1, string_format("server returned error code: %s: %s",
      answers.errcode, answers.errstr)
  end

  local addresses, ttl = a_records_and_min_ttl(answers)
  if #addresses == 0 then
    local msg = "no A record resolved"
    if qtype == resolver.TYPE_AAAA then msg = "no AAAA record resolved" end
    return nil, -1, msg
  end

  return addresses, ttl, nil
end

local function resolve_host(r, host)
  local dns_errors = {}

  for _, qtype in ipairs(QTYPES_TO_CHECK) do
    local addresses, ttl, err = resolve_host_for_qtype(r, host, qtype)
    if addresses and #addresses > 0 then
      return addresses, ttl, nil
    end
    table_insert(dns_errors, tostring(err))
  end

  return nil, nil, dns_errors
end

function _M.lookup(host)
  local cached_addresses = cache:get(host)
  if cached_addresses then
    return cached_addresses
  end

  local r, err = resolver:new{
    nameservers = resolv_conf.nameservers,
    retrans = 5,
    timeout = 2000,  -- 2 sec
  }

  if not r then
    ngx_log(ngx_ERR, string_format("failed to instantiate the resolver: %s", err))
    return { host }
  end

  local addresses, ttl, dns_errors

  -- when the queried domain is fully qualified
  -- then we don't go through resolv_conf.search
  -- NOTE(elvinefendi): currently FQDN as externalName will be supported starting
  -- with K8s 1.15: https://github.com/kubernetes/kubernetes/pull/78385
  if is_fully_qualified(host) then
    addresses, ttl, dns_errors = resolve_host(r, host)
    if addresses then
      cache_set(host, addresses, ttl)
      return addresses
    end

    ngx_log(ngx_ERR, "failed to query the DNS server for ",
      host, ":\n", table_concat(dns_errors, "\n"))

    return { host }
  end

  -- for non fully qualified domains if number of dots in
  -- the queried host is less than resolv_conf.ndots then we try
  -- with all the entries in resolv_conf.search before trying the original host
  --
  -- if number of dots is not less than resolv_conf.ndots then we start with
  -- the original host and then try entries in resolv_conf.search
  local _, host_ndots = host:gsub("%.", "")
  local search_start, search_end = 0, #resolv_conf.search
  if host_ndots < resolv_conf.ndots then
    search_start = 1
    search_end = #resolv_conf.search + 1
  end

  for i = search_start, search_end, 1 do
    local new_host = resolv_conf.search[i] and
      string_format("%s.%s", host, resolv_conf.search[i]) or host

    addresses, ttl, dns_errors = resolve_host(r, new_host)
    if addresses then
      cache_set(host, addresses, ttl)
      return addresses
    end
  end

  if #dns_errors > 0 then
    ngx_log(ngx_ERR, "failed to query the DNS server for ",
      host, ":\n", table_concat(dns_errors, "\n"))
  end

  return { host }
end

setmetatable(_M, {__index = { _cache = cache }})

return _M
