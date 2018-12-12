local certificate = require("certificate")

local PEM_CERT_KEY = "-----BEGIN CERTIFICATE-----\nMIID6DCCAlCgAwIBAgIQcfG0mA7BIFqhlnr/Zwh6TzANBgkqhkiG9w0BAQsFADBC\nMR4wHAYDVQQKExVta2NlcnQgZGV2ZWxvcG1lbnQgQ0ExIDAeBgNVBAsMF2hlbnJ5\ndHJhbkBPVFQtSGVucnlUcmFuMB4XDTE4MDcwOTIwMTY1MloXDTI4MDcwOTIwMTY1\nMlowSzEnMCUGA1UEChMebWtjZXJ0IGRldmVsb3BtZW50IGNlcnRpZmljYXRlMSAw\nHgYDVQQLDBdoZW5yeXRyYW5AT1RULUhlbnJ5VHJhbjCCASIwDQYJKoZIhvcNAQEB\nBQADggEPADCCAQoCggEBALIrsgHzjZZyKWPn3rGzTkaj9jADYAMhM+0wY3iky2Dx\ndr2YbKnZbbGxKLfVukYRsUUOK0SnBMTX15fsGanirj2hflMHfGvHilaVkVAkPJgD\nBTf2PkxFff99hS7/Ncz20MR6+E/vqp7Hx7IKDrg9lC9u1n82aotfN3gPhif8HyQu\n+P9cltsr9PewyPe4573WQmzXhTKaFm9+U9xZ2qS1J0DMEizRs45vuM040hxtiwVz\nM4Lm8DVpaYxMBWNI/zo9EZzoSJZH1sYUpTMwhNj+caEX+LK9PCM4Sht/yhPUc6aD\nnIEqraz+bS8dNFH5Ehp7n1SZL7YH6xT6da4F3ci7jEECAwEAAaNRME8wDgYDVR0P\nAQH/BAQDAgWgMBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwGgYD\nVR0RBBMwEYIPbXltaW5pa3ViZS5pbmZvMA0GCSqGSIb3DQEBCwUAA4IBgQC/TEpx\nJkL/ek37L2XKwGq96hNT10IZ9yE+RlndNGNY6eAc8y313sXBHTFbDCWQ2s0pWZZS\n+va20dnTQNyWzJAxFpdNvcKOCUGat4RPD/j+pBTEYk5n/oo7s2FWkG8kW6tKDilf\njWHpk7m9uYCO2sOZFiQPR81idR5PLox46SpJmIhDVfCi6VS4N+8fAT8Tbt9xkPmS\nmODhpnuIUt0NVTi62eqnxeO185qAt73xhz9Gj1KHntAK1ebcx0k3UxKRXQp9WY76\nF39sSz8OuhEhvv9ayl6uS6ZdSLvvb6kJrRRneKr01ridCOtiYB7cuXykDL1c6PUk\nugxDgTyCjiuPnRl0CLwxWT659PVozA2SO1YCW6UcoGdj2KMvsXezeWKpNGx3NHXO\nufdlxSbzWlamn+sPunWP3v2tfV0J8sHG3n1roeBO2N52197/ennGuCZfnF8C5MoG\n9YfMjKg9Z03G8sDpk9g5bHp9p28TO1X+Ht30PQzkUNhx3fjTO2DDvCyGk2k=\n-----END CERTIFICATE-----\n\n-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCyK7IB842Wcilj\n596xs05Go/YwA2ADITPtMGN4pMtg8Xa9mGyp2W2xsSi31bpGEbFFDitEpwTE19eX\n7Bmp4q49oX5TB3xrx4pWlZFQJDyYAwU39j5MRX3/fYUu/zXM9tDEevhP76qex8ey\nCg64PZQvbtZ/NmqLXzd4D4Yn/B8kLvj/XJbbK/T3sMj3uOe91kJs14UymhZvflPc\nWdqktSdAzBIs0bOOb7jNONIcbYsFczOC5vA1aWmMTAVjSP86PRGc6EiWR9bGFKUz\nMITY/nGhF/iyvTwjOEobf8oT1HOmg5yBKq2s/m0vHTRR+RIae59UmS+2B+sU+nWu\nBd3Iu4xBAgMBAAECggEBAIbuUIDp0fB9xJrEnwI0qLMWuPrjk3LLUmfunWZgZyWj\nuCkdpi17XHeVkyCl28v02itR77KuSg5I6B1F0Km34f0KsIBwyulU1I999e6bgsgc\ngXdAJS3d8u3qQVK2NChlQvWJq0PeXXiiE7nhpAQjnnXNmuP8cfPayEdEenUNmwfq\nxjEh2/oDzUTPD/4z5Hpw8n728SItgBolMNgGvmv5cC4JNLqujCUPwkLiZ3a2YbTY\nrmOO1xDkZnQWqyNP+baOwYwpu/kISPM3IveP5GGBNQsUsDxs6t80HNW/w8Ry0f50\n+gNTIuJVOLXfpVLIo87wTMEtRAqMzT4vxQIi+vj2XYECgYEA6QftSRKqur1G4P6Y\n9cseDnljJFWIjqex2q3NrvMaHbnlXp6AtPROoNz6L8H+PnBy8o1yZJwWnhKTvPaD\nsi+a1g7dqQIM4TjKLlidV57lt5ENw4ueW1o7Jbk75gawLhrBPSCFxR5xSqF+kQxn\nmWGjLnZoomD6fM2CG7EF1fg+wjsCgYEAw7t6db9tjGGM9J0rUe+jVtxMWiG9ArT0\nhmaLZQlKrOSFeEf8c4ZYBNxp/X+/jg0GWBX8P4KRubAz6bbn0A+07K9TClrvMFWq\nveqnK1JUsMGWsYQPp8dX8VS/jOFzYdiji9Ekyzs9RiXW8jzp0wzrccSjEr0de8HK\niEa9CH7cZ7MCgYB1V5mD51NzbyZG281YT+yVq0hiHnQCKa1kiYp+I0ouV9KJP9Vd\nyXvigwO0ksIc3PD09Ib65KJ6/K3KRHPygQg97ARwO2kS7E7a4aJxYcEZG4DLy/10\n0M3h5BGmdg23WZ+e0UarCPZRd1rNXWq5kLHkDpoH0j+wIqf2m8Bti3DGywKBgBEK\nn6zkz9rrG2So0n69yJDleVhXm6dCrg+NmhFf77qB4wUH73j3d25k6m2B0+HATI8a\nyu2Upq9uIfb1T9WTqIL6+NXr+OtSah1C8u8YqfsBv+cQwnQvLP78C/luH6ejPwoL\nWZLAQ6N54+8PUqRneZBcOH6HLKv7wXCACDFXKkV1AoGAfDb5GJ0NsovWhLfU5WEB\nSfdzHBplbp72q08S0aqTNm0wlTiCGYgm2Lle4IaOGoJ+7ipirL0KzuwisAZJFTvJ\nhsMqOmH/Ckledmf2JpLxyg8KB5KVA+RVQkrfVEv8yhqLcKQU6Z2n4jSTon7hXb1T\nf8neDpZ8DwO0W9cOdYLYyTg=\n-----END PRIVATE KEY-----"

