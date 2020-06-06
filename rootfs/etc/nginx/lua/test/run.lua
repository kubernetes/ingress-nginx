local busted_runner
do 
  -- avoid warning during test runs caused by
  -- https://github.com/openresty/lua-nginx-module/blob/2524330e59f0a385a9c77d4d1b957476dce7cb33/src/ngx_http_lua_util.c#L810

  local traceback = require "debug".traceback

  setmetatable(_G, { __newindex = function(table, key, value) rawset(table, key, value) end })
  busted_runner = require "busted.runner"

  -- if there's more constants need to be whitelisted for test runs, add here.
  local GLOBALS_ALLOWED_IN_TEST = {
    helpers = true,
  }
  local newindex = function(table, key, value)
    rawset(table, key, value)

    local phase = ngx.get_phase()
    if phase == "init_worker" or phase == "init" then
      return
    end

    -- we check only timer phase because resty-cli runs everything in timer phase
    if phase == "timer" and GLOBALS_ALLOWED_IN_TEST[key] then
      return
    end

    local message = "writing a global lua variable " .. key ..
      " which may lead to race conditions between concurrent requests, so prefer the use of 'local' variables " .. traceback('', 2)
    -- it's important to do print here because ngx.log is mocked below
    print(message)
  end
  setmetatable(_G, { __newindex = newindex })
end

_G.helpers = require("test.helpers")

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

busted_runner({ standalone = false })
