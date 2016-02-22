use Test::Nginx::Socket;
use Cwd qw(cwd);

plan tests => repeat_each() * 24;

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;;";
};

no_long_string();
run_tests();

__DATA__
=== TEST 1: Load module without errors.
--- http_config eval
"$::HttpConfig"
. q{
    init_by_lua '
        local DNS_Cache = require("resty.dns.cache")
    ';
}
--- config
    location /sanity {
        echo "OK";
    }
--- request
GET /sanity
--- no_error_log
[error]
--- response_body
OK


=== TEST 2: Can init cache - defaults
--- http_config eval
"$::HttpConfig"
. q{
    init_by_lua '
        local DNS_Cache = require("resty.dns.cache")
        DNS_Cache.init_cache()
    ';
}
--- config
    location /sanity {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            ngx.say(DNS_Cache.initted())
        ';
    }
--- request
GET /sanity
--- no_error_log
[error]
--- response_body
true

=== TEST 3: Can init cache - user config
--- http_config eval
"$::HttpConfig"
. q{
    init_by_lua '
        local DNS_Cache = require("resty.dns.cache")
        DNS_Cache.init_cache(300)
    ';
}
--- config
    location /sanity {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            ngx.say(DNS_Cache.initted())
        ';
    }
--- request
GET /sanity
--- no_error_log
[error]
--- response_body
true

=== TEST 4: Can init new instance - defaults
--- http_config eval
"$::HttpConfig"
. q{
    init_by_lua '
        local DNS_Cache = require("resty.dns.cache")
        DNS_Cache.init_cache(300)
    ';
}
--- config
    location /sanity {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new()
            if dns then
                ngx.say("OK")
            else
                ngx.say(err)
            end
        ';
    }
--- request
GET /sanity
--- no_error_log
[error]
--- response_body
OK

=== TEST 5: Can init new instance - user config
--- http_config eval
"$::HttpConfig"
. q{
    init_by_lua '
        local DNS_Cache = require("resty.dns.cache")
        DNS_Cache.init_cache(300)
    ';
}
--- config
    location /sanity {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                negative_ttl = 10,
                resolver = { nameservers = {"10.10.10.10"} }
            })
            if dns then
                ngx.say("OK")
            else
                ngx.say(err)
            end
        ';
    }
--- request
GET /sanity
--- no_error_log
[error]
--- response_body
OK

=== TEST 6: Resty DNS errors are passed through
--- http_config eval
"$::HttpConfig"
. q{
    init_by_lua '
        local DNS_Cache = require("resty.dns.cache")
        DNS_Cache.init_cache(300)
    ';
}
--- config
    location /sanity {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                resolver = {  }
            })
            if dns then
                ngx.say("OK")
            else
                ngx.say(err)
            end
        ';
    }
--- request
GET /sanity
--- no_error_log
[error]
--- response_body
no nameservers specified

=== TEST 7: Can create instance with shared dict
--- http_config eval
"$::HttpConfig"
. q{
    lua_shared_dict dns_cache 1m;
    init_by_lua '
        local DNS_Cache = require("resty.dns.cache")
        DNS_Cache.init_cache()
    ';
}
--- config
    location /sanity {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            ngx.say(DNS_Cache.initted())

            local dns, err = DNS_Cache.new({
                dict = "dns_cache"
            })
            if dns then
                ngx.say("OK")
            else
                ngx.say(err)
            end
        ';
    }
--- request
GET /sanity
--- no_error_log
[error]
--- response_body
true
OK

=== TEST 8: Can create instance with shared dict and no lru_cache
--- http_config eval
"$::HttpConfig"
. q{
    lua_shared_dict dns_cache 1m;
}
--- config
    location /sanity {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            ngx.say(DNS_Cache.initted())

            local dns, err = DNS_Cache.new({
                dict = "dns_cache"
            })
            if dns then
                ngx.say("OK")
            else
                ngx.say(err)
            end
        ';
    }
--- request
GET /sanity
--- no_error_log
[error]
--- response_body
false
OK
