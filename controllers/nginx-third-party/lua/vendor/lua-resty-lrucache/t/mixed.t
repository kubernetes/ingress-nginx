# vim:set ft= ts=4 sw=4 et fdm=marker:

use Test::Nginx::Socket::Lua;
use Cwd qw(cwd);

repeat_each(2);

plan tests => repeat_each() * (blocks() * 3);

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
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lrucache = require "resty.lrucache"
            local c = lrucache.new(2)

            collectgarbage()

            c:set("dog", 32)
            c:set("cat", 56)
            ngx.say("dog: ", c:get("dog"))
            ngx.say("cat: ", c:get("cat"))

            local lrucache = require "resty.lrucache.pureffi"
            local c2 = lrucache.new(2)

            ngx.say("dog: ", c2:get("dog"))
            ngx.say("cat: ", c2:get("cat"))

            c2:set("dog", 9)
            c2:set("cat", "hi")

            ngx.say("dog: ", c2:get("dog"))
            ngx.say("cat: ", c2:get("cat"))

            ngx.say("dog: ", c:get("dog"))
            ngx.say("cat: ", c:get("cat"))
        ';
    }
--- request
    GET /t
--- response_body
dog: 32
cat: 56
dog: nil
cat: nil
dog: 9
cat: hi
dog: 32
cat: 56

--- no_error_log
[error]

