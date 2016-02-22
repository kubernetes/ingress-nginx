use lib 't';
use TestDNS;
use Cwd qw(cwd);

plan tests => repeat_each() * 47;

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;;";
    lua_socket_log_errors off;
};

no_long_string();
run_tests();

__DATA__
=== TEST 1: Response comes from cache on second hit
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
    location /t {
        echo_location /_t;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
}
--- request
GET /t
--- error_log
lru_cache HIT
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}]

=== TEST 2: Response comes from dict on miss
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
    location /t {
        echo_location /_t;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            DNS_Cache.init_cache() -- reset cache
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
}
--- request
GET /t
--- error_log
lru_cache MISS
shared_dict HIT
lru_cache SET
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}]


=== TEST 3: Stale response from lru served if resolver down
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
    location /t {
        echo_location /_t;
        echo_sleep 2;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1954"}}, retrans = 1, timeout = 100}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if stale then
                answer = stale
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 1 }],
}
--- request
GET /t
--- error_log
lru_cache MISS
lru_cache STALE
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":-1}]


=== TEST 4: Stale response from dict served if resolver down
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
    location /t {
        echo_location /_t;
        echo_sleep 2;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1954"}}, retrans = 1, timeout = 100}
            })
            DNS_Cache.init_cache() -- reset cache
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if stale then
                answer = stale
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 1 }],
}
--- request
GET /t
--- error_log
lru_cache MISS
shared_dict STALE
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":-1}]

=== TEST 5: Stale response from lru served if resolver down, no dict
--- http_config eval
"$::HttpConfig"
. q{
    init_by_lua '
        local DNS_Cache = require("resty.dns.cache")
        DNS_Cache.init_cache()
    ';
}
--- config
    location /t {
        echo_location /_t;
        echo_sleep 2;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1954"}}, retrans = 1, timeout = 100}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if stale then
                answer = stale
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 1 }],
}
--- request
GET /t
--- error_log
lru_cache MISS
lru_cache STALE
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":-1}]

=== TEST 6: Stale response from dict served if resolver down, no lru
--- http_config eval
"$::HttpConfig"
. q{
    lua_shared_dict dns_cache 1m;
}
--- config
    location /t {
        echo_location /_t;
        echo_sleep 2;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1954"}}, retrans = 1, timeout = 100}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if stale then
                answer = stale
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 1 }],
}
--- request
GET /t
--- error_log
shared_dict STALE
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":-1}]

=== TEST 7: TTLs are reduced
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
    location /t {
        echo_location /_t;
        echo_sleep 2;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}, retrans = 1, timeout = 100}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(answer)
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 10 }],
}
--- request
GET /t
--- no_error_log
[error]
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":8}]

=== TEST 8: TTL reduction can be disabled
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
    location /t {
        echo_location /_t;
        echo_sleep 2;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                normalise_ttl = false,
                resolver = {nameservers = {{"127.0.0.1", "1953"}}, retrans = 1, timeout = 100}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(answer)
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 10 }],
}
--- request
GET /t
--- no_error_log
[error]
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":10}]

=== TEST 9: Negative responses are not cached by default
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
    location /t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns._debug(true)
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    rcode => 3,
    opcode => 0,
    qname => 'www.google.com',
}
--- request
GET /t
--- no_error_log
SET
--- response_body
{"errcode":3,"errstr":"name error"}


=== TEST 10: Negative responses can be cached
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
    location /t {
        echo_location /_t;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                negative_ttl = 10,
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                negative_ttl = 10,
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    rcode => 3,
    opcode => 0,
    qname => 'www.google.com',
}
--- request
GET /t
--- error_log
lru_cache HIT
--- response_body
{"errcode":3,"errstr":"name error"}

=== TEST 11: Cached negative responses are not returned by default
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
    location /t {
        echo_location /_t;
        echo_location /_t2;
    }
    location /_t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                negative_ttl = 10,
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns._debug(true)
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
        ';
    }
    location /_t2 {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1954"}, retrans = 1, timeout = 100}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    rcode => 3,
    opcode => 0,
    qname => 'www.google.com',
}
--- request
GET /t
--- error_log
lru_cache SET
lru_cache HIT
--- response_body
null

=== TEST 12: Cache TTL can be minimised
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
    location /t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                minimise_ttl = true,
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 },
        { name => "l.www.google.com", ipv6 => "::1", ttl => 10 },
    ],
}
--- request
GET /t
--- error_log
lru_cache SET: www.google.com|1 10
shared_dict SET: www.google.com|1 10
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456},{"address":"0:0:0:0:0:0:0:1","type":28,"class":1,"name":"l.www.google.com","ttl":10}]

=== TEST 13: Cache TTLs not minimised by default
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
    location /t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {nameservers = {{"127.0.0.1", "1953"}}}
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125
            dns._debug(true)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 },
        { name => "l.www.google.com", ipv6 => "::1", ttl => 10 },
    ],
}
--- request
GET /t
--- error_log
lru_cache SET: www.google.com|1 123456
shared_dict SET: www.google.com|1 123456
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456},{"address":"0:0:0:0:0:0:0:1","type":28,"class":1,"name":"l.www.google.com","ttl":10}]
