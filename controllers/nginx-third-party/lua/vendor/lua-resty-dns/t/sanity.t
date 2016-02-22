# vim:set ft= ts=4 sw=4 et:

use Test::Nginx::Socket::Lua;
use Cwd qw(cwd);

repeat_each(2);

plan tests => repeat_each() * (3 * blocks());

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/t/lib/?.lua;$pwd/lib/?.lua;;";
    lua_package_cpath "/usr/local/openresty-debug/lualib/?.so;/usr/local/openresty/lualib/?.so;;";
};

$ENV{TEST_NGINX_RESOLVER} ||= '8.8.8.8';

no_long_string();

run_tests();

__DATA__

=== TEST 1: A records
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[.*?"address":"(?:\d{1,3}\.){3}\d+".*?\]$
--- no_error_log
[error]



=== TEST 2: CNAME records
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("www.yahoo.com", { qtype = r.TYPE_CNAME })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[.*?"cname":"[-_a-z0-9.]+".*?\]$
--- no_error_log
[error]



=== TEST 3: AAAA records
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"
            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_AAAA })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[.*?"address":"[a-fA-F0-9]*(?::[a-fA-F0-9]*)+".*?\]$
--- no_error_log
[error]



=== TEST 4: compress ipv6 addr
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local c = resolver.compress_ipv6_addr

            ngx.say(c("1080:0:0:0:8:800:200C:417A"))
            ngx.say(c("FF01:0:0:0:0:0:0:101"))
            ngx.say(c("0:0:0:0:0:0:0:1"))
            ngx.say(c("1:5:0:0:0:0:0:0"))
            ngx.say(c("7:25:0:0:0:3:0:0"))
            ngx.say(c("0:0:0:0:0:0:0:0"))
        ';
    }
--- request
GET /t
--- response_body
1080::8:800:200C:417A
FF01::101
::1
1:5::
7:25::3:0:0
::
--- no_error_log
[error]



=== TEST 5: A records (TCP)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
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
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[.*?"address":"(?:\d{1,3}\.){3}\d+".*?\]$
--- no_error_log
[error]



=== TEST 6: MX records
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("gmail.com", { qtype = r.TYPE_MX })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[\{.*?"preference":\d+,.*?"exchange":"[^"]+".*?\}\]$
--- no_error_log
[error]



=== TEST 7: NS records
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("agentzh.org", { qtype = r.TYPE_NS })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[\{.*?"nsdname":"[^"]+".*?\}\]$
--- no_error_log
[error]



=== TEST 8: TXT query (no ans)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("agentzh.org", { qtype = r.TYPE_TXT })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body
records: {}
--- no_error_log
[error]
--- timeout: 10



=== TEST 9: TXT query (with ans)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("gmail.com", { qtype = r.TYPE_TXT })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[\{.*?"txt":"v=spf\d+\s[^"]+".*?\}\]$
--- no_error_log
[error]



=== TEST 10: PTR query
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("4.4.8.8.in-addr.arpa", { qtype = r.TYPE_PTR })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[\{.*?"ptrdname":"google-public-dns-b\.google\.com".*?\}\]$
--- no_error_log
[error]



=== TEST 11: domains with a trailing dot
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("www.google.com.", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[.*?"address":"(?:\d{1,3}\.){3}\d+".*?\]$
--- no_error_log
[error]



=== TEST 12: domains with a leading dot
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query(".www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body
failed to query: bad name
--- no_error_log
[error]



=== TEST 13: SRV records or XMPP
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("_xmpp-client._tcp.jabber.org", { qtype = r.TYPE_SRV })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local ljson = require "ljson"
            ngx.say("records: ", ljson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[(?:{"class":1,"name":"_xmpp-client._tcp.jabber.org","port":\d+,"priority":\d+,"target":"[\w.]+\.jabber.org","ttl":\d+,"type":33,"weight":\d+},?)+\]$
--- no_error_log
[error]



=== TEST 14: SPF query (with ans)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("linkedin.com", { qtype = r.TYPE_SPF })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body_like chop
^records: \[\{.*?"spf":"v=spf\d+\s[^"]+".*?\}\]$
--- no_error_log
[error]



=== TEST 15: SPF query (no ans)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{ nameservers = { "$TEST_NGINX_RESOLVER" } }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            local ans, err = r:query("agentzh.org", { qtype = r.TYPE_SPF })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- request
GET /t
--- response_body
records: {}
--- no_error_log
[error]
--- timeout: 10

