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

local function assert_certificate_is_set(cert)
  spy.on(ngx, "log")
  spy.on(ssl, "set_der_cert")
  spy.on(ssl, "set_der_priv_key")

  assert.has_no.errors(certificate.call)
  assert.spy(ngx.log).was_not_called_with(ngx.ERR, _)
  assert.spy(ssl.set_der_cert).was_called_with(ssl.cert_pem_to_der(cert))
  assert.spy(ssl.set_der_priv_key).was_called_with(ssl.priv_key_pem_to_der(cert))
end

local function refute_certificate_is_set()
  spy.on(ssl, "set_der_cert")
  spy.on(ssl, "set_der_priv_key")

  assert.has_no.errors(certificate.call)
  assert.spy(ssl.set_der_cert).was_not_called()
  assert.spy(ssl.set_der_priv_key).was_not_called()
end

local unmocked_ngx = _G.ngx

describe("Certificate", function()
  describe("call", function()
    before_each(function()
      ssl.server_name = function() return "hostname", nil end
      ssl.clear_certs = function() return true, "" end
      ssl.set_der_cert = function(cert) return true, "" end
      ssl.set_der_priv_key = function(priv_key) return true, "" end

      ngx.exit = function(status) end


      ngx.shared.certificate_data:set(DEFAULT_CERT_HOSTNAME, DEFAULT_CERT)
    end)

    after_each(function()
      ngx = unmocked_ngx
      ngx.shared.certificate_data:flush_all()
    end)

    it("sets certificate and key when hostname is found in dictionary", function()
      ngx.shared.certificate_data:set("hostname", EXAMPLE_CERT)

      assert_certificate_is_set(EXAMPLE_CERT)
    end)

    it("sets certificate and key for wildcard cert", function()
      ssl.server_name = function() return "sub.hostname", nil end
      ngx.shared.certificate_data:set("*.hostname", EXAMPLE_CERT)

      assert_certificate_is_set(EXAMPLE_CERT)
    end)

    it("sets certificate and key for domain with trailing dot", function()
      ssl.server_name = function() return "hostname.", nil end
      ngx.shared.certificate_data:set("hostname", EXAMPLE_CERT)

      assert_certificate_is_set(EXAMPLE_CERT)
    end)

    it("fallbacks to default certificate and key for domain with many trailing dots", function()
      ssl.server_name = function() return "hostname..", nil end
      ngx.shared.certificate_data:set("hostname", EXAMPLE_CERT)

      assert_certificate_is_set(DEFAULT_CERT)
    end)

    it("sets certificate and key for nested wildcard cert", function()
      ssl.server_name = function() return "sub.nested.hostname", nil end
      ngx.shared.certificate_data:set("*.nested.hostname", EXAMPLE_CERT)

      assert_certificate_is_set(EXAMPLE_CERT)
    end)

    it("logs error message when certificate in dictionary is invalid", function()
      ngx.shared.certificate_data:set("hostname", "something invalid")

      spy.on(ngx, "log")

      refute_certificate_is_set()
      assert.spy(ngx.log).was_called_with(ngx.ERR, "failed to convert certificate chain from PEM to DER: PEM_read_bio_X509_AUX() failed")
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
      ngx.shared.certificate_data:set(DEFAULT_CERT_HOSTNAME, "invalid")

      spy.on(ngx, "log")

      refute_certificate_is_set()
      assert.spy(ngx.log).was_called_with(ngx.ERR, "failed to convert certificate chain from PEM to DER: PEM_read_bio_X509_AUX() failed")
    end)
  end)
end)
