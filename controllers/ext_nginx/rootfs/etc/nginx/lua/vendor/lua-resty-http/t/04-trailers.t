# vim:set ft= ts=4 sw=4 et:

use Test::Nginx::Socket;
use Cwd qw(cwd);

plan tests => repeat_each() * (blocks() * 4);

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;;";
};

$ENV{TEST_NGINX_RESOLVER} = '8.8.8.8';

no_long_string();
#no_diff();

run_tests();

__DATA__
=== TEST 1: Trailers. Check Content-MD5 generated after the body is sent matches up.
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
                    ["TE"] = "trailers",
                }
            }

            local body = res:read_body()
            local hash = ngx.md5(body)
            res:read_trailers()

            if res.headers["Content-MD5"] == hash then
                ngx.say("OK")
            else
                ngx.say(res.headers["Content-MD5"])
            end
        ';
    }
    location = /b {
        content_by_lua '
            -- We use the raw socket to compose a response, since OpenResty
            -- doesnt support trailers natively.

            ngx.req.read_body()
            local sock, err = ngx.req.socket(true)
            if not sock then
                ngx.say(err)
            end

            local res = {}
            table.insert(res, "HTTP/1.1 200 OK")
            table.insert(res, "Date: " .. ngx.http_time(ngx.time()))
            table.insert(res, "Transfer-Encoding: chunked")
            table.insert(res, "Trailer: Content-MD5")
            table.insert(res, "")

            local body = "Hello, World"

            table.insert(res, string.format("%x", #body))
            table.insert(res, body)
            table.insert(res, "0")
            table.insert(res, "")
        
            table.insert(res, "Content-MD5: " .. ngx.md5(body))

            table.insert(res, "")
            table.insert(res, "")
            sock:send(table.concat(res, "\\r\\n"))
        ';
    }
--- request
GET /a
--- response_body
OK
--- no_error_log
[error]
[warn]


=== TEST 2: Advertised trailer does not exist, handled gracefully.
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
                    ["TE"] = "trailers",
                }
            }

            local body = res:read_body()
            local hash = ngx.md5(body)
            res:read_trailers()

            ngx.say("OK")
            httpc:close()
        ';
    }
    location = /b {
        content_by_lua '
            -- We use the raw socket to compose a response, since OpenResty
            -- doesnt support trailers natively.

            ngx.req.read_body()
            local sock, err = ngx.req.socket(true)
            if not sock then
                ngx.say(err)
            end

            local res = {}
            table.insert(res, "HTTP/1.1 200 OK")
            table.insert(res, "Date: " .. ngx.http_time(ngx.time()))
            table.insert(res, "Transfer-Encoding: chunked")
            table.insert(res, "Trailer: Content-MD5")
            table.insert(res, "")

            local body = "Hello, World"

            table.insert(res, string.format("%x", #body))
            table.insert(res, body)
            table.insert(res, "0")
            
            table.insert(res, "")
            table.insert(res, "")
            sock:send(table.concat(res, "\\r\\n"))
        ';
    }
--- request
GET /a
--- response_body
OK
--- no_error_log
[error]
[warn]


