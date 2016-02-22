#lua-resty-dns-cache

A wrapper for [lua-resty-dns](https://github.com/openresty/lua-resty-dns) to cache responses based on record TTLs.

Uses [lua-resty-lrucache](https://github.com/openresty/lua-resty-lrucache) and [ngx.shared.DICT](https://github.com/openresty/lua-nginx-module#ngxshareddict) to provide a 2 level cache.

Can repopulate cache in the background while returning stale answers.

#Overview

```lua
  lua_shared_dict dns_cache 1m;

  init_by_lua '
      require("resty.dns.cache").init_cache(200)
  ';

    server {

        listen 80;
        server_name dns_cache;

        location / {
            content_by_lua '
                local DNS_Cache = require("resty.dns.cache")
                local dns = DNS_Cache.new({
                        dict = "dns_cache",
                        negative_ttl = 30,
                        max_stale = 300,
                        resolver  = {
                            nameservers = {"123.123.123.123"}
                        }
                    })

                local host = ngx.req.get_uri_args()["host"] or "www.google.com"

                local answer, err, stale = dns:query(host)
                if err then
                    if stale then
                        ngx.header["Warning"] = "110: Response is stale"
                        answer = stale
                        ngx.log(ngx.ERR, err)
                    else
                        ngx.status = 500
                        ngx.say(err)
                        return ngx.exit(ngx.status)
                    end
                end

                local cjson = require "cjson"
                ngx.say(cjson.encode(answer))
            ';
        }
    }
```

#Methods
### init_cache
`syntax: ok, err = dns_cache.init_cache(max_items?)`

Creates a global lrucache object for caching responses.

Accepts an optional `max_items` argument, defaults to 200 entries.

Calling this repeatedly will reset the LRU cache

### initted
`syntax: ok = dns_cache.initted()`

Returns `true` if LRU Cache has been initialised

### new
`syntax: ok, err = dns_cache.new(opts)`

Returns a new DNS cache instance. Returns `nil` and a string on error

Accepts a table of options, if no shared dictionary is provided only lrucache is used.

* `dict` - Name of the [ngx.shared.DICT](https://github.com/openresty/lua-nginx-module#ngxshareddict) to use for cache.
* `resolver` - Table of options passed to [lua-resty-dns](https://github.com/openresty/lua-resty-dns#new). Defaults to using Google DNS.
* `normalise_ttl` - Boolean. Reduces TTL in cached answers to account for cached time. Defaults to `true`.
* `negative_ttl` - Time in seconds to cache negative / error responses. `nil` or `false` disables caching negative responses. Defaults to `false`
* `minimise_ttl` - Boolean. Set cache TTL based on the shortest DNS TTL in all responses rather than the first response. Defaults to `false`
* `max_stale` - Number of seconds past expiry to return stale content rather than querying. Stale hits will trigger a non-blocking background query to repopulate cache.


### query
`syntax: answers, err, stale = c:query(name, opts?)`

Passes through to lua-resty-dns' [query](https://github.com/openresty/lua-resty-dns#query) method.

Returns an extra `stale` variable containing stale data if a resolver cannot be contacted.

### tcp_query
`syntax: answers, err, stale = c:tcp_query(name, opts?)`

Passes through to lua-resty-dns' [tcp_query](https://github.com/openresty/lua-resty-dns#tcp_query) method.

Returns an extra `stale` variable containing stale data if a resolver cannot be contacted.

### set_timeout
`syntax: c:set_timeout(time)`

Passes through to lua-resty-dns' [set_timeout](https://github.com/openresty/lua-resty-dns#set_timeout) method.

## Constants
lua-resty-dns' [constants](https://github.com/openresty/lua-resty-dns#constants) are accessible on the `resty.dns.cache` object too.

## TODO
 * Cap'n'proto serialisation

