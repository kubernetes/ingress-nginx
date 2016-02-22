Name
====

lua-resty-lock - Simple shm-based nonblocking lock API

Table of Contents
=================

* [Name](#name)
* [Status](#status)
* [Synopsis](#synopsis)
* [Description](#description)
* [Methods](#methods)
    * [new](#new)
    * [lock](#lock)
    * [unlock](#unlock)
* [For Multiple Lua Light Threads](#for-multiple-lua-light-threads)
* [For Cache Locks](#for-cache-locks)
* [Prerequisites](#prerequisites)
* [Installation](#installation)
* [TODO](#todo)
* [Community](#community)
    * [English Mailing List](#english-mailing-list)
    * [Chinese Mailing List](#chinese-mailing-list)
* [Bugs and Patches](#bugs-and-patches)
* [Author](#author)
* [Copyright and License](#copyright-and-license)
* [See Also](#see-also)

Status
======

This library is still under early development and is production ready.

Synopsis
========

```lua
# nginx.conf

http {
    # you do not need the following line if you are using the
    #   ngx_openresty bundle:
    lua_package_path "/path/to/lua-resty-lock/lib/?.lua;;";

    lua_shared_dict my_locks 100k;

    server {
        ...

        location = /t {
            content_by_lua '
                local lock = require "resty.lock"
                for i = 1, 2 do
                    local lock = lock:new("my_locks")

                    local elapsed, err = lock:lock("my_key")
                    ngx.say("lock: ", elapsed, ", ", err)

                    local ok, err = lock:unlock()
                    if not ok then
                        ngx.say("failed to unlock: ", err)
                    end
                    ngx.say("unlock: ", ok)
                end
            ';
        }
    }
}
```

Description
===========

This library implements a simple mutex lock in a similar way to ngx_proxy module's [proxy_cache_lock directive](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_lock).

Under the hood, this library uses [ngx_lua](https://github.com/chaoslawful/lua-nginx-module) module's shared memory dictionaries. The lock waiting is nonblocking because we use stepwise [ngx.sleep](https://github.com/chaoslawful/lua-nginx-module#ngxsleep) to poll the lock periodically.

[Back to TOC](#table-of-contents)

Methods
=======

To load this library,

1. you need to specify this library's path in ngx_lua's [lua_package_path](https://github.com/chaoslawful/lua-nginx-module#lua_package_path) directive. For example, `lua_package_path "/path/to/lua-resty-lock/lib/?.lua;;";`.
2. you use `require` to load the library into a local Lua variable:

```lua
    local lock = require "resty.lock"
```

[Back to TOC](#table-of-contents)

new
---
`syntax: obj = lock:new(dict_name)`

`syntax: obj = lock:new(dict_name, opts)`

Creates a new lock object instance by specifying the shared dictionary name (created by [lua_shared_dict](http://https://github.com/chaoslawful/lua-nginx-module#lua_shared_dict)) and an optional options table `opts`.

The options table accepts the following options:

* `exptime`
Specifies expiration time (in seconds) for the lock entry in the shared memory dictionary. You can specify up to `0.001` seconds. Default to 30 (seconds). Even if the invoker does not call `unlock` or the object holding the lock is not GC'd, the lock will be released after this time. So deadlock won't happen even when the worker process holding the lock crashes.
* `timeout`
Specifies the maximal waiting time (in seconds) for the [lock](#lock) method calls on the current object instance. You can specify up to `0.001` seconds. Default to 5 (seconds). This option value cannot be bigger than `exptime`. This timeout is to prevent a [lock](#lock) method call from waiting forever.
You can specify `0` to make the [lock](#lock) method return immediately without waiting if it cannot acquire the lock right away.
* `step`
Specifies the initial step (in seconds) of sleeping when waiting for the lock. Default to `0.001` (seconds). When the [lock](#lock) method is waiting on a busy lock, it sleeps by steps. The step size is increased by a ratio (specified by the `ratio` option) until reaching the step size limit (specified by the `max_step` option).
* `ratio`
Specifies the step increasing ratio. Default to 2, that is, the step size doubles at each waiting iteration.
* `max_step`
Specifies the maximal step size (i.e., sleep interval, in seconds) allowed. See also the `step` and `ratio` options). Default to 0.5 (seconds).

[Back to TOC](#table-of-contents)

lock
----
`syntax: elapsed, err = obj:lock(key)`

Tries to lock a key across all the Nginx worker processes in the current Nginx server instance. Different keys are different locks.

The length of the key string must not be larger than 65535 bytes.

Returns the waiting time (in seconds) if the lock is successfully acquired. Otherwise returns `nil` and a string describing the error.

The waiting time is not from the wallclock, but rather is from simply adding up all the waiting "steps". A nonzero `elapsed` return value indicates that someone else has just hold this lock. But a zero return value cannot gurantee that no one else has just acquired and released the lock.

When this method is waiting on fetching the lock, no operating system threads will be blocked and the current Lua "light thread" will be automatically yielded behind the scene.

It is strongly recommended to always call the [unlock()](#unlock) method to actively release the lock as soon as possible.

If the [unlock()](#unlock) method is never called after this method call, the lock will get released when
1. the current `resty.lock` object instance is collected automatically by the Lua GC.
2. the `exptime` for the lock entry is reached.

Common errors for this method call is
* "timeout"
: The timeout threshold specified by the `timeout` option of the [new](#new) method is exceeded.
* "locked"
: The current `resty.lock` object instance is already holding a lock (not necessarily of the same key).

Other possible errors are from ngx_lua's shared dictionary API.

[Back to TOC](#table-of-contents)

unlock
------
`syntax: ok, err = obj:unlock()`

Releases the lock held by the current `resty.lock` object instance.

Returns `1` on success. Returns `nil` and a string describing the error otherwise.

If you call `unlock` when no lock is currently held, the error "unlocked" will be returned.

[Back to TOC](#table-of-contents)

For Multiple Lua Light Threads
==============================

It is always a bad idea to share a single `resty.lock` object instance across multiple ngx_lua "light threads" because the object itself is stateful and is vulnerable to race conditions. It is highly recommended to always allocate a separate `resty.lock` object instance for each "light thread" that needs one.

[Back to TOC](#table-of-contents)

For Cache Locks
===============

One common use case for this library is avoid the so-called "dog-pile effect", that is, to limit concurrent backend queries for the same key when a cache miss happens. This usage is similar to the standard ngx_proxy module's [proxy_cache_lock](http://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_cache_lock) directive.

The basic workflow for a cache lock is as follows:

1. Check the cache for a hit with the key. If a cache miss happens, proceed to step 2.
2. Instantiate a `resty.lock` object, call the [lock](#lock) method on the key, and check the 1st return value, i.e., the lock waiting time. If it is `nil`, handle the error; otherwise proceed to step 3.
3. Check the cache again for a hit. If it is still a miss, proceed to step 4; otherwise release the lock by calling [unlock](#unlock) and then return the cached value.
4. Query the backend (the data source) for the value, put the result into the cache, and then release the lock currently held by calling [unlock](#unlock).

Below is a kinda complete code example that demonstrates the idea.

```lua
    local resty_lock = require "resty.lock"
    local cache = ngx.shared.my_cache

    -- step 1:
    local val, err = cache:get(key)
    if val then
        ngx.say("result: ", val)
        return
    end

    if err then
        return fail("failed to get key from shm: ", err)
    end

    -- cache miss!
    -- step 2:
    local lock = resty_lock:new("my_locks")
    local elapsed, err = lock:lock(key)
    if not elapsed then
        return fail("failed to acquire the lock: ", err)
    end

    -- lock successfully acquired!

    -- step 3:
    -- someone might have already put the value into the cache
    -- so we check it here again:
    val, err = cache:get(key)
    if val then
        local ok, err = lock:unlock()
        if not ok then
            return fail("failed to unlock: ", err)
        end

        ngx.say("result: ", val)
        return
    end

    --- step 4:
    local val = fetch_redis(key)
    if not val then
        local ok, err = lock:unlock()
        if not ok then
            return fail("failed to unlock: ", err)
        end

        -- FIXME: we should handle the backend miss more carefully
        -- here, like inserting a stub value into the cache.

        ngx.say("no value found")
        return
    end

    -- update the shm cache with the newly fetched value
    local ok, err = cache:set(key, val, 1)
    if not ok then
        local ok, err = lock:unlock()
        if not ok then
            return fail("failed to unlock: ", err)
        end

        return fail("failed to update shm cache: ", err)
    end

    local ok, err = lock:unlock()
    if not ok then
        return fail("failed to unlock: ", err)
    end

    ngx.say("result: ", val)
```

Here we assume that we use the ngx_lua shared memory dictionary to cache the Redis query results and we have the following configurations in `nginx.conf`:

```nginx
    # you may want to change the dictionary size for your cases.
    lua_shared_dict my_cache 10m;
    lua_shared_dict my_locks 1m;
```

The `my_cache` dictionary is for the data cache while the `my_locks` dictionary is for `resty.lock` itself.

Several important things to note in the example above:

1. You need to release the lock as soon as possible, even when some other unrelated errors happen.
2. You need to update the cache with the result got from the backend *before* releasing the lock so other threads already waiting on the lock can get cached value when they get the lock afterwards.
3. When the backend returns no value at all, we should handle the case carefully by inserting some stub value into the cache.

[Back to TOC](#table-of-contents)

Prerequisites
=============

* [LuaJIT](http://luajit.org) 2.0+
* [ngx_lua](https://github.com/chaoslawful/lua-nginx-module) 0.8.10+

[Back to TOC](#table-of-contents)

Installation
============

It is recommended to use the latest [ngx_openresty bundle](http://openresty.org) directly where this library
is bundled and enabled by default. At least ngx_openresty 1.4.2.9 is required. And you need to enable LuaJIT when building your ngx_openresty
bundle by passing the `--with-luajit` option to its `./configure` script. No extra Nginx configuration is required.

If you want to use this library with your own Nginx build (with ngx_lua), then you need to
ensure you are using at least ngx_lua 0.8.10. Also, You need to configure
the [lua_package_path](https://github.com/chaoslawful/lua-nginx-module#lua_package_path) directive to
add the path of your lua-resty-lock source tree to ngx_lua's Lua module search path, as in

```nginx
    # nginx.conf
    http {
        lua_package_path "/path/to/lua-resty-lock/lib/?.lua;;";
        ...
    }
```

and then load the library in Lua:

```lua
    local lock = require "resty.lock"
```

[Back to TOC](#table-of-contents)

TODO
====

* We should simplify the current implementation when LuaJIT 2.1 gets support for `__gc` metamethod on normal Lua tables. Right now we are using an FFI cdata and a ref/unref memo table to work around this, which is rather ugly and a bit inefficient.

[Back to TOC](#table-of-contents)

Community
=========

[Back to TOC](#table-of-contents)

English Mailing List
--------------------

The [openresty-en](https://groups.google.com/group/openresty-en) mailing list is for English speakers.

[Back to TOC](#table-of-contents)

Chinese Mailing List
--------------------

The [openresty](https://groups.google.com/group/openresty) mailing list is for Chinese speakers.

[Back to TOC](#table-of-contents)

Bugs and Patches
================

Please report bugs or submit patches by

1. creating a ticket on the [GitHub Issue Tracker](http://github.com/openresty/lua-resty-lock/issues),
1. or posting to the [OpenResty community](#community).

[Back to TOC](#table-of-contents)

Author
======

Yichun "agentzh" Zhang (章亦春) <agentzh@gmail.com>, CloudFlare Inc.

[Back to TOC](#table-of-contents)

Copyright and License
=====================

This module is licensed under the BSD license.

Copyright (C) 2013-2014, by Yichun "agentzh" Zhang, CloudFlare Inc.

All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

[Back to TOC](#table-of-contents)

See Also
========
* the ngx_lua module: https://github.com/chaoslawful/lua-nginx-module
* OpenResty: http://openresty.org
[Back to TOC](#table-of-contents)

