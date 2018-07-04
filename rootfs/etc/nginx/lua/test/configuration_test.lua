package.path = "./rootfs/etc/nginx/lua/?.lua;./rootfs/etc/nginx/lua/test/mocks/?.lua;" .. package.path
_G._TEST = true

local _ngx = {
  shared = {
      configuration_data = {
        get = function(self, rsc) return {backend = true} end,
        set = function(self, a) end,
      }
  },
  print = function(msg) return msg end,
  log = function() end,
  var = {},
  req = {
      read_body = function() end,
      get_body_data = function() return {body_data = true} end,
  }
}
_G.ngx = _ngx

local configuration = require("configuration")

describe("Configuration", function()
    context("Request method is neither GET nor POST", function()
        it("should log 'Only POST and GET requests are allowed!'", function()
            ngx.var.request_method = "PUT"
            local s = spy.on(ngx, "print")
            assert.has_no.errors(function() configuration.call() end)
            assert.spy(s).was_called_with("Only POST and GET requests are allowed!")
        end)
    end)

    context("GET request to /configuration/backends", function()
        before_each(function()
            ngx.var.request_method = "GET"
            ngx.var.request_uri = "/configuration/backends"
        end)

        it("should call get_backends_data()", function()
            local s = spy.on(configuration, "get_backends_data")
            assert.has_no.errors(function() configuration.call() end)
            assert.spy(s).was_called()
        end)

        it("should call configuration_data:get('backends')", function()
            local s = spy.on(ngx.shared.configuration_data, "get")
            assert.has_no.errors(function() configuration.call() end)
            assert.spy(s).was_called_with(ngx.shared.configuration_data, "backends")
            assert.spy(s).returned_with({backend = true})
        end)
    end)

    context("POST request to /configuration/backends", function()
        it("should call configuration_data:set('backends')", function()
            ngx.var.request_method = "POST"
            ngx.var.request_uri = "/configuration/backends"

            local s = spy.on(ngx.shared.configuration_data, "set")
            assert.has_no.errors(function() configuration.call() end)
            assert.spy(s).was_called_with(ngx.shared.configuration_data, "backends", {body_data = true})
        end)
    end)


end)

