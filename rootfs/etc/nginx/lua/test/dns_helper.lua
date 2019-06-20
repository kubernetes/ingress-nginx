local _M = {}

local configuration = require("configuration")
local resolver = require("resty.dns.resolver")
local old_resolver_new = resolver.new

local function reset(nameservers)
  configuration.nameservers = nameservers or { "1.1.1.1" }
end

function _M.mock_new(func, nameservers)
  reset(nameservers)
  resolver.new = func
end

function _M.mock_dns_query(response, err)
  reset()
  resolver.new = function(self, options)
    local r = old_resolver_new(self, options)
    r.query = function(self, name, options, tries)
      return response, err
    end
    return r
  end
end

return _M
