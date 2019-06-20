describe("Balancer chash", function()
  local balancer_chash = require("balancer.chash")

  describe("balance()", function()
    it("uses correct key for given backend", function()
      _G.ngx = { var = { request_uri = "/alma/armud" }}

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
