use lib 't';
use TestDNS;
use Cwd qw(cwd);

plan tests => repeat_each() * 17;

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;;";
    lua_socket_log_errors off;
};

no_long_string();
run_tests();

__DATA__
=== TEST 1: Query is triggered when cache is expired
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
                resolver = {nameservers = {{"127.0.0.1", "1953"}}},
                max_stale = 10
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

            dns._debug(true)

            -- Sleep beyond response TTL
            ngx.sleep(1.1)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                if stale then
                    answer = stale
                else
                    ngx.say(err)
                end
            end

            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

            ngx.sleep(0.1)

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
Returning STALE
Attempting to repopulate 'www.google.com'
Repopulating 'www.google.com'
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":1}]
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":0}]

=== TEST 2: Query is not triggered when cache expires and max_stale is disabled
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
                resolver = {nameservers = {{"127.0.0.1", "1953"}}, retrans = 1, timeout = 50 },
                max_stale = 0
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end

            dns._debug(true)

            -- Sleep beyond response TTL
            ngx.sleep(1.1)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                if stale then
                    answer = stale
                else
                    ngx.say(err)
                end
            end

            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

            ngx.sleep(0.1)
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
--- no_error_log
Attempting to repopulate 'www.google.com'
Repopulating 'www.google.com'
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":0}]


=== TEST 3: Repopulate ignores max_stale
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
                resolver = {nameservers = {{"127.0.0.1", "1953"}}, retrans = 1, timeout = 50 },
                max_stale = 10,
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

            -- Sleep beyond response TTL
            ngx.sleep(1.1)

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                if stale then
                    answer = stale
                else
                    ngx.say(err)
                end
            end

            local cjson = require"cjson"
            ngx.say(cjson.encode(answer))

            ngx.sleep(0.1)
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
Repopulating 'www.google.com'
Querying: www.google.com
Resolver error
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":0}]

=== TEST 4: Multiple queries only trigger 1 repopulate timer
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
                resolver = {nameservers = {{"127.0.0.1", "1953"}}, retrans = 1, timeout = 50 },
                repopulate = true,
            })
            if not dns then
                ngx.say(err)
            end
            dns.resolver._id = 125

            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
            dns._debug(true)
            local answer, err, stale = dns:query("www.google.com", { qtype = dns.TYPE_A })
            if not answer then
                ngx.say(err)
            end
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
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 1 }],
}
--- request
GET /t
--- no_error_log
Attempting to repopulate www.google.com
--- response_body
[{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":1}]
