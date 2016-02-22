Name
====

lua-resty-dns - Lua DNS resolver for the ngx_lua based on the cosocket API

Table of Contents
=================

* [Name](#name)
* [Status](#status)
* [Description](#description)
* [Synopsis](#synopsis)
* [Methods](#methods)
    * [new](#new)
    * [query](#query)
    * [tcp_query](#tcp_query)
    * [set_timeout](#set_timeout)
    * [compress_ipv6_addr](#compress_ipv6_addr)
* [Constants](#constants)
    * [TYPE_A](#type_a)
    * [TYPE_NS](#type_ns)
    * [TYPE_CNAME](#type_cname)
    * [TYPE_PTR](#type_ptr)
    * [TYPE_MX](#type_mx)
    * [TYPE_TXT](#type_txt)
    * [TYPE_AAAA](#type_aaaa)
    * [TYPE_SRV](#type_srv)
    * [TYPE_SPF](#type_spf)
    * [CLASS_IN](#class_in)
* [Automatic Error Logging](#automatic-error-logging)
* [Limitations](#limitations)
* [TODO](#todo)
* [Author](#author)
* [Copyright and License](#copyright-and-license)
* [See Also](#see-also)

Status
======

This library is considered production ready.

Description
===========

This Lua library provies a DNS resolver for the ngx_lua nginx module:

http://wiki.nginx.org/HttpLuaModule

This Lua library takes advantage of ngx_lua's cosocket API, which ensures
100% nonblocking behavior.

Note that at least [ngx_lua 0.5.12](https://github.com/chaoslawful/lua-nginx-module/tags) or [ngx_openresty 1.2.1.11](http://openresty.org/#Download) is required.

Also, the [bit library](http://bitop.luajit.org/) is also required. If you're using LuaJIT 2.0 with ngx_lua, then the `bit` library is already available by default.

Note that, this library is bundled and enabled by default in the [ngx_openresty bundle](http://openresty.org/).

Synopsis
========

```lua
    lua_package_path "/path/to/lua-resty-dns/lib/?.lua;;";

    server {
        location = /dns {
            content_by_lua '
                local resolver = require "resty.dns.resolver"
                local r, err = resolver:new{
                    nameservers = {"8.8.8.8", {"8.8.4.4", 53} },
                    retrans = 5,  -- 5 retransmissions on receive timeout
                    timeout = 2000,  -- 2 sec
                }

                if not r then
                    ngx.say("failed to instantiate the resolver: ", err)
                    return
                end

                local answers, err = r:query("www.google.com")
                if not answers then
                    ngx.say("failed to query the DNS server: ", err)
                    return
                end

                if answers.errcode then
                    ngx.say("server returned error code: ", answers.errcode,
                            ": ", answers.errstr)
                end

                for i, ans in ipairs(answers) do
                    ngx.say(ans.name, " ", ans.address or ans.cname,
                            " type:", ans.type, " class:", ans.class,
                            " ttl:", ans.ttl)
                end
            ';
        }
    }
```

[Back to TOC](#table-of-contents)

Methods
=======

[Back to TOC](#table-of-contents)

new
---
`syntax: r, err = class:new(opts)`

Creates a dns.resolver object. Returns `nil` and an message string on error.

It accepts a `opts` table argument. The following options are supported:

* `nameservers`

	a list of nameservers to be used. Each nameserver entry can be either a single hostname string or a table holding both the hostname string and the port number. The nameserver is picked up by a simple round-robin algorithm for each `query` method call. This option is required.
* `retrans`

	the total number of times of retransmitting the DNS request when receiving a DNS response times out according to the `timeout` setting. Default to `5` times. When trying to retransmit the query, the next nameserver according to the round-robin algorithm will be picked up.
* `timeout`

	the time in milliseconds for waiting for the respond for a single attempt of request transmition. note that this is ''not'' the maximal total waiting time before giving up, the maximal total waiting time can be calculated by the expression `timeout x retrans`. The `timeout` setting can also be changed by calling the `set_timeout` method. The default `timeout` setting is 2000 milliseconds, or 2 seconds.
* `no_recurse`

	a boolean flag controls whether to disable the "recursion desired" (RD) flag in the UDP request. Default to `false`.

[Back to TOC](#table-of-contents)

query
-----
`syntax: answers, err = r:query(name, options?)`

Performs a DNS standard query to the nameservers specified by the `new` method,
and returns all the answer records in an array-like Lua table. In case of errors, it will
return `nil` and a string describing the error instead.

If the server returns a non-zero error code, the fields `errcode` and `errstr` will be set accordingly in the Lua table returned.

Each entry in the `answers` returned table value is also a hash-like Lua table
which usually takes some of the following fields:

* `name`

	The resource record name.
* `type`

	The current resource record type, possible values are `1` (`TYPE_A`), `5` (`TYPE_CNAME`), `28` (`TYPE_AAAA`), and any other values allowed by RFC 1035.
* `address`

	The IPv4 or IPv6 address in their textual representations when the resource record type is either `1` (`TYPE_A`) or `28` (`TYPE_AAAA`), respectively. Secussesive 16-bit zero groups in IPv6 addresses will not be compressed by default, if you want that, you need to call the `compress_ipv6_addr` static method instead.
* `cname`

	The (decoded) record data value for `CNAME` resource records. Only present for `CNAME` records.
* `ttl`

	The time-to-live (TTL) value in seconds for the current resource record.
* `class`

	The current resource record class, possible values are `1` (`CLASS_IN`) or any other values allowed by RFC 1035.
* `preference`

	The preference integer number for `MX` resource records. Only present for `MX` type records.
* `exchange`

	The exchange domain name for `MX` resource records. Only present for `MX` type records.
* `nsdname`

	A domain-name which specifies a host which should be authoritative for the specified class and domain. Usually present for `NS` type records.
* `rdata`

	The raw resource data (RDATA) for resource records that are not recognized.
* `txt`

	The record value for `TXT` records. When there is only one character string in this record, then this field takes a single Lua string. Otherwise this field takes a Lua table holding all the strings.
* `ptrdname`

	The record value for `PTR` records.

This method also takes an optional `options` argument table, which takes the following fields:

* `qtype`

	The type of the question. Possible values are `1` (`TYPE_A`), `5` (`TYPE_CNAME`), `28` (`TYPE_AAAA`), or any other QTYPE value specified by RFC 1035 and RFC 3596. Default to `1` (`TYPE_A`).

When data truncation happens, the resolver will automatically retry using the TCP transport mode
to query the current nameserver. All TCP connections are short lived.

[Back to TOC](#table-of-contents)

tcp_query
---------
`syntax: answers, err = r:tcp_query(name, options?)`

Just like the `query` method, but enforce the TCP transport mode instead of UDP.

All TCP connections are short lived.

Here is an example:

```lua
    local resolver = require "resty.dns.resolver"

    local r, err = resolver:new{
        nameservers = { "8.8.8.8" }
    }
    if not r then
        ngx.say("failed to instantiate resolver: ", err)
        return
    end

    local ans, err = r:tcp_query("www.google.com", { qtype = r.TYPE_A })
    if not ans then
        ngx.say("failed to query: ", err)
        return
    end

    local cjson = require "cjson"
    ngx.say("records: ", cjson.encode(ans))
```

[Back to TOC](#table-of-contents)

set_timeout
-----------
`syntax: r:set_timeout(time)`

Overrides the current `timeout` setting by the `time` argument in milliseconds for all the nameserver peers.

[Back to TOC](#table-of-contents)

compress_ipv6_addr
------------------
`syntax: compressed = resty.dns.resolver.compress_ipv6_addr(address)`

Compresses the successive 16-bit zero groups in the textual format of the IPv6 address.

For example,

```lua
    local resolver = require "resty.dns.resolver"
    local compress = resolver.compress_ipv6_addr
    local new_addr = compress("FF01:0:0:0:0:0:0:101")
```

will yield `FF01::101` in the `new_addr` return value.

[Back to TOC](#table-of-contents)

Constants
=========

[Back to TOC](#table-of-contents)

TYPE_A
------

The `A` resource record type, equal to the decimal number `1`.

[Back to TOC](#table-of-contents)

TYPE_NS
-------

The `NS` resource record type, equal to the decimal number `2`.

[Back to TOC](#table-of-contents)

TYPE_CNAME
----------

The `CNAME` resource record type, equal to the decimal number `5`.

[Back to TOC](#table-of-contents)

TYPE_PTR
--------

The `PTR` resource record type, equal to the decimal number `12`.

[Back to TOC](#table-of-contents)

TYPE_MX
-------

The `MX` resource record type, equal to the decimal number `15`.

[Back to TOC](#table-of-contents)

TYPE_TXT
--------

The `TXT` resource record type, equal to the decimal number `16`.

[Back to TOC](#table-of-contents)

TYPE_AAAA
---------
`syntax: typ = r.TYPE_AAAA`

The `AAAA` resource record type, equal to the decimal number `28`.

[Back to TOC](#table-of-contents)

TYPE_SRV
---------
`syntax: typ = r.TYPE_SRV`

The `SRV` resource record type, equal to the decimal number `33`.

See RFC 2782 for details.

[Back to TOC](#table-of-contents)

TYPE_SPF
---------
`syntax: typ = r.TYPE_SPF`

The `SPF` resource record type, equal to the decimal number `99`.

See RFC 4408 for details.

[Back to TOC](#table-of-contents)

CLASS_IN
--------
`syntax: class = r.CLASS_IN`

The `Internet` resource record type, equal to the decimal number `1`.

[Back to TOC](#table-of-contents)

Automatic Error Logging
=======================

By default the underlying [ngx_lua](http://wiki.nginx.org/HttpLuaModule) module
does error logging when socket errors happen. If you are already doing proper error
handling in your own Lua code, then you are recommended to disable this automatic error logging by turning off [ngx_lua](http://wiki.nginx.org/HttpLuaModule)'s [lua_socket_log_errors](http://wiki.nginx.org/HttpLuaModule#lua_socket_log_errors) directive, that is,

```nginx
    lua_socket_log_errors off;
```

[Back to TOC](#table-of-contents)

Limitations
===========

* This library cannot be used in code contexts like set_by_lua*, log_by_lua*, and
header_filter_by_lua* where the ngx_lua cosocket API is not available.
* The `resty.dns.resolver` object instance cannot be stored in a Lua variable at the Lua module level,
because it will then be shared by all the concurrent requests handled by the same nginx
 worker process (see
http://wiki.nginx.org/HttpLuaModule#Data_Sharing_within_an_Nginx_Worker ) and
result in bad race conditions when concurrent requests are trying to use the same `resty.dns.resolver` instance.
You should always initiate `resty.dns.resolver` objects in function local
variables or in the `ngx.ctx` table. These places all have their own data copies for
each request.

[Back to TOC](#table-of-contents)

TODO
====

* Concurrent (or parallel) query mode
* Better support for other resource record types like `TLSA`.

[Back to TOC](#table-of-contents)

Author
======

Yichun "agentzh" Zhang (章亦春) <agentzh@gmail.com>, CloudFlare Inc.

[Back to TOC](#table-of-contents)

Copyright and License
=====================

This module is licensed under the BSD license.

Copyright (C) 2012-2014, by Yichun "agentzh" Zhang (章亦春) <agentzh@gmail.com>, CloudFlare Inc.

All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

[Back to TOC](#table-of-contents)

See Also
========
* the ngx_lua module: http://wiki.nginx.org/HttpLuaModule
* the [lua-resty-memcached](https://github.com/agentzh/lua-resty-memcached) library.
* the [lua-resty-redis](https://github.com/agentzh/lua-resty-redis) library.
* the [lua-resty-mysql](https://github.com/agentzh/lua-resty-mysql) library.

[Back to TOC](#table-of-contents)

