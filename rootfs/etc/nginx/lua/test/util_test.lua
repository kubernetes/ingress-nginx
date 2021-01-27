local util

describe("utility", function()
  before_each(function()
    ngx.var = { remote_addr = "192.168.1.1", [1] = "nginx/regexp/1/group/capturing" }
    util = require_without_cache("util")
  end)

  after_each(function()
    reset_ngx()
  end)

  describe("ngx_complex_value", function()

    local ngx_complex_value = function(data)
      local ret, err = util.parse_complex_value(data)
      if err ~= nil then
        return ""
      end
      return util.generate_var_value(ret)
    end

    it("returns value of nginx var by key", function()
      assert.equal("192.168.1.1", ngx_complex_value("$remote_addr"))
    end)
 
    it("returns value of nginx var when key is number", function()
      assert.equal("nginx/regexp/1/group/capturing", ngx_complex_value("$1"))
    end)

    it("returns value of nginx var by multiple variables", function()
      assert.equal("192.168.1.1nginx/regexp/1/group/capturing", ngx_complex_value("$remote_addr$1"))
    end)

    it("returns value by the combination of variable and text value", function()
      assert.equal("192.168.1.1-text-value", ngx_complex_value("${remote_addr}-text-value"))
    end)

    it("returns empty when variable is not defined", function()
      assert.equal("", ngx_complex_value("$foo_bar"))
    end)
  end)

  describe("diff_endpoints", function()
    it("returns removed and added endpoints", function()
      local old = {
        { address = "10.10.10.1", port = "8080" },
        { address = "10.10.10.2", port = "8080" },
        { address = "10.10.10.3", port = "8080" },
      }
      local new = {
        { address = "10.10.10.1", port = "8080" },
        { address = "10.10.10.2", port = "8081" },
        { address = "11.10.10.2", port = "8080" },
        { address = "11.10.10.3", port = "8080" },
      }
      local expected_added = { "10.10.10.2:8081", "11.10.10.2:8080", "11.10.10.3:8080" }
      table.sort(expected_added)
      local expected_removed = { "10.10.10.2:8080", "10.10.10.3:8080" }
      table.sort(expected_removed)

      local added, removed = util.diff_endpoints(old, new)
      table.sort(added)
      table.sort(removed)

      assert.are.same(expected_added, added)
      assert.are.same(expected_removed, removed)
    end)

    it("returns empty results for empty inputs", function()
      local added, removed = util.diff_endpoints({}, {})

      assert.are.same({}, added)
      assert.are.same({}, removed)
    end)

    it("returns empty results for same inputs", function()
      local old = {
        { address = "10.10.10.1", port = "8080" },
        { address = "10.10.10.2", port = "8080" },
        { address = "10.10.10.3", port = "8080" },
      }
      local new = util.deepcopy(old)

      local added, removed = util.diff_endpoints(old, new)

      assert.are.same({}, added)
      assert.are.same({}, removed)
    end)

    it("handles endpoints with nil attribute", function()
      local old = {
        { address = nil, port = "8080" },
        { address = "10.10.10.2", port = "8080" },
        { address = "10.10.10.3", port = "8080" },
      }
      local new = util.deepcopy(old)
      new[2].port = nil

      local added, removed = util.diff_endpoints(old, new)
      assert.are.same({ "10.10.10.2:nil" }, added)
      assert.are.same({ "10.10.10.2:8080" }, removed)
    end)
  end)
end)
