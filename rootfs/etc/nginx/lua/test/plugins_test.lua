describe("plugins", function()
  describe("#run", function()
    it("runs the plugins in the given order", function()
      ngx.get_phase = function() return "rewrite" end
      local plugins = require("plugins")
      local called_plugins = {}
      local plugins_to_mock = {"plugins.pluginfirst.main", "plugins.pluginsecond.main", "plugins.pluginthird.main"}
      for i=1, 3, 1
      do
        package.loaded[plugins_to_mock[i]] = {
          rewrite = function()
            called_plugins[#called_plugins + 1] = plugins_to_mock[i]
          end
        }
      end
      assert.has_no.errors(function()
        plugins.init({"pluginfirst", "pluginsecond", "pluginthird"})
      end)
      assert.has_no.errors(plugins.run)
      assert.are.same(plugins_to_mock, called_plugins)
    end)
  end)

  describe("#init", function()
    it("marks hello_world plugin as a balancer implementation", function()
      local plugins = require("plugins")
      local balancer_implementations = plugins.balancer_implementations
      assert.is_nil(balancer_implementations["hello_world"])
      assert.has_no.errors(function()
        plugins.init({"hello_world"})
      end)
      assert.is_table(balancer_implementations["hello_world"])
      assert.is_function(balancer_implementations["hello_world"].new)
      assert.is_function(balancer_implementations["hello_world"].sync)
      assert.is_function(balancer_implementations["hello_world"].balance)
    end)
  end)
end)
