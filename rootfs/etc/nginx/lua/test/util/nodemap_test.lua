local util = require("util")
local nodemap = require("util.nodemap")

local function get_test_backend_single()
  return {
    name = "access-router-production-web-80",
    endpoints = {
      { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 }
    }
  }
end

local function get_test_backend_multi()
  return {
    name = "access-router-production-web-80",
    endpoints = {
      { address = "10.184.7.40", port = "8080", maxFails = 0, failTimeout = 0 },
      { address = "10.184.7.41", port = "8080", maxFails = 0, failTimeout = 0 }
    }
  }
end

local function get_test_nodes_ignore(endpoint)
  local ignore = {}
  ignore[endpoint] = true
  return ignore
end

describe("Node Map", function()

  local test_backend_single = get_test_backend_single()
  local test_backend_multi = get_test_backend_multi()
  local test_salt = test_backend_single.name
  local test_nodes_single = util.get_nodes(test_backend_single.endpoints)
  local test_nodes_multi = util.get_nodes(test_backend_multi.endpoints)
  local test_endpoint1 = test_backend_multi.endpoints[1].address .. ":" .. test_backend_multi.endpoints[1].port
  local test_endpoint2 = test_backend_multi.endpoints[2].address .. ":" .. test_backend_multi.endpoints[2].port
  local test_nodes_ignore = get_test_nodes_ignore(test_endpoint1)

  describe("new()", function()
    context("when no salt has been provided", function()
      it("random() returns an unsalted key", function()
        local nodemap_instance = nodemap:new(test_nodes_single, nil)
        local expected_endpoint = test_endpoint1
        local expected_hash_key = ngx.md5(expected_endpoint)
        local actual_endpoint
        local actual_hash_key

        actual_endpoint, actual_hash_key = nodemap_instance:random()

        assert.equal(actual_endpoint, expected_endpoint)
        assert.equal(expected_hash_key, actual_hash_key)
      end)
    end)

    context("when a salt has been provided", function()
      it("random() returns a salted key", function()
        local nodemap_instance = nodemap:new(test_nodes_single, test_salt)
        local expected_endpoint = test_endpoint1
        local expected_hash_key = ngx.md5(test_salt .. expected_endpoint)
        local actual_endpoint
        local actual_hash_key

        actual_endpoint, actual_hash_key = nodemap_instance:random()

        assert.equal(actual_endpoint, expected_endpoint)
        assert.equal(expected_hash_key, actual_hash_key)
      end)
    end)

    context("when no nodes have been provided", function()
      it("random() returns nil", function()
        local nodemap_instance = nodemap:new({}, test_salt)
        local actual_endpoint
        local actual_hash_key

        actual_endpoint, actual_hash_key = nodemap_instance:random()

        assert.equal(actual_endpoint, nil)
        assert.equal(expected_hash_key, nil)
      end)
    end)
  end)

  describe("find()", function()
    before_each(function()
      package.loaded["util.nodemap"] = nil
      nodemap = require("util.nodemap")
    end)

    context("when a hash key is valid", function()
      it("find() returns the correct endpoint", function()
        local nodemap_instance = nodemap:new(test_nodes_single, test_salt)
        local test_hash_key
        local expected_endpoint
        local actual_endpoint

        expected_endpoint, test_hash_key = nodemap_instance:random()
        assert.not_equal(expected_endpoint, nil)
        assert.not_equal(test_hash_key, nil)

        actual_endpoint = nodemap_instance:find(test_hash_key)
        assert.equal(actual_endpoint, expected_endpoint)
      end)
    end)

    context("when a hash key is invalid", function()
      it("find() returns nil", function()
        local nodemap_instance = nodemap:new(test_nodes_single, test_salt)
        local test_hash_key = "invalid or nonexistent hash key"
        local actual_endpoint

        actual_endpoint = nodemap_instance:find(test_hash_key)

        assert.equal(actual_endpoint, nil)
      end)
    end)
  end)


  describe("random_except()", function()
    before_each(function()
      package.loaded["util.nodemap"] = nil
      nodemap = require("util.nodemap")
    end)

    context("when nothing has been excluded", function()
      it("random_except() returns the correct endpoint", function()
        local nodemap_instance = nodemap:new(test_nodes_single, test_salt)
        local expected_endpoint = test_endpoint1
        local test_hash_key
        local actual_endpoint

        actual_endpoint, test_hash_key = nodemap_instance:random_except({})
        assert.equal(expected_endpoint, actual_endpoint)
        assert.not_equal(test_hash_key, nil)
      end)
    end)

    context("when everything has been excluded", function()
      it("random_except() returns nil", function()
        local nodemap_instance = nodemap:new(test_nodes_single, test_salt)
        local actual_hash_key
        local actual_endpoint

        actual_endpoint, actual_hash_key = nodemap_instance:random_except(test_nodes_ignore)

        assert.equal(actual_endpoint, nil)
        assert.equal(actual_hash_key, nil)
      end)
    end)

    context("when an endpoint has been excluded", function()
      it("random_except() does not return it", function()
        local nodemap_instance = nodemap:new(test_nodes_multi, test_salt)
        local expected_endpoint = test_endpoint2
        local actual_endpoint
        local test_hash_key

        actual_endpoint, test_hash_key = nodemap_instance:random_except(test_nodes_ignore)

        assert.equal(actual_endpoint, expected_endpoint)
        assert.not_equal(test_hash_key, nil)
      end)
    end)
  end)
end)
