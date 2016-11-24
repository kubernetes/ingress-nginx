# vim:set ft= ts=4 sw=4 et:

use Test::Nginx::Socket;
use Cwd qw(cwd);

plan tests => repeat_each() * (blocks() * 4) - 1;

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;;";
};

$ENV{TEST_NGINX_RESOLVER} = '8.8.8.8';

no_long_string();
#no_diff();

run_tests();

__DATA__
=== TEST 1: Chunked streaming body reader returns the right content length.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
            }

            local chunks = {}
            repeat
                local chunk = res.body_reader()
                if chunk then
                    table.insert(chunks, chunk)
                end
            until not chunk

            local body = table.concat(chunks)
            ngx.say(#body)
            ngx.say(res.headers["Transfer-Encoding"])

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
        ';
    }
--- request
GET /a
--- response_body
32768
chunked
--- no_error_log
[error]
[warn]


=== TEST 2: Non-Chunked streaming body reader returns the right content length.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
            }

            local chunks = {}
            repeat
                local chunk = res.body_reader()
                if chunk then
                    table.insert(chunks, chunk)
                end
            until not chunk

            local body = table.concat(chunks)
            ngx.say(#body)
            ngx.say(res.headers["Transfer-Encoding"])
            ngx.say(#chunks)

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
nil
1
--- no_error_log
[error]
[warn]


=== TEST 2b: Non-Chunked streaming body reader, buffer size becomes nil
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
            }

            local chunks = {}
            local buffer_size = 16384
            repeat
                local chunk = res.body_reader(buffer_size)
                if chunk then
                    table.insert(chunks, chunk)
                end

                buffer_size = nil
            until not chunk

            local body = table.concat(chunks)
            ngx.say(res.headers["Transfer-Encoding"])

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
nil
--- error_log
Buffer size not specified, bailing


=== TEST 3: HTTP 1.0 body reader with no max size returns the right content length.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
                version = 1.0,
            }

            local chunks = {}
            repeat
                local chunk = res.body_reader()
                if chunk then
                    table.insert(chunks, chunk)
                end
            until not chunk

            local body = table.concat(chunks)
            ngx.say(#body)
            ngx.say(res.headers["Transfer-Encoding"])
            ngx.say(#chunks)

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
nil
1
--- no_error_log
[error]
[warn]


=== TEST 4: HTTP 1.0 body reader with max chunk size returns the right content length.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
                version = 1.0,
            }

            local chunks = {}
            local size = 8192
            repeat
                local chunk = res.body_reader(size)
                if chunk then
                    table.insert(chunks, chunk)
                end
                size = size + size
            until not chunk

            local body = table.concat(chunks)
            ngx.say(#body)
            ngx.say(res.headers["Transfer-Encoding"])
            ngx.say(#chunks)

            httpc:close()
        ';
    }
    location = /b {
        chunked_transfer_encoding off;
        content_by_lua '
            local len = 32769
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
32769
nil
3
--- no_error_log
[error]
[warn]


=== TEST 4b: HTTP 1.0 body reader with no content length, stream works as expected.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
                version = 1.0,
            }

            local chunks = {}
            local size = 8192
            repeat
                local chunk = res.body_reader(size)
                if chunk then
                    table.insert(chunks, chunk)
                end
                size = size + size
            until not chunk

            local body = table.concat(chunks)
            ngx.say(#body)
            ngx.say(#chunks)

            httpc:close()
        ';
    }
    location = /b {
        content_by_lua '
            ngx.req.read_body()
            local sock, err = ngx.req.socket(true)
            if not sock then
                ngx.say(err)
            end

            local res = {}
            table.insert(res, "HTTP/1.0 200 OK")
            table.insert(res, "Date: " .. ngx.http_time(ngx.time()))
            table.insert(res, "")

            local len = 32769
            local t = {}
            for i=1,len do
                t[i] = 0
            end
            table.insert(res, table.concat(t))
            sock:send(table.concat(res, "\\r\\n"))
        ';
    }
--- request
GET /a
--- response_body
32769
3
--- no_error_log
[error]
[warn]


=== TEST 5: Chunked streaming body reader with max chunk size returns the right content length.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
            }

            local chunks = {}
            local size = 8192
            repeat
                local chunk = res.body_reader(size)
                if chunk then
                    table.insert(chunks, chunk)
                end
                size = size + size
            until not chunk

            local body = table.concat(chunks)
            ngx.say(#body)
            ngx.say(res.headers["Transfer-Encoding"])
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
        ';
    }
--- request
GET /a
--- response_body
32768
chunked
3
--- no_error_log
[error]
[warn]


=== TEST 6: Request reader correctly reads body
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        lua_need_request_body off;
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()

            local reader, err = httpc:get_client_body_reader(8192)

            repeat
                local chunk, err = reader()
                if chunk then
                    ngx.print(chunk)
                end
            until chunk == nil

        ';
    }

--- request
POST /a
foobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbaz
--- response_body: foobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbaz
--- no_error_log
[error]
[warn]

=== TEST 7: Request reader correctly reads body in chunks
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        lua_need_request_body off;
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()

            local reader, err = httpc:get_client_body_reader(64)

            local chunks = 0
            repeat
                chunks = chunks +1
                local chunk, err = reader()
                if chunk then
                    ngx.print(chunk)
                end
            until chunk == nil
            ngx.say("\\n"..chunks)
        ';
    }

--- request
POST /a
foobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbaz
--- response_body
foobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbaz
3
--- no_error_log
[error]
[warn]


=== TEST 8: Request reader passes into client
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        lua_need_request_body off;
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local reader, err = httpc:get_client_body_reader(64)

            local res, err = httpc:request{
                method = POST,
                path = "/b",
                body = reader,
                headers = ngx.req.get_headers(100, true),
            }

            local body = res:read_body()
            ngx.say(body)
            httpc:close()

        ';
    }

    location = /b {
        content_by_lua '
            ngx.req.read_body()
            local body, err = ngx.req.get_body_data()
            ngx.print(body)
        ';
    }

--- request
POST /a
foobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbaz
--- response_body
foobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbaz
--- no_error_log
[error]
[warn]


=== TEST 9: Body reader is a function returning nil when no body is present.
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/b",
                method = "HEAD",
            }

            repeat
                local chunk = res.body_reader()
            until not chunk

            httpc:close()
        ';
    }
    location = /b {
        content_by_lua '
            ngx.exit(200)
        ';
    }
--- request
GET /a
--- no_error_log
[error]
[warn]


=== TEST 10: Issue a notice (but do not error) if trying to read the request body in a subrequest
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        echo_location /b;
    }
    location = /b {
        lua_need_request_body off;
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()

            local reader, err = httpc:get_client_body_reader(8192)
            if not reader then
                ngx.log(ngx.NOTICE, err)
                return
            end

            repeat
                local chunk, err = reader()
                if chunk then
                    ngx.print(chunk)
                end
            until chunk == nil
        ';
    }

--- request
POST /a
foobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbazfoobarbaz
--- response_body: 
--- no_error_log
[error]
[warn]
--- error_log
attempt to read the request body in a subrequest
