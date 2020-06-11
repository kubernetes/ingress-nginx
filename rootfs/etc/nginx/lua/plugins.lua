local require = require
local ngx = ngx
local pairs = pairs
local ipairs = ipairs
local string_format = string.format
local ngx_log = ngx.log
local INFO = ngx.INFO
local ERR = ngx.ERR
local pcall = pcall

local _M = {}
local MAX_NUMBER_OF_PLUGINS = 20
local plugins = {}

local function load_plugin(name)
  local path = string_format("plugins.%s.main", name)

  local ok, plugin = pcall(require, path)
  if not ok then
    ngx_log(ERR, string_format("error loading plugin \"%s\": %s", path, plugin))
    return
  end

  plugins[name] = plugin
end

function _M.init(names)
  local count = 0
  for _, name in ipairs(names) do
    if count >= MAX_NUMBER_OF_PLUGINS then
      ngx_log(ERR, "the total number of plugins exceed the maximum number: ", MAX_NUMBER_OF_PLUGINS)
      break
    end
    load_plugin(name)
    count = count + 1 -- ignore loading failure, just count the total
  end
end

function _M.run()
  local phase = ngx.get_phase()

  for name, plugin in pairs(plugins) do
    if plugin[phase] then
      ngx_log(INFO, string_format("running plugin \"%s\" in phase \"%s\"", name, phase))

      -- TODO: consider sandboxing this, should we?
      -- probably yes, at least prohibit plugin from accessing env vars etc
      -- but since the plugins are going to be installed by ingress-nginx
      -- operator they can be assumed to be safe also
      local ok, err = pcall(plugin[phase])
      if not ok then
        ngx_log(ERR, string_format("error while running plugin \"%s\" in phase \"%s\": %s",
            name, phase, err))
      end
    end
  end
end

return _M
