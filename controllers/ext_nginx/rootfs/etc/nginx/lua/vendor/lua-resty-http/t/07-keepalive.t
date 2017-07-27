# vim:set ft= ts=4 sw=4 et:

use Test::Nginx::Socket;
use Cwd qw(cwd);

plan tests => repeat_each() * (blocks() * 4);

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;;";
    error_log logs/error.log debug;
};

$ENV{TEST_NGINX_RESOLVER} = '8.8.8.8';

no_long_string();
#no_diff();

run_tests();

__DATA__
=== TEST 1 Simple interface, Connection: Keep-alive. Test the connection is reused.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            local res, err = httpc:request_uri(
                "http://127.0.0.1:"..ngx.var.server_port.."/b", {
                }
            )

            ngx.say(res.headers["Connection"])

            httpc:connect("127.0.0.1", ngx.var.server_port)
            ngx.say(httpc:get_reused_times())
        ';
    }
    location = /b {
        content_by_lua '
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- response_body
keep-alive
1
--- no_error_log
[error]
[warn]


=== TEST 2 Simple interface, Connection: close, test we don't try to keepalive, but also that subsequent connections can keepalive.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            local res, err = httpc:request_uri(
                "http://127.0.0.1:"..ngx.var.server_port.."/b", {
                    version = 1.0,
                    headers = {
                        ["Connection"] = "close",
                    },
                }
            )

            httpc:connect("127.0.0.1", ngx.var.server_port)
            ngx.say(httpc:get_reused_times())
            
            httpc:set_keepalive()

            httpc:connect("127.0.0.1", ngx.var.server_port)
            ngx.say(httpc:get_reused_times())
        ';
    }
    location = /b {
        content_by_lua '
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- response_body
0
1
--- no_error_log
[error]
[warn]


=== TEST 3 Generic interface, Connection: Keep-alive. Test the connection is reused.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b"
            }

            local body = res:read_body()

            ngx.say(res.headers["Connection"])
            ngx.say(httpc:set_keepalive())

            httpc:connect("127.0.0.1", ngx.var.server_port)
            ngx.say(httpc:get_reused_times())
        ';
    }
    location = /b {
        content_by_lua '
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- response_body
keep-alive
1
1
--- no_error_log
[error]
[warn]


=== TEST 4 Generic interface, Connection: Close. Test we don't try to keepalive, but also that subsequent connections can keepalive.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                version = 1.0,
                headers = {
                    ["Connection"] = "Close",
                },
                path = "/b"
            }

            local body = res:read_body()

            ngx.say(res.headers["Connection"])
            local r, e = httpc:set_keepalive()
            ngx.say(r)
            ngx.say(e)

            httpc:connect("127.0.0.1", ngx.var.server_port)
            ngx.say(httpc:get_reused_times())

            httpc:set_keepalive()

            httpc:connect("127.0.0.1", ngx.var.server_port)
            ngx.say(httpc:get_reused_times())
        ';
    }
    location = /b {
        content_by_lua '
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- response_body
close
2
connection must be closed
0
1
--- no_error_log
[error]
[warn]


=== TEST 5: Generic interface, HTTP 1.0, no connection header. Test we don't try to keepalive, but also that subsequent connections can keepalive.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", 12345)

            local res, err = httpc:request{
                version = 1.0,
                path = "/b"
            }

            local body = res:read_body()
            ngx.print(body)

            ngx.say(res.headers["Connection"])

            local r, e = httpc:set_keepalive()
            ngx.say(r)
            ngx.say(e)

            httpc:connect("127.0.0.1", ngx.var.server_port)
            ngx.say(httpc:get_reused_times())

            httpc:set_keepalive()

            httpc:connect("127.0.0.1", ngx.var.server_port)
            ngx.say(httpc:get_reused_times())
        ';
    }
    location = /b {
        content_by_lua '
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- tcp_listen: 12345
--- tcp_reply
HTTP/1.0 200 OK
Date: Fri, 08 Aug 2016 08:12:31 GMT
Server: OpenResty

OK
--- response_body
OK
nil
2
connection must be closed
0
1
--- no_error_log
[error]
[warn]
