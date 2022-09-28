local _M = {}

local resty_dns_resolver = require("resty.dns.resolver")

local original_resty_dns_resolver_new = resty_dns_resolver.new
local original_io_open = io.open

function _M.with_resolv_conf(content, func)
  local new_resolv_conf_f = assert(io.tmpfile())
  new_resolv_conf_f:write(content)
  new_resolv_conf_f:seek("set", 0)

  io.open = function(path, mode)
    if path ~= "/etc/resolv.conf" then
      error("expected '/etc/resolv.conf' as path but got: " .. tostring(path))
    end
    if mode ~= "r" then
      error("expected 'r' as mode but got: " .. tostring(mode))
    end

    return new_resolv_conf_f, nil
  end

  func()

  io.open = original_io_open

  if io.type(new_resolv_conf_f) ~= "closed file" then
    error("file was left open")
  end
end

function _M.mock_resty_dns_new(func)
  resty_dns_resolver.new = func
end

function _M.mock_resty_dns_query(mocked_host, response, err)
  resty_dns_resolver.new = function(self, options)
    local r = original_resty_dns_resolver_new(self, options)
    r.query = function(self, host, options, tries)
      if mocked_host and mocked_host ~= host then
        return error(tostring(host) .. " is not mocked")
      end
      return response, err
    end
    return r
  end
end

return _M
