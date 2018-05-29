package.path = "./rootfs/etc/nginx/lua/?.lua;./rootfs/etc/nginx/lua/test/mocks/?.lua;" .. package.path

local util = require("util")

local _ngx = {
  shared = {
    balancer_ewma = { flush_all = function() end },
    balancer_ewma_last_touched_at = { flush_all = function() end }
  },
  log = function(...) end,
  now = function() return os.time() end,
}
_G.ngx = _ngx

describe("Balancer ewma", function()
  local balancer_ewma = require("balancer.ewma")

  describe("balance()", function()
    it("returns single endpoint when the given backend has only one endpoint", function()
      local backend = {
        name = "my-dummy-backend", ["load-balance"] = "ewma",
        endpoints = { { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 } }
      }
      local instance = balancer_ewma:new(backend)

      local host, port = instance:balance()
      assert.equal("10.184.7.40", host)
      assert.equal("8080", port)
    end)

    it("picks the endpoint with lowest score when there two of them", function()
      local backend = {
        name = "my-dummy-backend", ["load-balance"] = "ewma",
        endpoints = {
          { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 },
          { address = "10.184.97.100", port = "8080", maxFails = 0, failTimeout = 0 },
        }
      }
      local instance = balancer_ewma:new(backend)
      
      local stats = { ["10.184.7.40:8080"] = 0.5, ["10.184.97.100:8080"] = 0.3 }
      ngx.shared.balancer_ewma.get = function(self, key) return stats[key] end
      local t = ngx.now()-10
      ngx.shared.balancer_ewma_last_touched_at.get = function(self, key) return t end


      local host, port = instance:balance()
      assert.equal("10.184.97.100", host)
      assert.equal("8080", port)
    end)
  end)

  describe("sync()", function()
    local backend, instance

    before_each(function()
      backend = {
        name = "my-dummy-backend", ["load-balance"] = "ewma",
        endpoints = { { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 } }
      }
      instance = balancer_ewma:new(backend)
    end)

    it("does nothing when endpoints do not change", function()
      local new_backend = {
        endpoints = { { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 } }
      }

      local s1 = spy.on(ngx.shared.balancer_ewma, "flush_all")
      local s2 = spy.on(ngx.shared.balancer_ewma_last_touched_at, "flush_all")

      instance:sync(new_backend)

      assert.spy(s1).was_not_called()
      assert.spy(s2).was_not_called()
    end)

    it("updates endpoints", function()
      local new_backend = {
        endpoints = {
          { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 },
          { address = "10.184.97.100", port = "8080", maxFails = 0, failTimeout = 0 },
        }
      }

      instance:sync(new_backend)
      assert.are.same(new_backend.endpoints, instance.peers)
    end)

    it("resets stats", function()
      local new_backend = util.deepcopy(backend)
      new_backend.endpoints[1].maxFails = 3

      local s1 = spy.on(ngx.shared.balancer_ewma, "flush_all")
      local s2 = spy.on(ngx.shared.balancer_ewma_last_touched_at, "flush_all")

      instance:sync(new_backend)

      assert.spy(s1).was_called()
      assert.spy(s2).was_called()
    end)
  end)
end)
