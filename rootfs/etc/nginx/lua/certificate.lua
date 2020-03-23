local ssl = require("ngx.ssl")
local re_sub = ngx.re.sub

local _M = {}

local DEFAULT_CERT_HOSTNAME = "_"

local certificate_data = ngx.shared.certificate_data
local certificate_servers = ngx.shared.certificate_servers

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
  local hostname = re_sub(raw_hostname, "\\.$", "", "jo")

  local uid = certificate_servers:get(hostname)
  if uid then
    return uid
  end

  local wildcard_hosatname, _, err = re_sub(hostname, "^[^\\.]+\\.", "*.", "jo")
  if err then
    ngx.log(ngx.ERR, "error: ", err)
    return uid
  end

  if wildcard_hosatname then
    uid = ngx.shared.certificate_servers:get(wildcard_hosatname)
  end

  return uid
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
    ngx.log(ngx.INFO,
      "obtained hostname is nil (the client does not support SNI?), falling back to default certificate")
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
    ngx.log(ngx.ERR, "certificate not found, falling back to fake certificate for hostname: " .. tostring(hostname))
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

  -- TODO: based on `der_cert` find OCSP responder URL
  -- make OCSP request and POST it there and get the response and staple it to
  -- the current SSL connection if OCSP stapling is enabled
end

return _M
