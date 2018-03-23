local _M = {}
local string_len = string.len

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

-- http://nginx.org/en/docs/http/ngx_http_upstream_module.html#example
-- CAVEAT: nginx is giving out : instead of , so the docs are wrong
-- 127.0.0.1:26157 : 127.0.0.1:26157 , ngx.var.upstream_addr
-- 200 : 200 , ngx.var.upstream_status
-- 0.00 : 0.00, ngx.var.upstream_response_time
function _M.split_upstream_var(var)
  if not var then
    return nil, nil
  end
  local t = {}
  for v in var:gmatch("[^%s|,]+") do
    if v ~= ":" then
      t[#t+1] = v
    end
  end
  return t
end

function _M.get_first_value(var)
  local t = _M.split_upstream_var(var) or {}
  if #t == 0 then return nil end
  return t[1]
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

return _M
