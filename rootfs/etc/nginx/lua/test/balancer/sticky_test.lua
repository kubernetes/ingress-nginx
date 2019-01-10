local sticky = require("balancer.sticky")
local cookie = require("resty.cookie")
local util = require("util")

local original_ngx = ngx

function mock_ngx(mock)
  local _ngx = mock
  setmetatable(_ngx, {__index = _G.ngx})
  _G.ngx = _ngx
end

local function reset_ngx()
  _G.ngx = original_ngx
end

function get_mocked_cookie_new()
  return function(self)
    return {
      get = function(self, n) return nil, "error" end,
      set = function(self, n) return true, "" end
    }
  end
end

cookie.new = get_mocked_cookie_new()

local function get_test_backend()
  return {
    name = "access-router-production-web-80",
    endpoints = {
      { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 },
    },
    sessionAffinityConfig = {
      name = "cookie",
      cookieSessionAffinity = { name = "test_name", hash = "sha1" }
    },
  }
end

describe("Sticky", function()
  before_each(function()
    mock_ngx({ var = { location_path = "/", host = "test.com" } })
  end)

  after_each(function()
    reset_ngx()
  end)

  local test_backend = get_test_backend()
  local test_backend_endpoint= test_backend.endpoints[1].address .. ":" .. test_backend.endpoints[1].port

  describe("new(backend)", function()
    context("when backend specifies cookie name", function()
      it("returns an instance containing the corresponding cookie name", function()
        local sticky_balancer_instance = sticky:new(test_backend)
        local test_backend_cookie_name = test_backend.sessionAffinityConfig.cookieSessionAffinity.name
        assert.equal(sticky_balancer_instance.cookie_name, test_backend_cookie_name)
      end)
    end)

    context("when backend does not specify cookie name", function()
      it("returns an instance with 'route' as cookie name", function()
        local temp_backend = util.deepcopy(test_backend)
        temp_backend.sessionAffinityConfig.cookieSessionAffinity.name = nil
        local sticky_balancer_instance = sticky:new(temp_backend)
        local default_cookie_name = "route"
        assert.equal(sticky_balancer_instance.cookie_name, default_cookie_name)
      end)
    end)

    context("when backend specifies hash function", function()
      it("returns an instance with the corresponding hash implementation", function()
        local sticky_balancer_instance = sticky:new(test_backend)
        local test_backend_hash_fn = test_backend.sessionAffinityConfig.cookieSessionAffinity.hash
        local test_backend_hash_implementation = util[test_backend_hash_fn .. "_digest"]
        assert.equal(sticky_balancer_instance.digest_func, test_backend_hash_implementation)
      end)
    end)

    context("when backend does not specify hash function", function()
      it("returns an instance with the default implementation (md5)", function()
        local temp_backend = util.deepcopy(test_backend)
        temp_backend.sessionAffinityConfig.cookieSessionAffinity.hash = nil
        local sticky_balancer_instance = sticky:new(temp_backend)
        local default_hash_fn = "md5"
        local default_hash_implementation = util[default_hash_fn .. "_digest"]
        assert.equal(sticky_balancer_instance.digest_func, default_hash_implementation)
      end)
    end)
  end)

  describe("balance()", function()
    local mocked_cookie_new = cookie.new

    before_each(function()
      package.loaded["balancer.sticky"] = nil
      sticky = require("balancer.sticky")
    end)

    after_each(function()
      cookie.new = mocked_cookie_new
    end)

    context("when client doesn't have a cookie set and location is in cookie_locations", function()
      it("picks an endpoint for the client", function()
        local sticky_balancer_instance = sticky:new(test_backend)
        local peer = sticky_balancer_instance:balance()
        assert.equal(peer, test_backend_endpoint)
      end)

      it("sets a cookie on the client", function()
        local s = {}
        cookie.new = function(self)
          local test_backend_hash_fn = test_backend.sessionAffinityConfig.cookieSessionAffinity.hash
          local cookie_instance = {
            set = function(self, payload)
              assert.equal(payload.key, test_backend.sessionAffinityConfig.cookieSessionAffinity.name)
              local expected_len = #util[test_backend_hash_fn .. "_digest"]("anything")
              assert.equal(#payload.value, expected_len)
              assert.equal(payload.path, ngx.var.location_path)
              assert.equal(payload.domain, nil)
              assert.equal(payload.httponly, true)
              assert.equal(payload.secure, false)
              return true, nil
            end,
            get = function(k) return false end,
          }
          s = spy.on(cookie_instance, "set")
          return cookie_instance, false
        end
        local b = get_test_backend()
        b.sessionAffinityConfig.cookieSessionAffinity.locations = {}
        b.sessionAffinityConfig.cookieSessionAffinity.locations["test.com"] = {"/"}
        local sticky_balancer_instance = sticky:new(b)
        assert.has_no.errors(function() sticky_balancer_instance:balance() end)
        assert.spy(s).was_called()
      end)

      it("sets a secure cookie on the client when being in ssl mode", function()
        ngx.var.https = "on"
        local s = {}
        cookie.new = function(self)
          local test_backend_hash_fn = test_backend.sessionAffinityConfig.cookieSessionAffinity.hash
          local cookie_instance = {
            set = function(self, payload)
              assert.equal(payload.key, test_backend.sessionAffinityConfig.cookieSessionAffinity.name)
              local expected_len = #util[test_backend_hash_fn .. "_digest"]("anything")
              assert.equal(#payload.value, expected_len)
              assert.equal(payload.path, ngx.var.location_path)
              assert.equal(payload.domain, nil)
              assert.equal(payload.httponly, true)
              assert.equal(payload.secure, true)
              return true, nil
            end,
            get = function(k) return false end,
          }
          s = spy.on(cookie_instance, "set")
          return cookie_instance, false
        end
        local b = get_test_backend()
        b.sessionAffinityConfig.cookieSessionAffinity.locations = {}
        b.sessionAffinityConfig.cookieSessionAffinity.locations["test.com"] = {"/"}
        local sticky_balancer_instance = sticky:new(b)
        assert.has_no.errors(function() sticky_balancer_instance:balance() end)
        assert.spy(s).was_called()
      end)
    end)

    context("when client doesn't have a cookie set and location not in cookie_locations", function()
      it("picks an endpoint for the client", function()
        local sticky_balancer_instance = sticky:new(test_backend)
        local peer = sticky_balancer_instance:balance()
        assert.equal(peer, test_backend_endpoint)
      end)

      it("does not set a cookie on the client", function()
        local s = {}
        cookie.new = function(self)
          local test_backend_hash_fn = test_backend.sessionAffinityConfig.cookieSessionAffinity.hash
          local cookie_instance = {
            set = function(self, payload)
              assert.equal(payload.key, test_backend.sessionAffinityConfig.cookieSessionAffinity.name)
              local expected_len = #util[test_backend_hash_fn .. "_digest"]("anything")
              assert.equal(#payload.value, expected_len)
              assert.equal(payload.path, ngx.var.location_path)
              assert.equal(payload.domain, ngx.var.host)
              assert.equal(payload.httponly, true)
              return true, nil
            end,
            get = function(k) return false end,
          }
          s = spy.on(cookie_instance, "set")
          return cookie_instance, false
        end
        local sticky_balancer_instance = sticky:new(get_test_backend())        
        assert.has_no.errors(function() sticky_balancer_instance:balance() end)
        assert.spy(s).was_not_called()
      end)
    end)

    context("when client has a cookie set", function()
      it("does not set a cookie", function()
        local s = {}
        cookie.new = function(self)
          local return_obj = {
            set = function(v) return false, nil end,
            get = function(k) return test_backend_endpoint end,
          }
          s = spy.on(return_obj, "set")
          return return_obj, false
        end
        local sticky_balancer_instance = sticky:new(test_backend)
        assert.has_no.errors(function() sticky_balancer_instance:balance() end)
        assert.spy(s).was_not_called()
      end)

      it("returns the correct endpoint for the client", function()
        local sticky_balancer_instance = sticky:new(test_backend)
        local peer = sticky_balancer_instance:balance()
        assert.equal(peer, test_backend_endpoint)
      end)
    end)
  end)
end)
