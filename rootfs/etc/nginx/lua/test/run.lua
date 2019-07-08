local ffi = require("ffi")
local lua_ingress = require("lua_ingress")

-- without this we get errors such as "attempt to redefine XXX"
local old_cdef = ffi.cdef
local exists = {}
ffi.cdef = function(def)
  if exists[def] then
    return
  end
  exists[def] = true
  return old_cdef(def)
end

local old_udp = ngx.socket.udp
ngx.socket.udp = function(...)
  local socket = old_udp(...)
  socket.send = function(...)
    error("ngx.socket.udp:send please mock this to use in tests")
  end
  return socket
end

local old_tcp = ngx.socket.tcp
ngx.socket.tcp = function(...)
  local socket = old_tcp(...)
  socket.send = function(...)
    error("ngx.socket.tcp:send please mock this to use in tests")
  end
  return socket
end

ngx.log = function(...) end
ngx.print = function(...) end

lua_ingress.init_worker()

require "busted.runner"({ standalone = false })
