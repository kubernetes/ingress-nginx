
local function get_test_backend(n_endpoints) 
   local backend = {
    name = "my-dummy-backend", 
    ["upstreamHashByConfig"] = {
      ["upstream-hash-by"] = "$request_uri",
      ["upstream-hash-by-subset"] = true,
      ["upstream-hash-by-subset-size"] = 3
    },
    endpoints = {}
  }

  for i = 1, n_endpoints do
    backend.endpoints[i] = { address = "10.184.7." .. tostring(i), port = "8080", maxFails = 0, failTimeout = 0 }
  end

  return backend
end
 
describe("Balancer chash subset", function()
  local balancer_chashsubset = require("balancer.chashsubset")

  describe("balance()", function()
    it("returns peers from the same subset", function()
      _G.ngx = { var = { request_uri = "/alma/armud" }}

      local backend = get_test_backend(9)
      
      local instance = balancer_chashsubset:new(backend)

      instance:sync(backend)

      local first_node = instance:balance()
      local subset_id
      local endpoint_strings

      local function has_value (tab, val)
        for _, value in ipairs(tab) do
          if value == val then
            return true
           end
        end

        return false
      end

      for id, endpoints in pairs(instance["subsets"]) do
        endpoint_strings = {}
        for _, endpoint in pairs(endpoints) do
          local endpoint_string = endpoint.address .. ":" .. endpoint.port
          table.insert(endpoint_strings, endpoint_string)
          if first_node == endpoint_string then
            -- found the set of first_node
            subset_id = id
          end
        end
        if subset_id then
          break
        end
      end

      -- multiple calls to balance must return nodes from the same subset
      for i = 0, 10 do
        assert.True(has_value(endpoint_strings, instance:balance()))
      end
    end)
  end)
  describe("new(backend)", function()
    it("fills last subset correctly", function()
      _G.ngx = { var = { request_uri = "/alma/armud" }}

      local backend = get_test_backend(7)
      
      local instance = balancer_chashsubset:new(backend)

      instance:sync(backend)
      for id, endpoints in pairs(instance["subsets"]) do
        assert.are.equal(#endpoints, 3)
      end
    end)
  end)
end)
