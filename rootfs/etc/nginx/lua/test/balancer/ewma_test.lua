local util = require("util")

local original_ngx = ngx
local function reset_ngx()
  _G.ngx = original_ngx
end

local function mock_ngx(mock)
  local _ngx = mock
  setmetatable(_ngx, { __index = ngx })
  _G.ngx = _ngx
end

describe("Balancer ewma", function()
  local balancer_ewma = require("balancer.ewma")
  local ngx_now = 1543238266

  before_each(function()
    mock_ngx({ now = function() return ngx_now end })
  end)

  after_each(function()
    reset_ngx()
  end)

  describe("after_balance()", function()
    mock_ngx({ var = { upstream_response_time = "0.25", upstream_connect_time = "0.02", upstream_addr = "10.184.7.40:8080" } })

    it("updates EWMA stats", function()
      local backend = {
        name = "my-dummy-backend", ["load-balance"] = "ewma",
        endpoints = { { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 } }
      }
      local instance = balancer_ewma:new(backend)

      instance:after_balance()
      assert.equal(0.27, ngx.shared.balancer_ewma:get(ngx.var.upstream_addr))
      assert.equal(ngx_now, ngx.shared.balancer_ewma_last_touched_at:get(ngx.var.upstream_addr))
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
    local store_ewma_stats = function(endpoint_string, ewma, touched_at)
      ngx.shared.balancer_ewma:set(endpoint_string, ewma)
      ngx.shared.balancer_ewma_last_touched_at:set(endpoint_string, touched_at)
    end
    local assert_ewma_stats = function(endpoint_string, ewma, touched_at)
      assert.are.equals(ewma, ngx.shared.balancer_ewma:get(endpoint_string))
      assert.are.equals(touched_at, ngx.shared.balancer_ewma_last_touched_at:get(endpoint_string))
    end

    backend = {
      name = "namespace-service-port", ["load-balance"] = "ewma",
      endpoints = {
        { address = "10.10.10.1", port = "8080", maxFails = 0, failTimeout = 0 },
        { address = "10.10.10.2", port = "8080", maxFails = 0, failTimeout = 0 },
        { address = "10.10.10.3", port = "8080", maxFails = 0, failTimeout = 0 },
      }
    }
    store_ewma_stats("10.10.10.1:8080", 0.2, 1543238266)
    store_ewma_stats("10.10.10.2:8080", 0.3, 1543238269)
    store_ewma_stats("10.10.10.3:8080", 1.2, 1543238226)

    instance = balancer_ewma:new(backend)

    it("does not reset stats when endpoints do not change", function()
      local new_backend = util.deepcopy(backend)

      instance:sync(new_backend)

      assert.are.same(new_backend.endpoints, instance.peers)

      assert_ewma_stats("10.10.10.1:8080", 0.2, 1543238266)
      assert_ewma_stats("10.10.10.2:8080", 0.3, 1543238269)
      assert_ewma_stats("10.10.10.3:8080", 1.2, 1543238226)
    end)

    it("updates peers, deletes stats for old endpoints and sets average ewma score to new ones", function()
      local new_backend = util.deepcopy(backend)

      -- existing endpoint 10.10.10.2 got deleted
      -- and replaced with 11.10.10.2
      new_backend.endpoints[2].address = "10.10.10.4"
      -- and there's one new extra endpoint
      table.insert(new_backend.endpoints, { address = "10.10.10.5", port = "8080", maxFails = 0, failTimeout = 0 })

      instance:sync(new_backend)

      assert.are.same(new_backend.endpoints, instance.peers)

      assert_ewma_stats("10.10.10.1:8080", 0.2, 1543238266)
      assert_ewma_stats("10.10.10.2:8080", nil, nil)
      assert_ewma_stats("10.10.10.3:8080", 1.2, 1543238226)

      local slow_start_ewma = (0.2 + 1.2) / 2
      assert_ewma_stats("10.10.10.4:8080", slow_start_ewma, ngx_now)
      assert_ewma_stats("10.10.10.5:8080", slow_start_ewma, ngx_now)
    end)
  end)
end)
