-- Simple trie for URLs

local _M = {}

local mt = {
    __index = _M
}

-- http://lua-users.org/wiki/SplitJoin
local strfind, tinsert, strsub = string.find, table.insert, string.sub
function _M.strsplit(delimiter, text)
    local list = {}
    local pos = 1
    while 1 do
        local first, last = strfind(text, delimiter, pos)
        if first then -- found?
            tinsert(list, strsub(text, pos, first-1))
            pos = last+1
        else
            tinsert(list, strsub(text, pos))
            break
        end
    end
    return list
end

local strsplit = _M.strsplit

function _M.new()
    local t = { }
    return setmetatable(t, mt)
end

function _M.add(t, key, val)
    local parts = {}
    -- hack for just /
    if key == "/" then
        parts = { "" }
    else
        parts = strsplit("/", key)
    end

    local l = t
    for i = 1, #parts do
        local p = parts[i]
        if not l[p] then
            l[p] = {}
        end
        l =  l[p]
    end
    l.__value = val
end

function _M.get(t, key)
    local parts = strsplit("/", key)

    local l = t

    -- this may be nil
    local val = t.__value
    for i = 1, #parts do
        local p = parts[i]
        if l[p] then
            l = l[p]
            local v = l.__value
            if v then
                val = v
            end
        else
            break
        end
    end

    -- may be nil
    return val
end

return _M
