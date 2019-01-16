_G._TEST = true

local balancer, expected_implementations, backends

local function reset_balancer()
  package.loaded["balancer"] = nil
  balancer = require("balancer")
end

local function reset_expected_implementations()
  expected_implementations = {
    ["access-router-production-web-80"] = package.loaded["balancer.round_robin"],
    ["my-dummy-app-1"] = package.loaded["balancer.round_robin"],
    ["my-dummy-app-2"] = package.loaded["balancer.chash"],
    ["my-dummy-app-3"] = package.loaded["balancer.sticky"],
    ["my-dummy-app-4"] = package.loaded["balancer.ewma"],
    ["my-dummy-app-5"] = package.loaded["balancer.sticky"],
  }
end

local function reset_backends()
  backends = {
    {
      name = "access-router-production-web-80", port = "80", secure = false,
      secureCACert = { secret = "", caFilename = "", pemSha = "" },
      sslPassthrough = false,
      endpoints = {
        { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 },
        { address = "10.184.97.100", port = "8080", maxFails = 0, failTimeout = 0 },
        { address = "10.184.98.239", port = "8080", maxFails = 0, failTimeout = 0 },
      },
      sessionAffinityConfig = { name = "", cookieSessionAffinity = { name = "", hash = "" } },
    },
    { name = "my-dummy-app-1", ["load-balance"] = "round_robin", },
    { 
      name = "my-dummy-app-2", ["load-balance"] = "chash",
      upstreamHashByConfig = { ["upstream-hash-by"] = "$request_uri", },
    },
    {
      name = "my-dummy-app-3", ["load-balance"] = "ewma",
      sessionAffinityConfig = { name = "cookie", cookieSessionAffinity = { name = "route", hash = "sha1" } }
    },
    { name = "my-dummy-app-4", ["load-balance"] = "ewma", },
    {
      name = "my-dummy-app-5", ["load-balance"] = "ewma", ["upstream-hash-by"] = "$request_uri",
      sessionAffinityConfig = { name = "cookie", cookieSessionAffinity = { name = "route", hash = "sha1" } }
    },
  }
end

describe("Balancer", function()
  before_each(function()
    reset_balancer()
    reset_expected_implementations()
    reset_backends()
  end)

  describe("get_implementation()", function()
    it("returns correct implementation for given backend", function()
      for _, backend in pairs(backends) do
        local expected_implementation = expected_implementations[backend.name]
        local implementation = balancer.get_implementation(backend)
        assert.equal(expected_implementation, balancer.get_implementation(backend))
      end
    end)
  end)

  describe("sync_backend()", function()
    local backend, implementation

    before_each(function()
      backend = backends[1]
      implementation = expected_implementations[backend.name]
    end)

    it("initializes balancer for given backend", function()
      local s = spy.on(implementation, "new")

      assert.has_no.errors(function() balancer.sync_backend(backend) end)
      assert.spy(s).was_called_with(implementation, backend)
    end)

    it("resolves external name to endpoints when service is of type External name", function()
      backend = {
        name = "exmaple-com", service = { spec = { ["type"] = "ExternalName" } },
        endpoints = {
          { address = "example.com", port = "80", maxFails = 0, failTimeout = 0 }
        }
      }

      local dns_helper = require("test/dns_helper")
      dns_helper.mock_dns_query({
        {
          name = "example.com",
          address = "192.168.1.1",
          ttl = 3600,
        },
        {
          name = "example.com",
          address = "1.2.3.4",
          ttl = 60,
        }
      })
      expected_backend = {
        name = "exmaple-com", service = { spec = { ["type"] = "ExternalName" } },
        endpoints = {
          { address = "192.168.1.1", port = "80" },
          { address = "1.2.3.4", port = "80" },
        }
      }

      local mock_instance = { sync = function(backend) end }
      setmetatable(mock_instance, implementation)
      implementation.new = function(self, backend) return mock_instance end
      assert.has_no.errors(function() balancer.sync_backend(backend) end)
      stub(mock_instance, "sync")
      assert.has_no.errors(function() balancer.sync_backend(backend) end)
      assert.stub(mock_instance.sync).was_called_with(mock_instance, expected_backend)
    end)

    it("wraps IPv6 addresses into square brackets", function()
      local backend = {
        name = "exmaple-com",
        endpoints = {
          { address = "::1", port = "8080", maxFails = 0, failTimeout = 0 },
          { address = "192.168.1.1", port = "8080", maxFails = 0, failTimeout = 0 },
        }
      }
      local expected_backend = {
        name = "exmaple-com",
        endpoints = {
          { address = "[::1]", port = "8080", maxFails = 0, failTimeout = 0 },
          { address = "192.168.1.1", port = "8080", maxFails = 0, failTimeout = 0 },
        }
      }

      local mock_instance = { sync = function(backend) end }
      setmetatable(mock_instance, implementation)
      implementation.new = function(self, backend) return mock_instance end
      assert.has_no.errors(function() balancer.sync_backend(backend) end)
      stub(mock_instance, "sync")
      assert.has_no.errors(function() balancer.sync_backend(backend) end)
      assert.stub(mock_instance.sync).was_called_with(mock_instance, expected_backend)
    end)

    it("replaces the existing balancer when load balancing config changes for backend", function()
      assert.has_no.errors(function() balancer.sync_backend(backend) end)

      backend["load-balance"] = "ewma"
      local new_implementation = package.loaded["balancer.ewma"]

      local s_old = spy.on(implementation, "new")
      local s = spy.on(new_implementation, "new")
      local s_ngx_log = spy.on(ngx, "log")

      assert.has_no.errors(function() balancer.sync_backend(backend) end)
      assert.spy(s_ngx_log).was_called_with(ngx.INFO,
      "LB algorithm changed from round_robin to ewma, resetting the instance")
      -- TODO(elvinefendi) figure out why
      -- assert.spy(s).was_called_with(new_implementation, backend) does not work here
      assert.spy(s).was_called(1)
      assert.spy(s_old).was_not_called()
    end)

    it("calls sync(backend) on existing balancer instance when load balancing config does not change", function()
      local mock_instance = { sync = function(...) end }
      setmetatable(mock_instance, implementation)
      implementation.new = function(self, backend) return mock_instance end
      assert.has_no.errors(function() balancer.sync_backend(backend) end)

      stub(mock_instance, "sync")

      assert.has_no.errors(function() balancer.sync_backend(backend) end)
      assert.stub(mock_instance.sync).was_called_with(mock_instance, backend)
    end)
  end)
end)
