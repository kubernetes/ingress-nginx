-- Copyright (C) Yichun Zhang (agentzh)


-- local socket = require "socket"
local bit = require "bit"
local udp = ngx.socket.udp
local rand = math.random
local char = string.char
local byte = string.byte
local find = string.find
local gsub = string.gsub
local sub = string.sub
local format = string.format
local band = bit.band
local rshift = bit.rshift
local lshift = bit.lshift
local insert = table.insert
local concat = table.concat
local re_sub = ngx.re.sub
local tcp = ngx.socket.tcp
local log = ngx.log
local DEBUG = ngx.DEBUG
local randomseed = math.randomseed
local ngx_time = ngx.time
local setmetatable = setmetatable
local type = type


local DOT_CHAR = byte(".")


local TYPE_A      = 1
local TYPE_NS     = 2
local TYPE_CNAME  = 5
local TYPE_PTR    = 12
local TYPE_MX     = 15
local TYPE_TXT    = 16
local TYPE_AAAA   = 28
local TYPE_SRV    = 33
local TYPE_SPF    = 99

local CLASS_IN    = 1


local _M = {
    _VERSION    = '0.14',
    TYPE_A      = TYPE_A,
    TYPE_NS     = TYPE_NS,
    TYPE_CNAME  = TYPE_CNAME,
    TYPE_PTR    = TYPE_PTR,
    TYPE_MX     = TYPE_MX,
    TYPE_TXT    = TYPE_TXT,
    TYPE_AAAA   = TYPE_AAAA,
    TYPE_SRV    = TYPE_SRV,
    TYPE_SPF    = TYPE_SPF,
    CLASS_IN    = CLASS_IN,
}


local resolver_errstrs = {
    "format error",     -- 1
    "server failure",   -- 2
    "name error",       -- 3
    "not implemented",  -- 4
    "refused",          -- 5
}


local mt = { __index = _M }


function _M.new(class, opts)
    if not opts then
        return nil, "no options table specified"
    end

    local servers = opts.nameservers
    if not servers or #servers == 0 then
        return nil, "no nameservers specified"
    end

    local timeout = opts.timeout or 2000  -- default 2 sec

    local n = #servers

    local socks = {}

    for i = 1, n do
        local server = servers[i]
        local sock, err = udp()
        if not sock then
            return nil, "failed to create udp socket: " .. err
        end

        local host, port
        if type(server) == 'table' then
            host = server[1]
            port = server[2] or 53

        else
            host = server
            port = 53
            servers[i] = {host, port}
        end

        local ok, err = sock:setpeername(host, port)
        if not ok then
            return nil, "failed to set peer name: " .. err
        end

        sock:settimeout(timeout)

        insert(socks, sock)
    end

    local tcp_sock, err = tcp()
    if not tcp_sock then
        return nil, "failed to create tcp socket: " .. err
    end

    tcp_sock:settimeout(timeout)

    return setmetatable(
                { cur = rand(1, n), socks = socks,
                  tcp_sock = tcp_sock,
                  servers = servers,
                  retrans = opts.retrans or 5,
                  no_recurse = opts.no_recurse,
                }, mt)
end


local function pick_sock(self, socks)
    local cur = self.cur

    if cur == #socks then
        self.cur = 1
    else
        self.cur = cur + 1
    end

    return socks[cur]
end


