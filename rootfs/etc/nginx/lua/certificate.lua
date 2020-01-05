local ssl = require("ngx.ssl")
local configuration = require("configuration")
local re_sub = ngx.re.sub
local http = require "resty.http"
local ocsp = require "ngx.ocsp"

local certificate_servers = ngx.shared.certificate_servers

local _M = {}

local DEFAULT_CERT_HOSTNAME = "_"

local function set_pem_cert_key(pem_cert_key)
  local der_cert, der_cert_err = ssl.cert_pem_to_der(pem_cert_key)
  if not der_cert then
    return "failed to convert certificate chain from PEM to DER: " .. der_cert_err
  end

  local set_cert_ok, set_cert_err = ssl.set_der_cert(der_cert)
  if not set_cert_ok then
    return "failed to set DER cert: " .. set_cert_err
  end

  local der_priv_key, dev_priv_key_err = ssl.priv_key_pem_to_der(pem_cert_key)
  if not der_priv_key then
    return "failed to convert private key from PEM to DER: " .. dev_priv_key_err
  end

  local set_priv_key_ok, set_priv_key_err = ssl.set_der_priv_key(der_priv_key)
  if not set_priv_key_ok then
    return "failed to set DER private key: " .. set_priv_key_err
  end
end

local function get_pem_cert_key(raw_hostname)
  local hostname = re_sub(raw_hostname, "\\.$", "", "jo")

  local pem_cert_key = configuration.get_pem_cert_key(hostname)
  if pem_cert_key then
    return pem_cert_key
  end

  local wildcard_hosatname, _, err = re_sub(hostname, "^[^\\.]+\\.", "*.", "jo")
  if err then
    ngx.log(ngx.ERR, "error: ", err)
    return pem_cert_key
  end

  if wildcard_hosatname then
    pem_cert_key = configuration.get_pem_cert_key(wildcard_hosatname)
  end
  return pem_cert_key
end

local function get_ocsp_response(pem_cert_key)
  local der_cert_chain, der_cert_err = ssl.cert_pem_to_der(pem_cert_key)
  if not der_cert_chain then
    return nil, "failed to convert certificate chain from PEM to DER: " .. der_cert_err
  end

  local ocsp_url, ocsp_responder_err = ocsp.get_ocsp_responder_from_der_chain(der_cert_chain)
  if not ocsp_url then
    return nil, "failed to get OCSP URL: " .. (ocsp_responder_err or "")
  end

  local ocsp_req, ocsp_request_err = ocsp.create_ocsp_request(der_cert_chain)
  if not ocsp_req then
    return nil, "failed to create OCSP request: " .. (ocsp_request_err or "")
  end

  local httpc = http.new()
  httpc:set_timeout(10000)

  local res, req_err = httpc:request_uri(ocsp_url, {
    method = "POST",
    body = ocsp_req,
    headers = {
      ["Content-Type"] = "application/ocsp-request",
    }
  })

  -- Perform various checks to ensure we have a valid OCSP response.
  if not res then
    return nil, "OCSP responder query failed (" .. (ocsp_url or "") .. "): " .. (req_err or "")
  end

  if res.status ~= 200 then
    return nil, "OCSP responder returns bad HTTP status code (" .. (ocsp_url or "") .. "): " .. (res.status or "")
  end

  httpc:set_keepalive()

  local ocsp_resp = res.body
  if not ocsp_resp or ocsp_resp == "" then
    return nil, "OCSP responder returns bad response body (" .. (ocsp_url or "") .. "): " .. (ocsp_resp or "")
  end

  local ok, ocsp_validate_err = ocsp.validate_ocsp_response(ocsp_resp, der_cert_chain)
  if not ok then
    return nil, "failed to validate OCSP response (" .. (ocsp_url or "") .. "): " .. (ocsp_validate_err or "")
  end

  return ocsp_resp, nil
end

function _M.configured_for_current_request()
  if ngx.ctx.configured_for_current_request ~= nil then
    return ngx.ctx.configured_for_current_request
  end

  ngx.ctx.configured_for_current_request = get_pem_cert_key(ngx.var.host) ~= nil

  return ngx.ctx.configured_for_current_request
end

function _M.call()
  local hostname, hostname_err = ssl.server_name()
  if hostname_err then
    ngx.log(ngx.ERR, "error while obtaining hostname: " .. hostname_err)
  end
  if not hostname then
    ngx.log(ngx.INFO,
      "obtained hostname is nil (the client does not support SNI?), falling back to default certificate")
    hostname = DEFAULT_CERT_HOSTNAME
  end

  local pem_cert_key = get_pem_cert_key(hostname)
  if not pem_cert_key then
    pem_cert_key = get_pem_cert_key(DEFAULT_CERT_HOSTNAME)
  end
  if not pem_cert_key then
    ngx.log(ngx.ERR, "certificate not found, falling back to fake certificate for hostname: " .. tostring(hostname))
    return
  end

  local clear_ok, clear_err = ssl.clear_certs()
  if not clear_ok then
    ngx.log(ngx.ERR, "failed to clear existing (fallback) certificates: " .. clear_err)
    return ngx.exit(ngx.ERROR)
  end

  local set_pem_cert_key_err = set_pem_cert_key(pem_cert_key)
  if set_pem_cert_key_err then
    ngx.log(ngx.ERR, set_pem_cert_key_err)
    return ngx.exit(ngx.ERROR)
  end

  --- fake SSL certificate should not return OCSP responses
  if hostname == DEFAULT_CERT_HOSTNAME then
    return
  end

  --- only server names with a valid certificate can return OCSP responses
  local cert_exists = certificate_servers:get(hostname)
  if not cert_exists then
    return
  end

  --- check if the OCSP response is in the cache
  local ocsp_resp = ngx.shared.ocsp_cache:get(hostname)
  if not ocsp_resp then
    local ocsp_response_err
    ocsp_resp, ocsp_response_err = get_ocsp_response(pem_cert_key)
    if not ocsp_resp then
      ngx.log(ngx.ERR, "failed to get OCSP response for hostname " .. hostname .. " : " .. ocsp_response_err)
      return
    end

    local _, set_ocsp_err, set_ocsp_forcible = ngx.shared.ocsp_cache:set(hostname, ocsp_resp, 3600)
    if set_ocsp_err then
      ngx.log(ngx.ERR, "failed to set cache of OCSP response for " .. hostname .. ": ", set_ocsp_err)
    elseif set_ocsp_forcible then
      ngx.log(ngx.ERR, "'lua_shared_dict ocsp_cache' might be too small - consider increasing the size")
    end
  end

  local ok, ocsp_status_err = ocsp.set_ocsp_status_resp(ocsp_resp)
  if not ok then
    ngx.log(ngx.ERR, "failed to set OCSP information: " .. ocsp_status_err)
  end

end

return _M
