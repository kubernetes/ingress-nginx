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
=== TEST 1 Test that pipelined reqests can be read correctly.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)
            
            local responses = httpc:request_pipeline{
                {
                    path = "/b",
                },
                {
                    path = "/c",
                },
                {
                    path = "/d",
                }
            }

            for i,r in ipairs(responses) do
                if r.status then
                    ngx.say(r.status)
                    ngx.say(r.headers["X-Res"])
                    ngx.say(r:read_body())
                end
            end
        ';
    }
    location = /b {
        content_by_lua '
            ngx.status = 200
            ngx.header["X-Res"] = "B"
            ngx.print("B")
        ';
    }
    location = /c {
        content_by_lua '
            ngx.status = 404
            ngx.header["X-Res"] = "C"
            ngx.print("C")
        ';
    }
    location = /d {
        content_by_lua '
            ngx.status = 200
            ngx.header["X-Res"] = "D"
            ngx.print("D")
        ';
    }
--- request
GET /a
--- response_body
200
B
B
404
C
C
200
D
D
--- no_error_log
[error]
[warn]


=== TEST 2: Test we can handle timeouts on reading the pipelined requests.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)
            httpc:set_timeout(1)
            
            local responses = httpc:request_pipeline{
                {
                    path = "/b",
                },
                {
                    path = "/c",
                },
            }

            for i,r in ipairs(responses) do
                if r.status then
                    ngx.say(r.status)
                    ngx.say(r.headers["X-Res"])
                    ngx.say(r:read_body())
                end
            end
        ';
    }
    location = /b {
        content_by_lua '
            ngx.status = 200
            ngx.header["X-Res"] = "B"
            ngx.print("B")
        ';
    }
    location = /c {
        content_by_lua '
            ngx.status = 404
            ngx.header["X-Res"] = "C"
            ngx.sleep(1)
            ngx.print("C")
        ';
    }
--- request
GET /a
--- response_body
200
B
B
--- no_error_log
[warn]
--- error_log eval
[qr/timeout/]
