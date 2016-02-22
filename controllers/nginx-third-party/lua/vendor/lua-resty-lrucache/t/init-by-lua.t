# vim:set ft= ts=4 sw=4 et fdm=marker:

use Test::Nginx::Socket::Lua;
use Cwd qw(cwd);

repeat_each(1);

plan tests => repeat_each() * 13;

#no_diff();
#no_long_string();

my $pwd = cwd();

our $HttpConfig = <<"_EOC_";
    lua_package_path "$pwd/lib/?.lua;$pwd/../lua-resty-core/lib/?.lua;;";
    #init_by_lua '
    #local v = require "jit.v"
    #v.on("$Test::Nginx::Util::ErrLogFile")
    #require "resty.core"
    #';

_EOC_

no_long_string();
run_tests();

__DATA__

=== TEST 1: sanity
--- http_config eval
"$::HttpConfig"
. qq!
        init_by_lua '
            local function log(...)
                ngx.log(ngx.WARN, ...)
            end

            local lrucache = require "resty.lrucache"
            local c = lrucache.new(2)

            collectgarbage()

            c:set("dog", 32)
            c:set("cat", 56)
            log("dog: ", c:get("dog"))
            log("cat: ", c:get("cat"))

            c:set("dog", 32)
            c:set("cat", 56)
            log("dog: ", c:get("dog"))
            log("cat: ", c:get("cat"))

            c:delete("dog")
            c:delete("cat")
            log("dog: ", c:get("dog"))
            log("cat: ", c:get("cat"))
        ';
!

--- config
    location = /t {
        echo ok;
    }
--- request
    GET /t
--- response_body
ok
--- no_error_log
[error]
--- error_log
dog: 32
cat: 56
dog: 32
cat: 56
dog: nil
cat: nil



=== TEST 2: sanity
--- http_config eval
"$::HttpConfig"
. qq!
init_by_lua '
    lrucache = require "resty.lrucache"
    flv_index, err = lrucache.new(200)
    if not flv_index then
        ngx.log(ngx.ERR, "failed to create the cache: ", err)
        return
    end

    flv_meta, err = lrucache.new(200)
    if not flv_meta then
        ngx.log(ngx.ERR, "failed to create the cache: ", err)
        return
    end

    flv_channel, err = lrucache.new(200)
    if not flv_channel then
        ngx.log(ngx.ERR, "failed to create the cache: ", err)
        return
    end

    ngx.log(ngx.WARN, "3 lrucache initialized.")
';
!

--- config
    location = /t {
        echo ok;
    }
--- request
    GET /t
--- response_body
ok
--- no_error_log
[error]
--- error_log
3 lrucache initialized.

