local http = require("resty.http")
local ssl = require("ngx.ssl")
local ocsp = require("ngx.ocsp")
local ngx = ngx
local string = string
local tostring = tostring
local re_sub = ngx.re.sub
local unpack = unpack

local dns_lookup = require("util.dns").lookup

local _M = {
  is_ocsp_stapling_enabled = false
}

local DEFAULT_CERT_HOSTNAME = "_"

local certificate_data = ngx.shared.certificate_data
local certificate_servers = ngx.shared.certificate_servers
local ocsp_response_cache = ngx.shared.ocsp_response_cache

local function get_der_cert_and_priv_key(pem_cert_key)
  local der_cert, der_cert_err = ssl.cert_pem_to_der(pem_cert_key)
  if not der_cert then
    return nil, nil, "failed to convert certificate chain from PEM to DER: " .. der_cert_err
  end

  local der_priv_key, dev_priv_key_err = ssl.priv_key_pem_to_der(pem_cert_key)
  if not der_priv_key then
    return nil, nil, "failed to convert private key from PEM to DER: " .. dev_priv_key_err
  end

  return der_cert, der_priv_key, nil
end

local function set_der_cert_and_key(der_cert, der_priv_key)
  local set_cert_ok, set_cert_err = ssl.set_der_cert(der_cert)
  if not set_cert_ok then
    return "failed to set DER cert: " .. set_cert_err
  end

  local set_priv_key_ok, set_priv_key_err = ssl.set_der_priv_key(der_priv_key)
  if not set_priv_key_ok then
    return "failed to set DER private key: " .. set_priv_key_err
  end
end

