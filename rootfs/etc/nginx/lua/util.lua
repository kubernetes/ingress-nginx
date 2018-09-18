local string_len = string.len
local string_sub = string.sub
local resty_str = require("resty.string")
local resty_sha1 = require("resty.sha1")
local resty_md5 = require("resty.md5")

local _M = {}

function _M.get_nodes(endpoints)
  local nodes = {}
  local weight = 1

  for _, endpoint in pairs(endpoints) do
    local endpoint_string = endpoint.address .. ":" .. endpoint.port
    nodes[endpoint_string] = weight
  end

  return nodes
end

local function hash_digest(hash_factory, message)
  local hash = hash_factory:new()
  if not hash then
    return nil, "failed to create object"
  end
  local ok = hash:update(message)
  if not ok then
    return nil, "failed to add data"
  end
  local binary_digest = hash:final()
  if binary_digest == nil then
    return nil, "failed to create digest"
  end
  return resty_str.to_hex(binary_digest), nil
end

function _M.sha1_digest(message)
  return hash_digest(resty_sha1, message)
end

function _M.md5_digest(message)
  return hash_digest(resty_md5, message)
end

-- given an Nginx variable i.e $request_uri
-- it returns value of ngx.var[request_uri]
function _M.lua_ngx_var(ngx_var)
  local var_name = string_sub(ngx_var, 2)
  if var_name:match("^%d+$") then
    var_name = tonumber(var_name)
  end

  return ngx.var[var_name]
end

-- this implementation is taken from
-- https://web.archive.org/web/20131225070434/http://snippets.luacode.org/snippets/Deep_Comparison_of_Two_Values_3
-- and modified for use in this project
local function deep_compare(t1, t2, ignore_mt)
  local ty1 = type(t1)
  local ty2 = type(t2)
  if ty1 ~= ty2 then return false end
  -- non-table types can be directly compared
  if ty1 ~= 'table' and ty2 ~= 'table' then return t1 == t2 end
  -- as well as tables which have the metamethod __eq
  local mt = getmetatable(t1)
  if not ignore_mt and mt and mt.__eq then return t1 == t2 end
  for k1,v1 in pairs(t1) do
    local v2 = t2[k1]
    if v2 == nil or not deep_compare(v1,v2) then return false end
  end
  for k2,v2 in pairs(t2) do
    local v1 = t1[k2]
    if v1 == nil or not deep_compare(v1,v2) then return false end
  end
  return true
end
_M.deep_compare = deep_compare

function _M.is_blank(str)
  return str == nil or string_len(str) == 0
end

-- this implementation is taken from:
-- https://github.com/luafun/luafun/blob/master/fun.lua#L33
-- SHA: 04c99f9c393e54a604adde4b25b794f48104e0d0
local function deepcopy(orig)
  local orig_type = type(orig)
  local copy
  if orig_type == 'table' then
      copy = {}
      for orig_key, orig_value in next, orig, nil do
          copy[deepcopy(orig_key)] = deepcopy(orig_value)
      end
  else
      copy = orig
  end
  return copy
end
_M.deepcopy = deepcopy

local function tablelength(T)
  local count = 0
  for _ in pairs(T) do
      count = count + 1
  end
  return count
end
_M.tablelength = tablelength

-- replaces special character value a with value b for all occurences in a string
local function replace_special_char(str, a, b)
  return string.gsub(str, "%" .. a, b)
end
_M.replace_special_char = replace_special_char

return _M
