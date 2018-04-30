local resty_chash = require("resty.chash")
local util = require("util")
local string_sub = string.sub

local _M = {}
local instances = {}

-- given an Nginx variable i.e $request_uri
-- it returns value of ngx.var[request_uri]
local function get_lua_ngx_var(ngx_var)
  local var_name = string_sub(ngx_var, 2)
  return ngx.var[var_name]
end

function _M.balance(backend)
  local instance = instances[backend.name]
  if not instance then
    return nil
  end

  local key = get_lua_ngx_var(backend["upstream-hash-by"])
  local endpoint_string = instance:find(key)

  local address, port = util.split_pair(endpoint_string, ":")
  return { address = address, port = port }
end

function _M.reinit(backend)
  local instance = instances[backend.name]

  local nodes = {}
  -- we don't support weighted consistent hashing
  local weight = 1

  for _, endpoint in pairs(backend.endpoints) do
    local endpoint_string = endpoint.address .. ":" .. endpoint.port
    nodes[endpoint_string] = weight
  end

  if instance then
    instance:reinit(nodes)
  else
    instance = resty_chash:new(nodes)
    instances[backend.name] = instance
  end
end

return _M
