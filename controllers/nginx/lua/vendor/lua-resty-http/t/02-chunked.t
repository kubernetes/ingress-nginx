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
=== TEST 1: Non chunked.
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

            ngx.say(#body)
            httpc:close()
        ';
    }
    location = /b {
        chunked_transfer_encoding off;
        content_by_lua '
            local len = 32768
            local t = {}
            for i=1,len do
                t[i] = 0
            end
            ngx.print(table.concat(t))
        ';
    }
--- request
GET /a
--- response_body
32768
--- no_error_log
[error]
[warn]


=== TEST 2: Chunked. The number of chunks received when no max size is given proves the response was in fact chunked.
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

            local chunks = {}
            local c = 1
            repeat
                local chunk, err = res.body_reader()
                if chunk then
                    chunks[c] = chunk
                    c = c + 1
                end
            until not chunk

            local body = table.concat(chunks)

            ngx.say(#body)
            ngx.say(#chunks)
            httpc:close()
        ';
    }
    location = /b {
        content_by_lua '
            local len = 32768
            local t = {}
            for i=1,len do
                t[i] = 0
            end
            ngx.print(table.concat(t))
            local len = 32768
            local t = {}
            for i=1,len do
                t[i] = 0
            end
            ngx.print(table.concat(t))
        ';
    }
--- request
GET /a
--- response_body
65536
2
--- no_error_log
[error]
[warn]


=== TEST 3: Chunked using read_body method.
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

            ngx.say(#body)
            httpc:close()
        ';
    }
    location = /b {
        content_by_lua '
            local len = 32768
            local t = {}
            for i=1,len do
                t[i] = 0
            end
            ngx.print(table.concat(t))
            local len = 32768
            local t = {}
            for i=1,len do
                t[i] = 0
            end
            ngx.print(table.concat(t))
        ';
    }
--- request
GET /a
--- response_body
65536
--- no_error_log
[error]
[warn]