local function _get_cur_server(self)
    local cur = self.cur

    local servers = self.servers

    if cur == 1 then
        return servers[#servers]
    end

    return servers[cur - 1]
end


function _M.set_timeout(self, timeout)
    local socks = self.socks
    if not socks then
        return nil, "not initialized"
    end

    for i = 1, #socks do
        local sock = socks[i]
        sock:settimeout(timeout)
    end

    local tcp_sock = self.tcp_sock
    if not tcp_sock then
        return nil, "not initialized"
    end

    tcp_sock:settimeout(timeout)
end


local function _encode_name(s)
    return char(#s) .. s
end


local function _decode_name(buf, pos)
    local labels = {}
    local nptrs = 0
    local p = pos
    while nptrs < 128 do
        local fst = byte(buf, p)

        if not fst then
            return nil, 'truncated';
        end

        -- print("fst at ", p, ": ", fst)

        if fst == 0 then
            if nptrs == 0 then
                pos = pos + 1
            end
            break
        end

        if band(fst, 0xc0) ~= 0 then
            -- being a pointer
            if nptrs == 0 then
                pos = pos + 2
            end

            nptrs = nptrs + 1

            local snd = byte(buf, p + 1)
            if not snd then
                return nil, 'truncated'
            end

            p = lshift(band(fst, 0x3f), 8) + snd + 1

            -- print("resolving ptr ", p, ": ", byte(buf, p))

        else
            -- being a label
            local label = sub(buf, p + 1, p + fst)
            insert(labels, label)

            -- print("resolved label ", label)

            p = p + fst + 1

            if nptrs == 0 then
                pos = p
            end
        end
    end

    return concat(labels, "."), pos
end


local function _build_request(qname, id, no_recurse, opts)
    local qtype

    if opts then
        qtype = opts.qtype
    end

    if not qtype then
        qtype = 1  -- A record
    end

    local ident_hi = char(rshift(id, 8))
    local ident_lo = char(band(id, 0xff))

    local flags
    if no_recurse then
        -- print("found no recurse")
        flags = "\0\0"
    else
        flags = "\1\0"
    end

    local nqs = "\0\1"
    local nan = "\0\0"
    local nns = "\0\0"
    local nar = "\0\0"
    local typ = "\0" .. char(qtype)
    local class = "\0\1"    -- the Internet class

    if byte(qname, 1) == DOT_CHAR then
        return nil, "bad name"
    end

    local name = gsub(qname, "([^.]+)%.?", _encode_name) .. '\0'

    return {
        ident_hi, ident_lo, flags, nqs, nan, nns, nar,
        name, typ, class
    }
end


local function parse_response(buf, id)
    local n = #buf
    if n < 12 then
        return nil, 'truncated';
    end

    -- header layout: ident flags nqs nan nns nar

    local ident_hi = byte(buf, 1)
    local ident_lo = byte(buf, 2)
    local ans_id = lshift(ident_hi, 8) + ident_lo

    -- print("id: ", id, ", ans id: ", ans_id)

    if ans_id ~= id then
        -- identifier mismatch and throw it away
        log(DEBUG, "id mismatch in the DNS reply: ", ans_id, " ~= ", id)
        return nil, "id mismatch"
    end

    local flags_hi = byte(buf, 3)
    local flags_lo = byte(buf, 4)
    local flags = lshift(flags_hi, 8) + flags_lo

    -- print(format("flags: 0x%x", flags))

    if band(flags, 0x8000) == 0 then
        return nil, format("bad QR flag in the DNS response")
    end

    if band(flags, 0x200) ~= 0 then
        return nil, "truncated"
    end

    local code = band(flags, 0x7f)

    -- print(format("code: %d", code))

    local nqs_hi = byte(buf, 5)
    local nqs_lo = byte(buf, 6)
    local nqs = lshift(nqs_hi, 8) + nqs_lo

    -- print("nqs: ", nqs)

    if nqs ~= 1 then
        return nil, format("bad number of questions in DNS response: %d", nqs)
    end

    local nan_hi = byte(buf, 7)
    local nan_lo = byte(buf, 8)
    local nan = lshift(nan_hi, 8) + nan_lo

    -- print("nan: ", nan)

    -- skip the question part

    local ans_qname, pos = _decode_name(buf, 13)
    if not ans_qname then
        return nil, pos
    end

    -- print("qname in reply: ", ans_qname)

    -- print("question: ", sub(buf, 13, pos))

    if pos + 3 + nan * 12 > n then
        -- print(format("%d > %d", pos + 3 + nan * 12, n))
        return nil, 'truncated';
    end

    -- question section layout: qname qtype(2) qclass(2)

    local type_hi = byte(buf, pos)
    local type_lo = byte(buf, pos + 1)
    local ans_type = lshift(type_hi, 8) + type_lo

    -- print("ans qtype: ", ans_type)

    local class_hi = byte(buf, pos + 2)
    local class_lo = byte(buf, pos + 3)
    local qclass = lshift(class_hi, 8) + class_lo

    -- print("ans qclass: ", qclass)

    if qclass ~= 1 then
        return nil, format("unknown query class %d in DNS response", qclass)
    end

    pos = pos + 4

    local answers = {}

    if code ~= 0 then
        answers.errcode = code
        answers.errstr = resolver_errstrs[code] or "unknown"
    end

    for i = 1, nan do
        -- print(format("ans %d: qtype:%d qclass:%d", i, qtype, qclass))

        local ans = {}
        insert(answers, ans)

        local name
        name, pos = _decode_name(buf, pos)
        if not name then
            return nil, pos
        end

        ans.name = name

        -- print("name: ", name)

        type_hi = byte(buf, pos)
        type_lo = byte(buf, pos + 1)
        local typ = lshift(type_hi, 8) + type_lo

        ans.type = typ

        -- print("type: ", typ)

        class_hi = byte(buf, pos + 2)
        class_lo = byte(buf, pos + 3)
        local class = lshift(class_hi, 8) + class_lo

        ans.class = class

        -- print("class: ", class)

        local ttl_bytes = { byte(buf, pos + 4, pos + 7) }

        -- print("ttl bytes: ", concat(ttl_bytes, " "))

        local ttl = lshift(ttl_bytes[1], 24) + lshift(ttl_bytes[2], 16)
                    + lshift(ttl_bytes[3], 8) + ttl_bytes[4]

        -- print("ttl: ", ttl)

        ans.ttl = ttl

        local len_hi = byte(buf, pos + 8)
        local len_lo = byte(buf, pos + 9)
        local len = lshift(len_hi, 8) + len_lo

        -- print("record len: ", len)

        pos = pos + 10

        if typ == TYPE_A then

            if len ~= 4 then
                return nil, "bad A record value length: " .. len
            end

            local addr_bytes = { byte(buf, pos, pos + 3) }
            local addr = concat(addr_bytes, ".")
            -- print("ipv4 address: ", addr)

            ans.address = addr

            pos = pos + 4

        elseif typ == TYPE_CNAME then

            local cname, p = _decode_name(buf, pos)
            if not cname then
                return nil, pos
            end

            if p - pos ~= len then
                return nil, format("bad cname record length: %d ~= %d",
                                   p - pos, len)
            end

            pos = p

            -- print("cname: ", cname)

            ans.cname = cname

        elseif typ == TYPE_AAAA then

            if len ~= 16 then
                return nil, "bad AAAA record value length: " .. len
            end

            local addr_bytes = { byte(buf, pos, pos + 15) }
            local flds = {}
            local comp_begin, comp_end
            for i = 1, 16, 2 do
                local a = addr_bytes[i]
                local b = addr_bytes[i + 1]
                if a == 0 then
                    insert(flds, format("%x", b))

                else
                    insert(flds, format("%x%02x", a, b))
                end
            end

            -- we do not compress the IPv6 addresses by default
            --  due to performance considerations

            ans.address = concat(flds, ":")

            pos = pos + 16

        elseif typ == TYPE_MX then

            -- print("len = ", len)

            if len < 3 then
                return nil, "bad MX record value length: " .. len
            end

            local pref_hi = byte(buf, pos)
            local pref_lo = byte(buf, pos + 1)

            ans.preference = lshift(pref_hi, 8) + pref_lo

            local host, p = _decode_name(buf, pos + 2)
            if not host then
                return nil, pos
            end

            if p - pos ~= len then
                return nil, format("bad cname record length: %d ~= %d",
                                   p - pos, len)
            end

            ans.exchange = host

            pos = p

        elseif typ == TYPE_SRV then
            if len < 7 then
                return nil, "bad SRV record value length: " .. len
            end

            local prio_hi = byte(buf, pos)
            local prio_lo = byte(buf, pos + 1)
            ans.priority = lshift(prio_hi, 8) + prio_lo

            local weight_hi = byte(buf, pos + 2)
            local weight_lo = byte(buf, pos + 3)
            ans.weight = lshift(weight_hi, 8) + weight_lo

            local port_hi = byte(buf, pos + 4)
            local port_lo = byte(buf, pos + 5)
            ans.port = lshift(port_hi, 8) + port_lo

            local name, p = _decode_name(buf, pos + 6)
            if not name then
                return nil, pos
            end

            if p - pos ~= len then
                return nil, format("bad srv record length: %d ~= %d",
                                   p - pos, len)
            end

            ans.target = name

            pos = p

        elseif typ == TYPE_NS then

            local name, p = _decode_name(buf, pos)
            if not name then
                return nil, pos
            end

            if p - pos ~= len then
                return nil, format("bad cname record length: %d ~= %d",
                                   p - pos, len)
            end

            pos = p

            -- print("name: ", name)

            ans.nsdname = name

        elseif typ == TYPE_TXT or typ == TYPE_SPF then

            local key = (typ == TYPE_TXT) and "txt" or "spf"

            local slen = byte(buf, pos)
            if slen + 1 > len then
                -- truncate the over-run TXT record data
                slen = len
            end

            -- print("slen: ", len)

            local val = sub(buf, pos + 1, pos + slen)
            local last = pos + len
            pos = pos + slen + 1

            if pos < last then
                -- more strings to be processed
                -- this code path is usually cold, so we do not
                -- merge the following loop on this code path
                -- with the processing logic above.

                val = {val}
                local idx = 2
                repeat
                    local slen = byte(buf, pos)
                    if pos + slen + 1 > last then
                        -- truncate the over-run TXT record data
                        slen = last - pos - 1
                    end

                    val[idx] = sub(buf, pos + 1, pos + slen)
                    idx = idx + 1
                    pos = pos + slen + 1

                until pos >= last
            end

            ans[key] = val

        elseif typ == TYPE_PTR then

            local name, p = _decode_name(buf, pos)
            if not name then
                return nil, pos
            end

            if p - pos ~= len then
                return nil, format("bad cname record length: %d ~= %d",
                                   p - pos, len)
            end

            pos = p

            -- print("name: ", name)

            ans.ptrdname = name

        else
            -- for unknown types, just forward the raw value

            ans.rdata = sub(buf, pos, pos + len - 1)
            pos = pos + len
        end
    end

    return answers
end


local function _gen_id(self)
    local id = self._id   -- for regression testing
    if id then
        return id
    end
    return rand(0, 65535)   -- two bytes
end


local function _tcp_query(self, query, id)
    local sock = self.tcp_sock
    if not sock then
        return nil, "not initialized"
    end

    log(DEBUG, "query the TCP server due to reply truncation")

    local server = _get_cur_server(self)

    local ok, err = sock:connect(server[1], server[2])
    if not ok then
        return nil, "failed to connect to TCP server "
            .. concat(server, ":") .. ": " .. err
    end

    query = concat(query, "")
    local len = #query

    local len_hi = char(rshift(len, 8))
    local len_lo = char(band(len, 0xff))

    local bytes, err = sock:send({len_hi, len_lo, query})
    if not bytes then
        return nil, "failed to send query to TCP server "
            .. concat(server, ":") .. ": " .. err
    end

    local buf, err = sock:receive(2)
    if not buf then
        return nil, "failed to receive the reply length field from TCP server "
            .. concat(server, ":") .. ": " .. err
    end

    local len_hi = byte(buf, 1)
    local len_lo = byte(buf, 2)
    local len = lshift(len_hi, 8) + len_lo

    -- print("tcp message len: ", len)

    buf, err = sock:receive(len)
    if not buf then
        return nil, "failed to receive the reply message body from TCP server "
            .. concat(server, ":") .. ": " .. err
    end

    local answers, err = parse_response(buf, id)
    if not answers then
        return nil, "failed to parse the reply from the TCP server "
            .. concat(server, ":") .. ": " .. err
    end

    sock:close()

    return answers
end


function _M.tcp_query(self, qname, opts)
    local socks = self.socks
    if not socks then
        return nil, "not initialized"
    end

    pick_sock(self, socks)

    local id = _gen_id(self)

    local query, err = _build_request(qname, id, self.no_recurse, opts)
    if not query then
        return nil, err
    end

    return _tcp_query(self, query, id)
end


function _M.query(self, qname, opts)
    local socks = self.socks
    if not socks then
        return nil, "not initialized"
    end

    local id = _gen_id(self)

    local query, err = _build_request(qname, id, self.no_recurse, opts)
    if not query then
        return nil, err
    end

    -- local cjson = require "cjson"
    -- print("query: ", cjson.encode(concat(query, "")))

    local retrans = self.retrans

    -- print("retrans: ", retrans)

    for i = 1, retrans do
        local sock = pick_sock(self, socks)

        local ok, err = sock:send(query)
        if not ok then
            local server = _get_cur_server(self)
            return nil, "failed to send request to UDP server "
                .. concat(server, ":") .. ": " .. err
        end

        local buf, err

        for j = 1, 128 do
            buf, err = sock:receive(4096)

            if err then
                break
            end

            if buf then
                local answers
                answers, err = parse_response(buf, id)
                if not answers then
                    if err == "truncated" then
                        return _tcp_query(self, query, id)
                    end

                    if err ~= "id mismatch" then
                        return nil, err
                    end

                    -- retry receiving when err == "id mismatch"
                else
                    return answers
                end
            end
        end

        if err ~= "timeout" or i == retrans then
            local server = _get_cur_server(self)
            return nil, "failed to receive reply from UDP server "
                .. concat(server, ":") .. ": " .. err
        end
    end

    -- impossible to reach here
end


function _M.compress_ipv6_addr(addr)
    local addr = re_sub(addr, "^(0:)+|(:0)+$|:(0:)+", "::", "jo")
    if addr == "::0" then
        addr = "::"
    end

    return addr
end


randomseed(ngx_time())


return _M
