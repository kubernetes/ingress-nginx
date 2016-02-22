# vim:set ft= ts=4 sw=4 et:

use lib 't';
use TestDNS;
use Cwd qw(cwd);

repeat_each(2);

plan tests => repeat_each() * (3 * blocks() + 14);

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;$pwd/t/lib/?.lua;;";
    lua_package_cpath "/usr/local/openresty-debug/lualib/?.so;/usr/local/openresty/lualib/?.so;;";
};

$ENV{TEST_NGINX_RESOLVER} ||= '8.8.8.8';

log_level('notice');

no_long_string();

run_tests();

__DATA__

=== TEST 1: single answer reply, good A answer
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
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
--- udp_query eval
"\x{00}}\x{01}\x{00}\x{00}\x{01}\x{00}\x{00}\x{00}\x{00}\x{00}\x{00}\x{03}www\x{06}google\x{03}com\x{00}\x{00}\x{01}\x{00}\x{01}"
--- response_body
records: [{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}]
--- no_error_log
[error]



=== TEST 2: empty answer reply
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    qname => 'www.google.com',
    opcode => 0,
}
--- request
GET /t
--- response_body
records: {}
--- no_error_log
[error]



=== TEST 3: one byte reply, truncated, without TCP server
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
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
--- udp_listen: 1953
--- udp_reply: a
--- request
GET /t
--- response_body
failed to query: failed to connect to TCP server 127.0.0.1:1953: connection refused
--- error_log
connect() failed



=== TEST 4: empty reply
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply:
--- request
GET /t
--- response_body
failed to query: failed to connect to TCP server 127.0.0.1:1953: connection refused
--- error_log
connect() failed



=== TEST 5: two answers reply that contains AAAA records
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
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
        { name => "l.www.google.com", ipv6 => "::1", ttl => 0 },
    ],
}
--- request
GET /t
--- response_body
records: [{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456},{"address":"0:0:0:0:0:0:0:1","type":28,"class":1,"name":"l.www.google.com","ttl":0}]
--- no_error_log
[error]



=== TEST 6: good CNAME answer
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "www.google.com", cname => "blah.google.com", ttl => 125 },
    ],
}
--- request
GET /t
--- response_body
records: [{"ttl":125,"type":5,"class":1,"name":"www.google.com","cname":"blah.google.com"}]
--- no_error_log
[error]



=== TEST 7: CNAME answer with bad rd length
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "www.google.com", cname => "blah.google.com", ttl => 125, rdlength => 3 },
    ],
}
--- request
GET /t
--- response_body
failed to query: bad cname record length: 17 ~= 3
--- no_error_log
[error]



=== TEST 8: single answer reply, bad A answer, wrong record length
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456, rdlength => 1 }],
}
--- request
GET /t
--- response_body
failed to query: bad A record value length: 1
--- no_error_log
[error]



=== TEST 9: bad AAAA record, wrong len
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "l.www.google.com", ipv6 => "::1", ttl => 0, rdlength => 21 },
    ],
}
--- request
GET /t
--- response_body
failed to query: bad AAAA record value length: 21
--- no_error_log
[error]



=== TEST 10: timeout
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 1,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(100)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply_delay: 200ms
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "l.www.google.com", ipv6 => "::1", ttl => 0 },
    ],
}
--- request
GET /t
--- response_body
failed to query: failed to receive reply from UDP server 127.0.0.1:1953: timeout
--- error_log
lua udp socket read timed out
--- timeout: 3



=== TEST 11: not timeout finally (re-transmission works)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply_delay: 500ms
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "l.www.google.com", ipv6 => "FF01::101", ttl => 0 },
    ],
}
--- request
GET /t
--- response_body
records: [{"address":"ff01:0:0:0:0:0:0:101","type":28,"class":1,"name":"l.www.google.com","ttl":0}]
--- error_log
lua udp socket read timed out
--- timeout: 3



=== TEST 12: timeout finally (re-transmission works but not enough retrans times)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 2,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply_delay: 500ms
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "l.www.google.com", ipv6 => "FF01::101", ttl => 0 },
    ],
}
--- request
GET /t
--- response_body
failed to query: failed to receive reply from UDP server 127.0.0.1:1953: timeout
--- error_log
lua udp socket read timed out
--- timeout: 3



=== TEST 13: RCODE - format error
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    rcode => 1,
    opcode => 0,
    qname => 'www.google.com',
}
--- request
GET /t
--- response_body
records: {"errcode":1,"errstr":"format error"}
--- no_error_log
[error]



