# vim:set ft= ts=4 sw=4 et:

use Test::Nginx::Socket;
use Cwd qw(cwd);

plan tests => repeat_each() * (blocks() * 4) + 6;

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
=== TEST 1: Simple URI interface
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            local res, err = httpc:request_uri("http://127.0.0.1:"..ngx.var.server_port.."/b?a=1&b=2")
            
            if not res then
                ngx.log(ngx.ERR, err)
            end
            ngx.status = res.status

            ngx.header["X-Header-A"] = res.headers["X-Header-A"]
            ngx.header["X-Header-B"] = res.headers["X-Header-B"]

            ngx.print(res.body)
        ';
    }
    location = /b {
        content_by_lua '
            for k,v in pairs(ngx.req.get_uri_args()) do
                ngx.header["X-Header-" .. string.upper(k)] = v
            end
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- response_headers
X-Header-A: 1
X-Header-B: 2
--- response_body
OK
--- no_error_log
[error]
[warn]


=== TEST 2: Simple URI interface HTTP 1.0
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            local res, err = httpc:request_uri(
                "http://127.0.0.1:"..ngx.var.server_port.."/b?a=1&b=2", {
                }
            )
            
            ngx.status = res.status
            
            ngx.header["X-Header-A"] = res.headers["X-Header-A"]
            ngx.header["X-Header-B"] = res.headers["X-Header-B"]

            ngx.print(res.body)
        ';
    }
    location = /b {
        content_by_lua '
            for k,v in pairs(ngx.req.get_uri_args()) do
                ngx.header["X-Header-" .. string.upper(k)] = v
            end
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- response_headers
X-Header-A: 1
X-Header-B: 2
--- response_body
OK
--- no_error_log
[error]
[warn]


=== TEST 3 Simple URI interface, params override
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            local res, err = httpc:request_uri(
                "http://127.0.0.1:"..ngx.var.server_port.."/b?a=1&b=2", {
                    path = "/c",
                    query = {
                        a = 2,
                        b = 3,
                    },
                }
            )
            
            ngx.status = res.status
            
            ngx.header["X-Header-A"] = res.headers["X-Header-A"]
            ngx.header["X-Header-B"] = res.headers["X-Header-B"]

            ngx.print(res.body)
        ';
    }
    location = /c {
        content_by_lua '
            for k,v in pairs(ngx.req.get_uri_args()) do
                ngx.header["X-Header-" .. string.upper(k)] = v
            end
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- response_headers
X-Header-A: 2
X-Header-B: 3
--- response_body
OK
--- no_error_log
[error]
[warn]
