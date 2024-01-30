local certificate = require("certificate")
local ssl = require("ngx.ssl")

local function read_file(path)
  local file = assert(io.open(path, "rb"))
  local content = file:read("*a")
  file:close()
  return content
end

local EXAMPLE_CERT = read_file("rootfs/etc/nginx/lua/test/fixtures/example-com-cert.pem")
local DEFAULT_CERT = read_file("rootfs/etc/nginx/lua/test/fixtures/default-cert.pem")
local DEFAULT_CERT_HOSTNAME = "_"
local UUID = "2ea8adb5-8ebb-4b14-a79b-0cdcd892e884"
local DEFAULT_UUID = "00000000-0000-0000-0000-000000000000"

local function assert_certificate_is_set(cert)
  spy.on(ngx, "log")
  spy.on(ssl, "set_cert")
  spy.on(ssl, "set_priv_key")

  assert.has_no.errors(certificate.call)
  assert.spy(ngx.log).was_not_called_with(ngx.ERR, _)
  assert.spy(ssl.set_cert).was_called_with("cert")
  assert.spy(ssl.set_priv_key).was_called_with("priv_key")
end

local function refute_certificate_is_set()
  spy.on(ssl, "set_cert")
  spy.on(ssl, "set_priv_key")

  assert.has_no.errors(certificate.call)
  assert.spy(ssl.set_cert).was_not_called()
  assert.spy(ssl.set_priv_key).was_not_called()
end

local function set_certificate(hostname, certificate, uuid)
  local success, err = ngx.shared.certificate_servers:set(hostname, uuid)
  if not success then
    error(err)
  end
  success, err = ngx.shared.certificate_data:set(uuid, certificate)
  if not success then
    error(err)
  end
end

local unmocked_ngx = _G.ngx

describe("Certificate", function()
  describe("call", function()
    before_each(function()
      ssl.server_name = function() return "hostname", nil end
      ssl.clear_certs = function() return true, "" end
      ssl.parse_pem_cert = function(cert)
        if cert == "invalid" then
          return nil, "bad format"
        else
          return "cert", nil
        end
      end
      ssl.parse_pem_priv_key = function(priv_key)
        if priv_key == "invalid" then
          return nil, "bad format"
        else
          return "priv_key", nil
        end
      end
      ssl.cert_pem_to_der = function(cert)
        if cert == "invalid" then
          return nil, "bad format"
        else
          return "der_cert", nil
        end
      end
      ssl.set_cert = function(cert) return true, "" end
      ssl.set_priv_key = function(priv_key) return true, "" end

      ngx.exit = function(status) end

      certificate.set_cache_size(1000)
      set_certificate(DEFAULT_CERT_HOSTNAME, DEFAULT_CERT, DEFAULT_UUID)
    end)

    after_each(function()
      ngx = unmocked_ngx
      ngx.shared.certificate_data:flush_all()
      ngx.shared.certificate_servers:flush_all()
      certificate.flush_cache()
    end)

    it("sets certificate and key when hostname is found in dictionary", function()
      set_certificate("hostname", EXAMPLE_CERT, UUID)
      assert_certificate_is_set(EXAMPLE_CERT)
    end)

    it("parses certificate and key once when cache hits", function()
      spy.on(ssl, "parse_pem_cert")

      set_certificate("hostname", EXAMPLE_CERT, UUID)
      assert_certificate_is_set(EXAMPLE_CERT)

      assert_certificate_is_set(EXAMPLE_CERT)

      assert.spy(ssl.parse_pem_cert).was.called(1)
    end)

    it("parses certificate and key again when cache hits but cert content changes", function()
      spy.on(ssl, "parse_pem_cert")

      set_certificate("hostname", EXAMPLE_CERT, UUID)
      assert_certificate_is_set(EXAMPLE_CERT)

      set_certificate("hostname", DEFAULT_CERT, UUID)
      assert_certificate_is_set(DEFAULT_CERT)

      assert.spy(ssl.parse_pem_cert).was.called(2)
    end)

    it("sets certificate and key for wildcard cert", function()
      ssl.server_name = function() return "sub.hostname", nil end
      set_certificate("*.hostname", EXAMPLE_CERT, UUID)

      assert_certificate_is_set(EXAMPLE_CERT)
    end)

    it("sets certificate and key for domain with trailing dot", function()
      ssl.server_name = function() return "hostname.", nil end
      set_certificate("hostname", EXAMPLE_CERT, UUID)

      assert_certificate_is_set(EXAMPLE_CERT)
    end)

    it("fallbacks to default certificate and key for domain with many trailing dots", function()
      ssl.server_name = function() return "hostname..", nil end
      set_certificate("hostname", EXAMPLE_CERT, UUID)

      assert_certificate_is_set(DEFAULT_CERT)
    end)

    it("sets certificate and key for nested wildcard cert", function()
      ssl.server_name = function() return "sub.nested.hostname", nil end
      set_certificate("*.nested.hostname", EXAMPLE_CERT, UUID)

      assert_certificate_is_set(EXAMPLE_CERT)
    end)

    it("logs error message when certificate in dictionary is invalid", function()
      set_certificate("hostname", "invalid", UUID)

      spy.on(ngx, "log")

      refute_certificate_is_set()
      assert.spy(ngx.log).was_called_with(ngx.ERR, "failed to convert certificate chain from PEM to DER: bad format")
    end)

    it("uses default certificate when there's none found for given hostname", function()
      assert_certificate_is_set(DEFAULT_CERT)
    end)

    it("uses default certificate when hostname can not be obtained", function()
      ssl.server_name = function() return nil, "crazy hostname error" end

      assert_certificate_is_set(DEFAULT_CERT)
      assert.spy(ngx.log).was_called_with(ngx.ERR, "error while obtaining hostname: crazy hostname error")
    end)

    it("fails when hostname does not have certificate and default cert is invalid", function()
      set_certificate(DEFAULT_CERT_HOSTNAME, "invalid", UUID)

      spy.on(ngx, "log")

      refute_certificate_is_set()
      assert.spy(ngx.log).was_called_with(ngx.ERR, "failed to convert certificate chain from PEM to DER: bad format")
    end)

    describe("OCSP stapling", function()
      before_each(function()
        certificate.is_ocsp_stapling_enabled = true
      end)

      after_each(function()
        certificate.is_ocsp_stapling_enabled = false
      end)

      it("fetches and caches OCSP response when there is no cached response", function()
      end)

      it("fetches and caches OCSP response when cached response is stale", function()
      end)

      it("staples using cached OCSP response", function()
      end)

      it("staples using cached stale OCSP response", function()
      end)

      it("does negative caching when OCSP response URL extraction fails", function()
      end)

      it("does negative caching when the request to OCSP responder fails", function()
      end)
    end)
  end)

  describe("configured_for_current_request", function()
    before_each(function()
      local _ngx = { var = { host = "hostname" } }
      setmetatable(_ngx, {__index = _G.ngx})
      _G.ngx = _ngx
      ngx.ctx.cert_configured_for_current_request = nil

      package.loaded["certificate"] = nil
      certificate = require("certificate")

      set_certificate("hostname", EXAMPLE_CERT, UUID)
    end)

    it("returns true when certificate exists for given server", function()
      assert.is_true(certificate.configured_for_current_request())
    end)

    it("returns false when certificate does not exist for given server", function()
      ngx.var.host = "hostname.xyz"
      assert.is_false(certificate.configured_for_current_request())
    end)

    it("returns cached value from ngx.ctx", function()
      ngx.ctx.cert_configured_for_current_request = false
      assert.is_false(certificate.configured_for_current_request())
    end)
  end)
end)
