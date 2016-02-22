local ngx_null = ngx.null
local tostring = tostring
local byte = string.byte
local gsub = string.gsub
local sort = table.sort
local pairs = pairs
local ipairs = ipairs
local concat = table.concat

local ok, new_tab = pcall(require, "table.new")
if not ok then
    new_tab = function (narr, nrec) return {} end
end

local _M = {}

local metachars = {
    ['\t'] = '\\t',
    ["\\"] = "\\\\",
    ['"'] = '\\"',
    ['\r'] = '\\r',
    ['\n'] = '\\n',
}

local function encode_str(s)
    -- XXX we will rewrite this when string.buffer is implemented
    -- in LuaJIT 2.1 because string.gsub cannot be JIT compiled.
    return gsub(s, '["\\\r\n\t]', metachars)
end

local function is_arr(t)
    local exp = 1
    for k, _ in pairs(t) do
        if k ~= exp then
            return nil
        end
        exp = exp + 1
    end
    return exp - 1
end

local encode

encode = function (v)
    if v == nil or v == ngx_null then
        return "null"
    end

    local typ = type(v)
    if typ == 'string' then
        return '"' .. encode_str(v) .. '"'
    end

    if typ == 'number' or typ == 'boolean' then
        return tostring(v)
    end

    if typ == 'table' then
        local n = is_arr(v)
        if n then
            local bits = new_tab(n, 0)
            for i, elem in ipairs(v) do
                bits[i] = encode(elem)
            end
            return "[" .. concat(bits, ",") .. "]"
        end

        local keys = {}
        local i = 0
        for key, _ in pairs(v) do
            i = i + 1
            keys[i] = key
        end
        sort(keys)

        local bits = new_tab(0, i)
        i = 0
        for _, key in ipairs(keys) do
            i = i + 1
            bits[i] = encode(key) .. ":" .. encode(v[key])
        end
        return "{" .. concat(bits, ",") .. "}"
    end

    return '"<' .. typ .. '>"'
end
_M.encode = encode

return _M
