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

local function flush_all_ewma_stats()
  ngx.shared.balancer_ewma:flush_all()
  ngx.shared.balancer_ewma_last_touched_at:flush_all()
end

local function store_ewma_stats(endpoint_string, ewma, touched_at)
  ngx.shared.balancer_ewma:set(endpoint_string, ewma)
  ngx.shared.balancer_ewma_last_touched_at:set(endpoint_string, touched_at)
end

local function assert_ewma_stats(endpoint_string, ewma, touched_at)
  assert.are.equals(ewma, ngx.shared.balancer_ewma:get(endpoint_string))
  assert.are.equals(touched_at, ngx.shared.balancer_ewma_last_touched_at:get(endpoint_string))
end


describe("Balancer ewma", function()
  local balancer_ewma = require("balancer.ewma")
  local ngx_now = 1543238266
  local backend, instance

  before_each(function()
    mock_ngx({ now = function() return ngx_now end, var = { balancer_ewma_score = -1 } })
    package.loaded["balancer.ewma"] = nil
    balancer_ewma = require("balancer.ewma")

    backend = {
      name = "namespace-service-port", ["load-balance"] = "ewma",
      endpoints = {
        { address = "10.10.10.1", port = "8080", maxFails = 0, failTimeout = 0 },
        { address = "10.10.10.2", port = "8080", maxFails = 0, failTimeout = 0 },
        { address = "10.10.10.3", port = "8080", maxFails = 0, failTimeout = 0 },
      }
    }
    store_ewma_stats("10.10.10.1:8080", 0.2, ngx_now - 1)
    store_ewma_stats("10.10.10.2:8080", 0.3, ngx_now - 5)
    store_ewma_stats("10.10.10.3:8080", 1.2, ngx_now - 20)

    instance = balancer_ewma:new(backend)
  end)

  after_each(function()
    reset_ngx()
    flush_all_ewma_stats()
  end)

  describe("after_balance()", function()
    it("updates EWMA stats", function()
      ngx.var = { upstream_addr = "10.10.10.2:8080", upstream_connect_time = "0.02", upstream_response_time = "0.1" }

      instance:after_balance()

      local weight = math.exp(-5 / 10)
      local expected_ewma = 0.3 * weight + 0.12 * (1.0 - weight)

      assert.are.equals(expected_ewma, ngx.shared.balancer_ewma:get(ngx.var.upstream_addr))
      assert.are.equals(ngx_now, ngx.shared.balancer_ewma_last_touched_at:get(ngx.var.upstream_addr))
    end)

    it("updates EWMA stats with the latest result", function()
      ngx.var = { upstream_addr = "10.10.10.1:8080, 10.10.10.2:8080", upstream_connect_time = "0.05, 0.02", upstream_response_time = "0.2, 0.1" }

      instance:after_balance()

      local weight = math.exp(-5 / 10)
      local expected_ewma = 0.3 * weight + 0.12 * (1.0 - weight)

      assert.are.equals(expected_ewma, ngx.shared.balancer_ewma:get("10.10.10.2:8080"))
      assert.are.equals(ngx_now, ngx.shared.balancer_ewma_last_touched_at:get("10.10.10.2:8080"))
    end)
  end)

  describe("balance()", function()
    it("returns single endpoint when the given backend has only one endpoint", function()
      local single_endpoint_backend = util.deepcopy(backend)
      table.remove(single_endpoint_backend.endpoints, 3)
      table.remove(single_endpoint_backend.endpoints, 2)
      local single_endpoint_instance = balancer_ewma:new(single_endpoint_backend)

      local peer = single_endpoint_instance:balance()

      assert.are.equals("10.10.10.1:8080", peer)
      assert.are.equals(-1, ngx.var.balancer_ewma_score)
    end)

    it("picks the endpoint with lowest decayed score", function()
      local two_endpoints_backend = util.deepcopy(backend)
      table.remove(two_endpoints_backend.endpoints, 2)
      local two_endpoints_instance = balancer_ewma:new(two_endpoints_backend)

      local peer = two_endpoints_instance:balance()

      -- even though 10.10.10.1:8080 has a lower ewma score
      -- algorithm picks 10.10.10.3:8080 because its decayed score is even lower
      assert.equal("10.10.10.3:8080", peer)
      assert.equal(true, ngx.ctx.balancer_ewma_tried_endpoints["10.10.10.3:8080"])
      assert.are.equals(0.16240233988393523723, ngx.var.balancer_ewma_score)
    end)

    it("doesn't pick the tried endpoint while retry", function()
      local two_endpoints_backend = util.deepcopy(backend)
      table.remove(two_endpoints_backend.endpoints, 2)
      local two_endpoints_instance = balancer_ewma:new(two_endpoints_backend)

      ngx.ctx.balancer_ewma_tried_endpoints = {
        ["10.10.10.3:8080"] = true,
      }
      local peer = two_endpoints_instance:balance()
      assert.equal("10.10.10.1:8080", peer)
      assert.equal(true, ngx.ctx.balancer_ewma_tried_endpoints["10.10.10.1:8080"])
    end)

    it("all the endpoints are tried, pick the one with lowest score", function()
      local two_endpoints_backend = util.deepcopy(backend)
      table.remove(two_endpoints_backend.endpoints, 2)
      local two_endpoints_instance = balancer_ewma:new(two_endpoints_backend)

      ngx.ctx.balancer_ewma_tried_endpoints = {
        ["10.10.10.1:8080"] = true,
        ["10.10.10.3:8080"] = true,
      }
      local peer = two_endpoints_instance:balance()
      assert.equal("10.10.10.3:8080", peer)
    end)
  end)

  describe("sync()", function()
    it("does not reset stats when endpoints do not change", function()
      local new_backend = util.deepcopy(backend)

      instance:sync(new_backend)

      assert.are.same(new_backend.endpoints, instance.peers)

      assert_ewma_stats("10.10.10.1:8080", 0.2, ngx_now - 1)
      assert_ewma_stats("10.10.10.2:8080", 0.3, ngx_now - 5)
      assert_ewma_stats("10.10.10.3:8080", 1.2, ngx_now - 20)
    end)

    it("resets alternative backends and traffic shaping policy even if endpoints do not change", function()
      assert.are.same(nil, instance.alternativeBackends)
      assert.are.same(nil, instance.trafficShapingPolicy)

      local new_backend = util.deepcopy(backend)
      new_backend.alternativeBackends = {"my-canary-namespace-my-canary-service-my-port"}
      new_backend.trafficShapingPolicy = {
        cookie = "",
        header = "",
        headerPattern = "",
        headerValue = "",
        weight = 20,
      }

      instance:sync(new_backend)

      assert.are.same(new_backend.alternativeBackends, instance.alternative_backends)
      assert.are.same(new_backend.trafficShapingPolicy, instance.traffic_shaping_policy)
      assert.are.same(new_backend.endpoints, instance.peers)

      assert_ewma_stats("10.10.10.1:8080", 0.2, ngx_now - 1)
      assert_ewma_stats("10.10.10.2:8080", 0.3, ngx_now - 5)
      assert_ewma_stats("10.10.10.3:8080", 1.2, ngx_now - 20)
    end)

    it("updates peers, deletes stats for old endpoints and sets average ewma score to new ones", function()
      local new_backend = util.deepcopy(backend)

      -- existing endpoint 10.10.10.2 got deleted
      -- and replaced with 10.10.10.4
      new_backend.endpoints[2].address = "10.10.10.4"
      -- and there's one new extra endpoint
      table.insert(new_backend.endpoints, { address = "10.10.10.5", port = "8080", maxFails = 0, failTimeout = 0 })

      instance:sync(new_backend)

      assert.are.same(new_backend.endpoints, instance.peers)

      assert_ewma_stats("10.10.10.1:8080", 0.2, ngx_now - 1)
      assert_ewma_stats("10.10.10.2:8080", nil, nil)
      assert_ewma_stats("10.10.10.3:8080", 1.2, ngx_now - 20)

      local slow_start_ewma = (0.2 + 1.2) / 2
      assert_ewma_stats("10.10.10.4:8080", slow_start_ewma, ngx_now)
      assert_ewma_stats("10.10.10.5:8080", slow_start_ewma, ngx_now)
    end)

    it("does not set slow_start_ewma when there is no existing ewma", function()
      local new_backend = util.deepcopy(backend)
      table.insert(new_backend.endpoints, { address = "10.10.10.4", port = "8080", maxFails = 0, failTimeout = 0 })

      -- when the LB algorithm instance is just instantiated it won't have any
      -- ewma value set for the initial endpoints (because it has not processed any request yet),
      -- this test is trying to simulate that by flushing existing ewma values
      flush_all_ewma_stats()

      instance:sync(new_backend)

      assert_ewma_stats("10.10.10.1:8080", nil, nil)
      assert_ewma_stats("10.10.10.2:8080", nil, nil)
      assert_ewma_stats("10.10.10.3:8080", nil, nil)
      assert_ewma_stats("10.10.10.4:8080", nil, nil)
    end)
  end)
end)
