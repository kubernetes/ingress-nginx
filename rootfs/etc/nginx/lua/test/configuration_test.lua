local cjson = require("cjson")
local configuration = require("configuration")

local unmocked_ngx = _G.ngx
local certificate_data = ngx.shared.certificate_data
local certificate_servers = ngx.shared.certificate_servers
local ocsp_response_cache = ngx.shared.ocsp_response_cache

local function get_backends()
  return {
    {
      name = "my-dummy-backend-1", ["load-balance"] = "sticky",
      endpoints = { { address = "10.183.7.40", port = "8080", maxFails = 0, failTimeout = 0 } },
      sessionAffinityConfig = { name = "cookie", cookieSessionAffinity = { name = "route" } },
    },
    {
      name = "my-dummy-backend-2", ["load-balance"] = "ewma",
      endpoints = {
        { address = "10.184.7.40", port = "7070", maxFails = 3, failTimeout = 2 },
        { address = "10.184.7.41", port = "7070", maxFails = 2, failTimeout = 1 },
      }
    },
    {
      name = "my-dummy-backend-3", ["load-balance"] = "round_robin",
      endpoints = {
        { address = "10.185.7.40", port = "6060", maxFails = 0, failTimeout = 0 },
        { address = "10.185.7.41", port = "6060", maxFails = 2, failTimeout = 1 },
      }
    },
  }
end

local function get_mocked_ngx_env()
  local _ngx = {
    status = ngx.HTTP_OK,
    var = {},
    req = {
      read_body = function() end,
      get_body_data = function() return cjson.encode(get_backends()) end,
      get_body_file = function() return nil end,
    },
    log = function(msg) end,
  }
  setmetatable(_ngx, {__index = _G.ngx})
  return _ngx
end

