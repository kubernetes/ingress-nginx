local ngx_re_split = require("ngx.re").split

local certificate_configured_for_current_request =
  require("certificate").configured_for_current_request
local global_throttle = require("global_throttle")

local ngx = ngx
local io = io
local math = math
local string = string
local original_randomseed = math.randomseed
local string_format = string.format
local ngx_redirect = ngx.redirect

local _M = {}

local seeds = {}
-- general Nginx configuration passed by controller to be used in this module
local config

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
    ngx.log(ngx.WARN, string.format("ignoring math.randomseed(%d) since PRNG "
      .. "is already seeded for worker %d", seed, pid))
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

local function redirect_to_https(location_config)
  if location_config.force_no_ssl_redirect then
    return false
  end

  if location_config.force_ssl_redirect and ngx.var.pass_access_scheme == "http" then
    return true
  end

  if ngx.var.pass_access_scheme ~= "http" then
    return false
  end

  return location_config.ssl_redirect and certificate_configured_for_current_request()
end

local function redirect_host()
  local host_port, err = ngx_re_split(ngx.var.best_http_host, ":")
  if err then
    ngx.log(ngx.ERR, "could not parse variable: ", err)
    return ngx.var.best_http_host;
  end

  return host_port[1];
end

local function parse_x_forwarded_host()
  local hosts, err = ngx_re_split(ngx.var.http_x_forwarded_host, ",")
  if err then
    ngx.log(ngx.ERR, string_format("could not parse variable: %s", err))
    return ""
  end

  return hosts[1]
end

function _M.init_worker()
  randomseed()
end

function _M.set_config(new_config)
  config = new_config
end

-- rewrite gets called in every location context.
-- This is where we do variable assignments to be used in subsequent
-- phases or redirection
function _M.rewrite(location_config)
  ngx.var.pass_access_scheme = ngx.var.scheme

  ngx.var.best_http_host = ngx.var.http_host or ngx.var.host

  if config.use_forwarded_headers then
    -- trust http_x_forwarded_proto headers correctly indicate ssl offloading
    if ngx.var.http_x_forwarded_proto then
      ngx.var.pass_access_scheme = ngx.var.http_x_forwarded_proto
    end

    if ngx.var.http_x_forwarded_port then
      ngx.var.pass_server_port = ngx.var.http_x_forwarded_port
    end

    -- Obtain best http host
    if ngx.var.http_x_forwarded_host then
      ngx.var.best_http_host = parse_x_forwarded_host()
    end
  end

  if config.use_proxy_protocol then
    if ngx.var.proxy_protocol_server_port == "443" then
      ngx.var.pass_access_scheme = "https"
    end
  end

  ngx.var.pass_port = ngx.var.pass_server_port
  if config.is_ssl_passthrough_enabled then
    if ngx.var.pass_server_port == config.listen_ports.ssl_proxy then
      ngx.var.pass_port = 443
    end
  elseif ngx.var.pass_server_port == config.listen_ports.https then
    ngx.var.pass_port = 443
  end

  if redirect_to_https(location_config) then
    local request_uri = ngx.var.request_uri
    -- do not append a trailing slash on redirects unless enabled by annotations
    if location_config.preserve_trailing_slash == false then
      if string.byte(request_uri, -1, -1) == string.byte('/') then
        request_uri = string.sub(request_uri, 1, -2)
      end
    end

    local uri = string_format("https://%s%s", redirect_host(), request_uri)

    if location_config.use_port_in_redirects then
      uri = string_format("https://%s:%s%s", redirect_host(),
        config.listen_ports.https, request_uri)
    end

    return ngx_redirect(uri, config.http_redirect_code)
  end

  global_throttle.throttle(config.global_throttle, location_config.global_throttle)
end

function _M.header()
  if config.hsts and ngx.var.scheme == "https" and certificate_configured_for_current_request then
    local value = "max-age=" .. config.hsts_max_age
    if config.hsts_include_subdomains then
      value = value .. "; includeSubDomains"
    end
    if config.hsts_preload then
      value = value .. "; preload"
    end
    ngx.header["Strict-Transport-Security"] = value
  end
end

return _M
