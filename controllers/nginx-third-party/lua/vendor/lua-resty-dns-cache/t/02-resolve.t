
use lib 't';
use TestDNS;
use Cwd qw(cwd);

plan tests => repeat_each() * 12;

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;;";
};

no_long_string();
run_tests();

__DATA__
=== TEST 1: Can resolve with lru + dict
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
                resolver = {
                    nameservers = { {"127.0.0.1", "1953"} }
                }
            })
            if not dns then
                ngx.say(err)
            end
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
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
}
--- request
GET /t
--- no_error_log
[error]
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}]

=== TEST 2: Can resolve with lru only
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
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                resolver = {
                    nameservers = { {"127.0.0.1", "1953"} }
                }
            })
            if not dns then
                ngx.say(err)
            end
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
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
}
--- request
GET /t
--- no_error_log
[error]
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}]

=== TEST 3: Can resolve with dict only
--- http_config eval
"$::HttpConfig"
. q{
    lua_shared_dict dns_cache 1m;
}
--- config
    location /t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                dict = "dns_cache",
                resolver = {
                    nameservers = { {"127.0.0.1", "1953"} }
                }
            })
            if not dns then
                ngx.say(err)
            end
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
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
}
--- request
GET /t
--- no_error_log
[error]
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}]

=== TEST 4: Can resolve with no cache, error thrown
--- http_config eval
"$::HttpConfig"
--- config
    location /t {
        content_by_lua '
            local DNS_Cache = require("resty.dns.cache")
            local dns, err = DNS_Cache.new({
                resolver = {
                    nameservers = { {"127.0.0.1", "1953"} }
                }
            })
            if not dns then
                ngx.say(err)
            end
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
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
}
--- request
GET /t
--- error_log
No cache defined
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}]

