local util = require("util")
local say  = require("say")

local original_ngx = ngx
local function reset_ngx()
  _G.ngx = original_ngx
end

local function included_in(state, arguments)
   if not type(arguments[1]) == "table" or #arguments ~= 2 then
    return false
  end

  local table = arguments[1]
  for _, value in pairs(table) do
    if value == arguments[2] then
      return true
    end
  end
  return false
end
assert:register("assertion", "included_in", included_in, "assertion.has_property.positive", "assertion.has_property.negative")

local function mock_ngx(mock)
  local _ngx = mock
  setmetatable(_ngx, { __index = ngx })
  _G.ngx = _ngx
end

local function flush_connection_count()
  ngx.shared.balancer_leastconn:flush_all()
end

local function set_backend_count(endpoint_string, count)
  ngx.shared.balancer_leastconn:set(endpoint_string, count)
end

describe("Balancer leastconn", function()
  local balancer_leastconn = require("balancer.leastconn")
  local ngx_now = 1543238266
  local backend, instance

  before_each(function()
    package.loaded["balancer.leastconn"] = nil
    balancer_leastconn = require("balancer.leastconn")

    backend = {
      name = "namespace-service-port",
      ["load-balance"] = "least_connections",
      endpoints = {
        { address = "10.10.10.1", port = "8080" },
        { address = "10.10.10.2", port = "8080" },
        { address = "10.10.10.3", port = "8080" },
      }
    }
    set_backend_count("10.10.10.1:8080", 0)
    set_backend_count("10.10.10.2:8080", 1)
    set_backend_count("10.10.10.3:8080", 5)

    instance = balancer_leastconn:new(backend)
  end)

  after_each(function()
    reset_ngx()
    flush_connection_count()
  end)

  describe("after_balance()", function()
    it("updates connection count", function()
      ngx.var = { upstream_addr = "10.10.10.2:8080" }

      local count_before = ngx.shared.balancer_leastconn:get(ngx.var.upstream_addr)
      instance:after_balance()
      local count_after = ngx.shared.balancer_leastconn:get(ngx.var.upstream_addr)

      assert.are.equals(count_before - 1, count_after)
    end)
  end)

  describe("balance()", function()
    it("increments connection count on selected peer", function()
      local single_endpoint_backend = util.deepcopy(backend)
      table.remove(single_endpoint_backend.endpoints, 3)
      table.remove(single_endpoint_backend.endpoints, 2)
      local single_endpoint_instance = balancer_leastconn:new(single_endpoint_backend)

      local upstream = single_endpoint_backend.endpoints[1]
      local upstream_name = upstream.address .. ":" .. upstream.port

      set_backend_count(upstream_name, 0)
      single_endpoint_instance:balance()
      local count_after = ngx.shared.balancer_leastconn:get(upstream_name)

      assert.are.equals(1, count_after)
    end)

    it("returns single endpoint when the given backend has only one endpoint", function()
      local single_endpoint_backend = util.deepcopy(backend)
      table.remove(single_endpoint_backend.endpoints, 3)
      table.remove(single_endpoint_backend.endpoints, 2)
      local single_endpoint_instance = balancer_leastconn:new(single_endpoint_backend)

      local peer = single_endpoint_instance:balance()

      assert.are.equals("10.10.10.1:8080", peer)
    end)

    it("picks the endpoint with lowest connection count", function()
      local two_endpoints_backend = util.deepcopy(backend)
      table.remove(two_endpoints_backend.endpoints, 2)
      local two_endpoints_instance = balancer_leastconn:new(two_endpoints_backend)

      local peer = two_endpoints_instance:balance()

      assert.equal("10.10.10.1:8080", peer)
    end)

    it("picks one of the endpoints with tied lowest connection count", function()
      set_backend_count("10.10.10.1:8080", 8)
      set_backend_count("10.10.10.2:8080", 5)
      set_backend_count("10.10.10.3:8080", 5)

      local peer = instance:balance()
      assert.included_in({"10.10.10.2:8080", "10.10.10.3:8080"}, peer)
    end)

  end)

  describe("sync()", function()
    it("does not reset stats when endpoints do not change", function()
      local new_backend = util.deepcopy(backend)

      instance:sync(new_backend)

      assert.are.same(new_backend.endpoints, instance.peers)
      assert.are.same(new_backend.endpoints, backend.endpoints)
    end)

    it("updates peers, deletes stats for old endpoints and sets connection count to zero for new ones", function()
      local new_backend = util.deepcopy(backend)

      -- existing endpoint 10.10.10.2 got deleted
      -- and replaced with 10.10.10.4
      new_backend.endpoints[2].address = "10.10.10.4"
      -- and there's one new extra endpoint
      table.insert(new_backend.endpoints, { address = "10.10.10.5", port = "8080" })

      instance:sync(new_backend)

      assert.are.same(new_backend.endpoints, instance.peers)

      assert.are.equals(ngx.shared.balancer_leastconn:get("10.10.10.1:8080"), 0)
      assert.are.equals(ngx.shared.balancer_leastconn:get("10.10.10.2:8080"), nil)
      assert.are.equals(ngx.shared.balancer_leastconn:get("10.10.10.3:8080"), 5)
      assert.are.equals(ngx.shared.balancer_leastconn:get("10.10.10.4:8080"), 0)
      assert.are.equals(ngx.shared.balancer_leastconn:get("10.10.10.5:8080"), 0)
    end)
  end)

end)