=== TEST 14: RCODE - server failure
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"

            if ans.errcode then
                ngx.say("error code: ", ans.errcode, ": ", ans.errstr)
            end

            for i, rec in ipairs(ans) do
                ngx.say("record: ", cjson.encode(rec))
            end
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    rcode => 2,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
}
--- request
GET /t
--- response_body
error code: 2: server failure
record: {"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}
--- no_error_log
[error]



=== TEST 15: RCODE - name error
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
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
--- response_body
records: {"errcode":3,"errstr":"name error"}
--- no_error_log
[error]



=== TEST 16: RCODE - not implemented
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    rcode => 4,
    opcode => 0,
    qname => 'www.google.com',
}
--- request
GET /t
--- response_body
records: {"errcode":4,"errstr":"not implemented"}
--- no_error_log
[error]



=== TEST 17: RCODE - refused
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    rcode => 5,
    opcode => 0,
    qname => 'www.google.com',
}
--- request
GET /t
--- response_body
records: {"errcode":5,"errstr":"refused"}
--- no_error_log
[error]



=== TEST 18: RCODE - unknown
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    rcode => 6,
    opcode => 0,
    qname => 'www.google.com',
}
--- request
GET /t
--- response_body
records: {"errcode":6,"errstr":"unknown"}
--- no_error_log
[error]



=== TEST 19: TC (TrunCation) = 1, no TCP server
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(1000)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    tc => 1,
    qname => 'www.google.com',
}
--- request
GET /t
--- response_body
failed to query: failed to connect to TCP server 127.0.0.1:1953: connection refused
--- error_log
connect() failed



=== TEST 20: bad QR flag (0)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(1000)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    qr => 0,
    qname => 'www.google.com',
}
--- request
GET /t
--- response_body
failed to query: bad QR flag in the DNS response
--- no_error_log
[error]



=== TEST 21: Recursion Desired off
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
                no_recurse = true,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(200)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_query eval
"\x{00}}\x{00}\x{00}\x{00}\x{01}\x{00}\x{00}\x{00}\x{00}\x{00}\x{00}\x{03}www\x{06}google\x{03}com\x{00}\x{00}\x{01}\x{00}\x{01}"
--- udp_reply dns
{
    id => 125,
    qname => 'www.google.com',
}
--- request
GET /t
--- response_body
records: {}
--- no_error_log
[error]



=== TEST 22: id mismatch (timeout)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                timeout = 10,
                retrans = 2,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 126,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
}
--- request
GET /t
--- udp_query eval
"\x{00}}\x{01}\x{00}\x{00}\x{01}\x{00}\x{00}\x{00}\x{00}\x{00}\x{00}\x{03}www\x{06}google\x{03}com\x{00}\x{00}\x{01}\x{00}\x{01}"
--- response_body
failed to query: failed to receive reply from UDP server 127.0.0.1:1953: timeout
--- error_log
id mismatch in the DNS reply: 126 ~= 125
--- log_level: debug



=== TEST 23: id mismatch (and then match)
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                timeout = 10,
                retrans = 2,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
[
{
    id => 126,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
},
{
    id => 127,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
},
{
    id => 120,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
},
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 }],
},
]
--- request
GET /t
--- udp_query eval
"\x{00}}\x{01}\x{00}\x{00}\x{01}\x{00}\x{00}\x{00}\x{00}\x{00}\x{00}\x{03}www\x{06}google\x{03}com\x{00}\x{00}\x{01}\x{00}\x{01}"
--- response_body
records: [{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456}]
--- error_log
id mismatch in the DNS reply: 126 ~= 125
id mismatch in the DNS reply: 120 ~= 125
id mismatch in the DNS reply: 127 ~= 125
--- log_level: debug



=== TEST 24: TC (TrunCation) = 1, with TCP server
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} },
                retrans = 3,
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r:set_timeout(1000)   -- in ms

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    tc => 1,
    qname => 'www.google.com',
}
--- tcp_listen: 1953
--- tcp_query_len: 34
--- tcp_reply dns=tcp
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 },
        { name => "l.www.google.com", ipv6 => "::1", ttl => 0 },
    ],
}
--- request
GET /t
--- response_body
records: [{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456},{"address":"0:0:0:0:0:0:0:1","type":28,"class":1,"name":"l.www.google.com","ttl":0}]
--- no_error_log
[error]
--- error_log
query the TCP server due to reply truncation
--- log_level: debug