describe("Configuration", function()
  before_each(function()
    _G.ngx = get_mocked_ngx_env()
    package.loaded["configuration"] = nil
    configuration = require("configuration")
  end)

  after_each(function()
    _G.ngx = unmocked_ngx
  end)

  describe("Backends", function()
    context("Request method is neither GET nor POST", function()
      it("sends 'Only POST and GET requests are allowed!' in the response body", function()
        ngx.var.request_method = "PUT"
        local s = spy.on(ngx, "print")
        assert.has_no.errors(configuration.call)
        assert.spy(s).was_called_with("Only POST and GET requests are allowed!")
      end)

      it("returns a status code of 400", function()
        ngx.var.request_method = "PUT"
        assert.has_no.errors(configuration.call)
        assert.equal(ngx.status, ngx.HTTP_BAD_REQUEST)
      end)
    end)

    context("GET request to /configuration/backends", function()
      before_each(function()
        ngx.var.request_method = "GET"
        ngx.var.request_uri = "/configuration/backends"
      end)

      it("returns the current configured backends on the response body", function()
        -- Encoding backends since comparing tables fail due to reference comparison
        local encoded_backends = cjson.encode(get_backends())
        ngx.shared.configuration_data:set("backends", encoded_backends)
        local s = spy.on(ngx, "print")
        assert.has_no.errors(configuration.call)
        assert.spy(s).was_called_with(encoded_backends)
      end)

      it("returns a status of 200", function()
        assert.has_no.errors(configuration.call)
        assert.equal(ngx.status, ngx.HTTP_OK)
      end)
    end)

    context("POST request to /configuration/backends", function()
      before_each(function()
        ngx.var.request_method = "POST"
        ngx.var.request_uri = "/configuration/backends"
      end)

      it("stores the posted backends on the shared dictionary", function()
        -- Encoding backends since comparing tables fail due to reference comparison
        assert.has_no.errors(configuration.call)
        assert.equal(ngx.shared.configuration_data:get("backends"), cjson.encode(get_backends()))
      end)

      context("Failed to read request body", function()
        local mocked_get_body_data = ngx.req.get_body_data
        before_each(function()
          ngx.req.get_body_data = function() return nil end
        end)

        teardown(function()
          ngx.req.get_body_data = mocked_get_body_data
        end)

        it("returns a status of 400", function()
          local original_io_open = _G.io.open
          _G.io.open = function(filename, extension) return false end
          assert.has_no.errors(configuration.call)
          assert.equal(ngx.status, ngx.HTTP_BAD_REQUEST)
          _G.io.open = original_io_open
        end)

        it("logs 'dynamic-configuration: unable to read valid request body to stderr'", function()
          local original_io_open = _G.io.open
          _G.io.open = function(filename, extension) return false end
          local s = spy.on(ngx, "log")
          assert.has_no.errors(configuration.call)
          assert.spy(s).was_called_with(ngx.ERR, "dynamic-configuration: unable to read valid request body")
          _G.io.open = original_io_open
        end)
      end)

      context("Failed to set the new backends to the configuration dictionary", function()
        local resty_configuration_data_set = ngx.shared.configuration_data.set
        before_each(function()
          ngx.shared.configuration_data.set = function(key, value) return false, "" end
        end)

        teardown(function()
          ngx.shared.configuration_data.set = resty_configuration_data_set
        end)

        it("returns a status of 400", function()
          assert.has_no.errors(configuration.call)
          assert.equal(ngx.status, ngx.HTTP_BAD_REQUEST)
        end)

        it("logs 'dynamic-configuration: error updating configuration:' to stderr", function()
          local s = spy.on(ngx, "log")
          assert.has_no.errors(configuration.call)
          assert.spy(s).was_called_with(ngx.ERR, "dynamic-configuration: error updating configuration: ")
        end)
      end)

      context("Succeeded to update backends configuration", function()
        it("returns a status of 201", function()
          assert.has_no.errors(configuration.call)
          assert.equal(ngx.status, ngx.HTTP_CREATED)
        end)
      end)
    end)
  end)

  describe("handle_servers()", function()
    local UUID = "2ea8adb5-8ebb-4b14-a79b-0cdcd892e884"

    local function mock_ssl_configuration(configuration)
      local json = cjson.encode(configuration)
      ngx.req.get_body_data = function() return json end
    end

    before_each(function()
      ngx.var.request_method = "POST"
    end)

    it("should not accept non POST methods", function()
      ngx.var.request_method = "GET"

      local s = spy.on(ngx, "print")
      assert.has_no.errors(configuration.handle_servers)
      assert.spy(s).was_called_with("Only POST requests are allowed!")
      assert.same(ngx.status, ngx.HTTP_BAD_REQUEST)
    end)

    it("should not delete ocsp_response_cache if certificate remain the same", function()
      ngx.shared.certificate_data.get = function(self, uid)
        return "pemCertKey"
      end

      mock_ssl_configuration({
        servers = { ["hostname"] = UUID },
        certificates = { [UUID] = "pemCertKey" }
      })

      local s = spy.on(ngx.shared.ocsp_response_cache, "delete")
      assert.has_no.errors(configuration.handle_servers)
      assert.spy(s).was_not_called()
    end)

    it("should not delete ocsp_response_cache if certificate is empty", function()
      ngx.shared.certificate_data.get = function(self, uid)
          return nil
      end

      mock_ssl_configuration({
        servers = { ["hostname"] = UUID },
        certificates = { [UUID] = "pemCertKey" }
      })

      local s = spy.on(ngx.shared.ocsp_response_cache, "delete")
      assert.has_no.errors(configuration.handle_servers)
      assert.spy(s).was_not_called()
    end)

    it("should delete ocsp_response_cache if certificate changed", function()
      local stored_entries = {
          [UUID] = "pemCertKey"
      }

      ngx.shared.certificate_data.get = function(self, uid)
          return stored_entries[uid]
      end

      mock_ssl_configuration({
        servers = { ["hostname"] = UUID },
        certificates = { [UUID] = "pemCertKey2" }
      })

      local s = spy.on(ngx.shared.ocsp_response_cache, "delete")

      assert.has_no.errors(configuration.handle_servers)
      assert.spy(s).was.called_with(ocsp_response_cache, UUID)
    end)

    it("deletes server with empty UID without touching the corresponding certificate", function()
      mock_ssl_configuration({
        servers = { ["hostname"] = UUID },
        certificates = { [UUID] = "pemCertKey" }
      })
      assert.has_no.errors(configuration.handle_servers)
      assert.same("pemCertKey", certificate_data:get(UUID))
      assert.same(UUID, certificate_servers:get("hostname"))
      assert.same(ngx.HTTP_CREATED, ngx.status)

      local EMPTY_UID = "-1"
      mock_ssl_configuration({
        servers = { ["hostname"] = EMPTY_UID },
        certificates = { [UUID] = "pemCertKey" }
      })
      assert.has_no.errors(configuration.handle_servers)
      assert.same("pemCertKey", certificate_data:get(UUID))
      assert.same(nil, certificate_servers:get("hostname"))
      assert.same(ngx.HTTP_CREATED, ngx.status)
    end)

    it("should successfully update certificates and keys for each host", function()
      mock_ssl_configuration({
        servers = { ["hostname"] = UUID },
        certificates = { [UUID] = "pemCertKey" }
      })

      assert.has_no.errors(configuration.handle_servers)
      assert.same("pemCertKey", certificate_data:get(UUID))
      assert.same(UUID, certificate_servers:get("hostname"))
      assert.same(ngx.HTTP_CREATED, ngx.status)
    end)

    it("should log an err and set status to Internal Server Error when a certificate cannot be set", function()
      local uuid2 = "8ea8adb5-8ebb-4b14-a79b-0cdcd892e999"
      ngx.shared.certificate_data.set = function(self, uuid, certificate)
        return false, "error", nil
      end

      mock_ssl_configuration({
        servers = { ["hostname"] = UUID, ["hostname2"] = uuid2 },
        certificates = { [UUID] = "pemCertKey", [uuid2] = "pemCertKey2" }
      })

      local s = spy.on(ngx, "log")
      assert.has_no.errors(configuration.handle_servers)
      assert.same(ngx.HTTP_INTERNAL_SERVER_ERROR, ngx.status)
    end)

    it("logs a warning when entry is forcibly stored", function()
      local uuid2 = "8ea8adb5-8ebb-4b14-a79b-0cdcd892e999"
      local stored_entries = {}

      ngx.shared.certificate_data.set = function(self, uuid, certificate)
        stored_entries[uuid] = certificate
        return true, nil, true
      end
      mock_ssl_configuration({
        servers = { ["hostname"] = UUID, ["hostname2"] = uuid2 },
        certificates = { [UUID] = "pemCertKey", [uuid2] = "pemCertKey2" }
      })

      local s1 = spy.on(ngx, "log")
      assert.has_no.errors(configuration.handle_servers)
      assert.spy(s1).was_called_with(ngx.WARN, string.format("certificate_data dictionary is full, LRU entry has been removed to store %s", UUID))
      assert.equal("pemCertKey", stored_entries[UUID])
      assert.equal("pemCertKey2", stored_entries[uuid2])
      assert.same(ngx.HTTP_CREATED, ngx.status)
    end)
  end)
end)
