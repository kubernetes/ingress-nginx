package.path = "./rootfs/etc/nginx/lua/?.lua;./rootfs/etc/nginx/lua/test/mocks/?.lua;" .. package.path

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
        name = "my-dummy-backend", ["upstream-hash-by"] = "$request_uri",
        endpoints = { { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 } }
      }
      local instance = balancer_chash:new(backend)

      local host, port = instance:balance()
      assert.equal("10.184.7.40", host)
      assert.equal("8080", port)
    end)
  end)
end)
