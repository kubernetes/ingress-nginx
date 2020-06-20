function mock_ngx(mock)
  local _ngx = mock
  setmetatable(_ngx, {__index = _G.ngx})
  _G.ngx = _ngx
end

describe("Balancer chash", function()

  describe("balance()", function()
    it("uses correct key for given backend", function()
      mock_ngx({var = { request_uri = "/alma/armud"}})
      local balancer_chash = require("balancer.chash")

      local resty_chash = package.loaded["resty.chash"]
      resty_chash.new = function(self, nodes)
        return {
          find = function(self, key)
            assert.equal("/alma/armud", key)
            return "10.184.7.40:8080"
          end
        }
      end

      local backend = {
        name = "my-dummy-backend", upstreamHashByConfig = { ["upstream-hash-by"] = "$request_uri" },
        endpoints = { { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 } }
      }
      local instance = balancer_chash:new(backend)

      local peer = instance:balance()
      assert.equal("10.184.7.40:8080", peer)
    end)
  end)
end)
