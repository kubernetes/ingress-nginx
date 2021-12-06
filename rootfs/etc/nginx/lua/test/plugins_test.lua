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
end)