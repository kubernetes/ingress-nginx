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
=== TEST 1: Test header normalisation
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http_headers = require "resty.http_headers"

            local headers = http_headers.new()

            headers.x_a_header = "a"
            headers["x-b-header"] = "b"
            headers["X-C-Header"] = "c"
            headers["X_d-HEAder"] = "d"

            ngx.say(headers["X-A-Header"])
            ngx.say(headers.x_b_header)
            
            for k,v in pairs(headers) do
                ngx.say(k, ": ", v)
            end
        ';
    }
--- request
GET /a
--- response_body
a
b
x-b-header: b
x-a-header: a
X-d-HEAder: d
X-C-Header: c
--- no_error_log
[error]
[warn]


=== TEST 2: Test headers can be accessed in all cases
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

            ngx.status = res.status
            ngx.say(res.headers["X-Foo-Header"])
            ngx.say(res.headers["x-fOo-heaDeR"])
            ngx.say(res.headers.x_foo_header)
            
            httpc:close()
        ';
    }
    location = /b {
        content_by_lua '
            ngx.header["X-Foo-Header"] = "bar"
            ngx.say("OK")
        ';
    }
--- request
GET /a
--- response_body
bar
bar
bar
--- no_error_log
[error]
[warn]


=== TEST 3: Test request headers are normalised
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
                headers = {
                    ["uSeR-AgENT"] = "test_user_agent",
                    x_foo = "bar",
                },
            }

            ngx.status = res.status
            ngx.print(res:read_body())
            
            httpc:close()
        ';
    }
    location = /b {
        content_by_lua '
            ngx.say(ngx.req.get_headers()["User-Agent"])
            ngx.say(ngx.req.get_headers()["X-Foo"])
        ';
    }
--- request
GET /a
--- response_body
test_user_agent
bar
--- no_error_log
[error]


=== TEST 4: Test that headers remain unique
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http_headers = require "resty.http_headers"

            local headers = http_headers.new()

            headers["x-a-header"] = "a"
            headers["X-A-HEAder"] = "b"

            for k,v in pairs(headers) do
            ngx.log(ngx.DEBUG, k, ": ", v)
                ngx.header[k] = v
            end
        ';
    }
--- request
GET /a
--- response_headers
x-a-header: b
--- no_error_log
[error]
[warn]
[warn]