local function get_pem_cert_uid(raw_hostname)
  -- Convert hostname to ASCII lowercase (see RFC 6125 6.4.1) so that requests with uppercase
  -- host would lead to the right certificate being chosen (controller serves certificates for
  -- lowercase hostnames as specified in Ingress object's spec.rules.host)
  local hostname = re_sub(raw_hostname, "\\.$", "", "jo"):gsub("[A-Z]",
    function(c) return c:lower() end)

  local uid = certificate_servers:get(hostname)
  if uid then
    return uid
  end

  local wildcard_hostname, _, err = re_sub(hostname, "^[^\\.]+\\.", "*.", "jo")
  if err then
    ngx.log(ngx.ERR, "error: ", err)
    return uid
  end

  if wildcard_hostname then
    uid = certificate_servers:get(wildcard_hostname)
  end

  return uid
end

local function is_ocsp_stapling_enabled_for(_)
  -- TODO: implement per ingress OCSP stapling control
  -- and make use of uid. The idea is to have configureCertificates
  -- in controller side to push uid -> is_ocsp_enabled data to Lua land.

  return _M.is_ocsp_stapling_enabled
end

local function get_resolved_url(parsed_url)
  local scheme, host, port, path = unpack(parsed_url)
  local ip = dns_lookup(host)[1]
  return string.format("%s://%s:%s%s", scheme, ip, port, path)
end

local function do_ocsp_request(url, ocsp_request)
  local httpc = http.new()
  httpc:set_timeout(1000, 1000, 2000)

  local parsed_url, err = httpc:parse_uri(url)
  if not parsed_url then
    return nil, err
  end

  local resolved_url = get_resolved_url(parsed_url)

  local http_response
  http_response, err = httpc:request_uri(resolved_url, {
    method = "POST",
    headers = {
      ["Content-Type"] = "application/ocsp-request",
      ["Host"] = parsed_url[2],
    },
    body = ocsp_request,
  })
  if not http_response then
    return nil, err
  end
  if http_response.status ~= 200 then
    return nil, "unexpected OCSP responder status code: " .. tostring(http_response.status)
  end

  return http_response.body, nil
end

-- TODO: ideally this function should have a lock around to ensure
-- only one instance runs at a time. Otherwise it is theoretically possible
-- that this function gets called from multiple Nginx workers at the same time.
-- While this has no functional implications, it generates extra load on OCSP servers.
local function fetch_and_cache_ocsp_response(uid, der_cert)
  local url, err = ocsp.get_ocsp_responder_from_der_chain(der_cert)
  if not url then
    ngx.log(ngx.ERR, "could not extract OCSP responder URL: ", err)
    return
  end

  local request
  request, err = ocsp.create_ocsp_request(der_cert)
  if not request then
    ngx.log(ngx.ERR, "could not create OCSP request: ", err)
    return
  end

  local ocsp_response
  ocsp_response, err = do_ocsp_request(url, request)
  if err then
    ngx.log(ngx.ERR, "could not get OCSP response: ", err)
    return
  end
  if not ocsp_response or #ocsp_response == 0 then
    ngx.log(ngx.ERR, "OCSP responder returned an empty response")
    return
  end

  local ok
  ok, err = ocsp.validate_ocsp_response(ocsp_response, der_cert)
  if not ok then
    -- We are doing the same thing as vanilla Nginx here - if response status is not "good"
    -- we do not use it - no stapling.
    -- We can look into differentiation of validation errors and when status is i.e "revoked"
    -- we might want to continue with stapling - it is at the least counterintuitive that
    -- one would not staple response when certificate is revoked (I have not managed to find
    -- and spec about this). Also one would expect browsers to do all these verifications
    -- comprehensively, so why we bother doing this on server side? This can be tricky though:
    -- imagine the certificate is not revoked but its OCSP responder is having some issues
    -- and not generating a valid OCSP response. We would then staple that invalid OCSP response
    -- and then browser would fail the connection because of invalid OCSP response - as a result
    -- user request fails. But as a server we can validate response here and not staple it
    -- to the connection if it is invalid. But if browser/client has must-staple enabled
    -- then this will break anyway. So for must-staple there's no difference from users'
    -- perspective. When must-staple is not enabled though it is better to not staple
    -- invalid response and let the client/browser to fallback to CRL check or retry OCSP
    -- on its own.
    --

    -- Also we should do negative caching here to avoid sending too many request to
    -- the OCSP responder. Imagine OCSP responder is having an intermittent issue
    -- and we keep sending request. It might make things worse for the responder.

    ngx.log(ngx.NOTICE, "OCSP response validation failed: ", err)
    return
  end

  -- Normally this should be (nextUpdate - thisUpdate), but Lua API does not expose
  -- those attributes.
  local expiry = 3600 * 24 * 3
  local success, forcible
  success, err, forcible = ocsp_response_cache:set(uid, ocsp_response, expiry)
  if not success then
    ngx.log(ngx.ERR, "failed to cache OCSP response: ", err)
  end
  if forcible then
    ngx.log(ngx.NOTICE, "removed an existing item when saving OCSP response, ",
      "consider increasing shared dictionary size for 'ocsp_reponse_cache'")
  end
end

-- ocsp_staple looks at the cache and staples response from cache if it exists
-- if there is no cached response or the existing response is stale,
-- it enqueues fetch_and_cache_ocsp_response function to refetch the response.
-- This design tradeoffs lack of OCSP response in the first request with better latency.
--
-- Serving stale response ensures that we don't serve another request without OCSP response
-- when the cache entry expires. Instead we serve the signle request with stale response
-- and enqueue fetch_and_cache_ocsp_response for refetch.
local function ocsp_staple(uid, der_cert)
  local response, _, is_stale = ocsp_response_cache:get_stale(uid)
  if not response or is_stale then
    ngx.timer.at(0, function() fetch_and_cache_ocsp_response(uid, der_cert) end)
    return false, nil
  end

  local ok, err = ocsp.set_ocsp_status_resp(response)
  if not ok then
    return false, err
  end

  return true, nil
end

function _M.configured_for_current_request()
  if ngx.ctx.cert_configured_for_current_request == nil then
    ngx.ctx.cert_configured_for_current_request = get_pem_cert_uid(ngx.var.host) ~= nil
  end

  return ngx.ctx.cert_configured_for_current_request
end

function _M.call()
  local hostname, hostname_err = ssl.server_name()
  if hostname_err then
    ngx.log(ngx.ERR, "error while obtaining hostname: " .. hostname_err)
  end
  if not hostname then
    ngx.log(ngx.INFO, "obtained hostname is nil (the client does "
      .. "not support SNI?), falling back to default certificate")
    hostname = DEFAULT_CERT_HOSTNAME
  end

  local pem_cert
  local pem_cert_uid = get_pem_cert_uid(hostname)
  if not pem_cert_uid then
    pem_cert_uid = get_pem_cert_uid(DEFAULT_CERT_HOSTNAME)
  end
  if pem_cert_uid then
    pem_cert = certificate_data:get(pem_cert_uid)
  end
  if not pem_cert then
    ngx.log(ngx.ERR, "certificate not found, falling back to fake certificate for hostname: "
      .. tostring(hostname))
    return
  end

  local clear_ok, clear_err = ssl.clear_certs()
  if not clear_ok then
    ngx.log(ngx.ERR, "failed to clear existing (fallback) certificates: " .. clear_err)
    return ngx.exit(ngx.ERROR)
  end

  local der_cert, der_priv_key, der_err = get_der_cert_and_priv_key(pem_cert)
  if der_err then
    ngx.log(ngx.ERR, der_err)
    return ngx.exit(ngx.ERROR)
  end

  local set_der_err = set_der_cert_and_key(der_cert, der_priv_key)
  if set_der_err then
    ngx.log(ngx.ERR, set_der_err)
    return ngx.exit(ngx.ERROR)
  end

  if is_ocsp_stapling_enabled_for(pem_cert_uid) then
    local _, err = ocsp_staple(pem_cert_uid, der_cert)
    if err then
      ngx.log(ngx.ERR, "error during OCSP stapling: ", err)
    end
  end
end

return _M
