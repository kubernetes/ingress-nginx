local util = require("util")

describe("Balancer ewma", function()
  local balancer_ewma = require("balancer.ewma")

  describe("after_balance()", function()
    local ngx_now = 1543238266
    _G.ngx.now = function() return ngx_now end
    _G.ngx.var = { upstream_response_time = "0.25", upstream_connect_time = "0.02", upstream_addr = "10.184.7.40:8080" }

    it("updates EWMA stats", function()
      local backend = {
        name = "my-dummy-backend", ["load-balance"] = "ewma",
        endpoints = { { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 } }
      }
      local instance = balancer_ewma:new(backend)

      instance:after_balance()
      assert.equal(0.27, instance.ewma[ngx.var.upstream_addr])
      assert.equal(ngx_now, instance.ewma_last_touched_at[ngx.var.upstream_addr])
    end)
  end)

  describe("balance()", function()
    it("returns single endpoint when the given backend has only one endpoint", function()
      local backend = {
        name = "my-dummy-backend", ["load-balance"] = "ewma",
        endpoints = { { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 } }
      }
      local instance = balancer_ewma:new(backend)

      local peer = instance:balance()
      assert.equal("10.184.7.40:8080", peer)
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
      instance.ewma =  { ["10.184.7.40:8080"] = 0.5, ["10.184.97.100:8080"] = 0.3 }
      instance.ewma_last_touched_at =  { ["10.184.7.40:8080"] = ngx.now(), ["10.184.97.100:8080"] = ngx.now() }

      local peer = instance:balance()
      assert.equal("10.184.97.100:8080", peer)
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

      instance:sync(new_backend)
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

      instance:sync(new_backend)
    end)
  end)
end)
