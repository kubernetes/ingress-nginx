local ngx_re_split = require("ngx.re").split
local string_format = string.format
local tonumber = tonumber

local ngx_log = ngx.log
local ngx_ERR = ngx.ERR

local CONF_PATH = "/etc/resolv.conf"

local nameservers, search, ndots = {}, {}, 1

local function set_search(parts)
  local length = #parts

  for i = 2, length, 1 do
    search[i-1] = parts[i]
  end
end

local function set_ndots(parts)
  local option = parts[2]
  if not option then
    return
  end

  local option_parts, err = ngx_re_split(option, ":")
  if err then
    ngx_log(ngx_ERR, err)
    return
  end

  if option_parts[1] ~= "ndots" then
    return
  end

  ndots = tonumber(option_parts[2])
end

local function is_comment(line)
  return line:sub(1, 1) == "#"
end

local function parse_line(line)
  if is_comment(line) then
    return
  end

  local parts, err = ngx_re_split(line, "\\s+")
  if err then
    ngx_log(ngx_ERR, err)
  end

  local keyword, value = parts[1], parts[2]

  if keyword == "nameserver" then
    if not value:match("^%d+.%d+.%d+.%d+$") then
      value = string_format("[%s]", value)
    end
    nameservers[#nameservers + 1] = value
  elseif keyword == "search" then
    set_search(parts)
  elseif keyword == "options" then
    set_ndots(parts)
  end
end

do
  local f, err = io.open(CONF_PATH, "r")
  if not f then
    error("could not open " .. CONF_PATH .. ": " .. tostring(err))
  end

  for line in f:lines() do
    parse_line(line)
  end

  f:close()
end

return {
  nameservers = nameservers,
  search = search,
  ndots = ndots,
}