local unmocked_ngx = _G.ngx

describe("Certificate", function()
  describe("call", function()
    local ssl = require("ngx.ssl")
    local match = require("luassert.match")

    before_each(function()
      ssl.server_name = function() return "hostname", nil end
      ssl.clear_certs = function() return true, "" end
      ssl.set_der_cert = function(cert) return true, "" end
      ssl.set_der_priv_key = function(priv_key) return true, "" end

      ngx.exit = function(status) end
    end)

    after_each(function()
      ngx = unmocked_ngx
      ngx.shared.certificate_data:flush_all()
    end)

    it("does not clear fallback certificates and logs error message when host is not in dictionary", function()
      ngx.shared.certificate_data:set("hostname", "")

      spy.on(ngx, "log")
      spy.on(ssl, "clear_certs")
      spy.on(ssl, "set_der_cert")
      spy.on(ssl, "set_der_priv_key")

      assert.has_no.errors(certificate.call)
      assert.spy(ngx.log).was_called_with(ngx.ERR, "Certificate not found, falling back on default certificate for hostname: hostname")
      assert.spy(ssl.clear_certs).was_not_called()
      assert.spy(ssl.set_der_cert).was_not_called()
      assert.spy(ssl.set_der_priv_key).was_not_called()
    end)

    it("does not clear fallback certificates and logs error message when the cert is empty for given host", function()
      spy.on(ngx, "log")
      spy.on(ssl, "clear_certs")
      spy.on(ssl, "set_der_cert")
      spy.on(ssl, "set_der_priv_key")

      assert.has_no.errors(certificate.call)
      assert.spy(ngx.log).was_called_with(ngx.ERR, "Certificate not found, falling back on default certificate for hostname: hostname")
      assert.spy(ssl.clear_certs).was_not_called()
      assert.spy(ssl.set_der_cert).was_not_called()
      assert.spy(ssl.set_der_priv_key).was_not_called()
    end)

    it("successfully sets SSL certificate and key when hostname is found in dictionary", function()
      ngx.shared.certificate_data:set("hostname", PEM_CERT_KEY)

      spy.on(ngx, "log")
      spy.on(ssl, "set_der_cert")
      spy.on(ssl, "set_der_priv_key")

      assert.has_no.errors(certificate.call)
      assert.spy(ngx.log).was_not_called_with(ngx.ERR, _)
      assert.spy(ssl.set_der_cert).was_called_with(ssl.cert_pem_to_der(PEM_CERT_KEY))
      assert.spy(ssl.set_der_priv_key).was_called_with(ssl.priv_key_pem_to_der(PEM_CERT_KEY))
    end)

    it("successfully sets SSL certificate and key for wildcard cert", function()
      ssl.server_name = function() return "sub.hostname", nil end
      ngx.shared.certificate_data:set("*.hostname", PEM_CERT_KEY)

      spy.on(ngx, "log")
      spy.on(ssl, "set_der_cert")
      spy.on(ssl, "set_der_priv_key")

      assert.has_no.errors(certificate.call)
      assert.spy(ngx.log).was_not_called_with(ngx.ERR, _)
      assert.spy(ssl.set_der_cert).was_called_with(ssl.cert_pem_to_der(PEM_CERT_KEY))
      assert.spy(ssl.set_der_priv_key).was_called_with(ssl.priv_key_pem_to_der(PEM_CERT_KEY))
    end)

    it("successfully sets SSL certificate and key for nested wildcard cert", function()
      ssl.server_name = function() return "sub.nested.hostname", nil end
      ngx.shared.certificate_data:set("*.nested.hostname", PEM_CERT_KEY)

      spy.on(ngx, "log")
      spy.on(ssl, "set_der_cert")
      spy.on(ssl, "set_der_priv_key")

      assert.has_no.errors(certificate.call)
      assert.spy(ngx.log).was_not_called_with(ngx.ERR, _)
      assert.spy(ssl.set_der_cert).was_called_with(ssl.cert_pem_to_der(PEM_CERT_KEY))
      assert.spy(ssl.set_der_priv_key).was_called_with(ssl.priv_key_pem_to_der(PEM_CERT_KEY))
    end)

    it("logs error message when certificate in dictionary is invalid", function()
      ngx.shared.certificate_data:set("hostname", "something invalid")

      spy.on(ngx, "log")
      spy.on(ssl, "set_der_cert")
      spy.on(ssl, "set_der_priv_key")

      assert.has_no.errors(certificate.call)
      assert.spy(ngx.log).was_called_with(ngx.ERR, "failed to convert certificate chain from PEM to DER: PEM_read_bio_X509_AUX() failed")
      assert.spy(ssl.set_der_cert).was_not_called()
      assert.spy(ssl.set_der_priv_key).was_not_called()
    end)

    it("does not clear fallback certificates and logs error message when hostname could not be fetched", function()
      ssl.server_name = function() return nil, "error" end

      spy.on(ngx, "log")
      spy.on(ssl, "clear_certs")
      spy.on(ssl, "set_der_cert")
      spy.on(ssl, "set_der_priv_key")

      assert.has_no.errors(certificate.call)
      assert.spy(ngx.log).was_called_with(ngx.ERR, "Error getting the hostname, falling back on default certificate: error")
      assert.spy(ssl.clear_certs).was_not_called()
      assert.spy(ssl.set_der_cert).was_not_called()
      assert.spy(ssl.set_der_priv_key).was_not_called()
    end)
  end)
end)