=== TEST 25: one byte reply, truncated, with TCP server
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_A })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local cjson = require "cjson"
            ngx.say("records: ", cjson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply: a
--- tcp_listen: 1953
--- tcp_query_len: 34
--- tcp_reply dns=tcp
{
    id => 125,
    opcode => 0,
    qname => 'www.google.com',
    answer => [
        { name => "www.google.com", ipv4 => "127.0.0.1", ttl => 123456 },
        { name => "l.www.google.com", ipv6 => "::1", ttl => 0 },
    ],
}
--- request
GET /t
--- response_body
records: [{"address":"127.0.0.1","type":1,"class":1,"name":"www.google.com","ttl":123456},{"address":"0:0:0:0:0:0:0:1","type":28,"class":1,"name":"l.www.google.com","ttl":0}]
--- no_error_log
[error]
--- error_log
query the TCP server due to reply truncation
--- log_level: debug



=== TEST 26: single answer reply, TXT answer with a single char string
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_TXT })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local ljson = require "ljson"
            ngx.say("records: ", ljson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qtype => 16,  # TXT
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", txt => "\5hello", ttl => 123456 }],
}
--- request
GET /t
--- udp_query eval
"\x{00}}\x{01}\x{00}\x{00}\x{01}\x{00}\x{00}\x{00}\x{00}\x{00}\x{00}\x{03}www\x{06}google\x{03}com\x{00}\x{00}\x{10}\x{00}\x{01}"
--- response_body
records: [{"class":1,"name":"www.google.com","ttl":123456,"txt":"hello","type":16}]
--- no_error_log
[error]



=== TEST 27: single answer reply, TXT answer with a null char string
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_TXT })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local ljson = require "ljson"
            ngx.say("records: ", ljson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qtype => 16,  # TXT
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", txt => "\0", ttl => 123456 }],
}
--- request
GET /t
--- udp_query eval
"\x{00}}\x{01}\x{00}\x{00}\x{01}\x{00}\x{00}\x{00}\x{00}\x{00}\x{00}\x{03}www\x{06}google\x{03}com\x{00}\x{00}\x{10}\x{00}\x{01}"
--- response_body
records: [{"class":1,"name":"www.google.com","ttl":123456,"txt":"","type":16}]
--- no_error_log
[error]



=== TEST 28: single answer reply, TXT answer with a multiple char strings
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_TXT })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local ljson = require "ljson"
            ngx.say("records: ", ljson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qtype => 16,  # TXT
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", txt => "\5hello\5world", ttl => 123456 }],
}
--- request
GET /t
--- udp_query eval
"\x{00}}\x{01}\x{00}\x{00}\x{01}\x{00}\x{00}\x{00}\x{00}\x{00}\x{00}\x{03}www\x{06}google\x{03}com\x{00}\x{00}\x{10}\x{00}\x{01}"
--- response_body
records: [{"class":1,"name":"www.google.com","ttl":123456,"txt":["hello","world"],"type":16}]
--- no_error_log
[error]



=== TEST 29: single answer reply, multiple TXT answers
--- http_config eval: $::HttpConfig
--- config
    location /t {
        content_by_lua '
            local resolver = require "resty.dns.resolver"

            local r, err = resolver:new{
                nameservers = { {"127.0.0.1", 1953} }
            }
            if not r then
                ngx.say("failed to instantiate resolver: ", err)
                return
            end

            r._id = 125

            local ans, err = r:query("www.google.com", { qtype = r.TYPE_TXT })
            if not ans then
                ngx.say("failed to query: ", err)
                return
            end

            local ljson = require "ljson"
            ngx.say("records: ", ljson.encode(ans))
        ';
    }
--- udp_listen: 1953
--- udp_reply dns
{
    id => 125,
    opcode => 0,
    qtype => 16,  # TXT
    qname => 'www.google.com',
    answer => [{ name => "www.google.com", txt => "\5hello\6world!", ttl => 123456 }, { name => "www.google.com", txt => "\4blah", ttl => 123456 }],
}
--- request
GET /t
--- udp_query eval
"\x{00}}\x{01}\x{00}\x{00}\x{01}\x{00}\x{00}\x{00}\x{00}\x{00}\x{00}\x{03}www\x{06}google\x{03}com\x{00}\x{00}\x{10}\x{00}\x{01}"
--- response_body
records: [{"class":1,"name":"www.google.com","ttl":123456,"txt":["hello","world!"],"type":16},{"class":1,"name":"www.google.com","ttl":123456,"txt":"blah","type":16}]
--- no_error_log
[error]

