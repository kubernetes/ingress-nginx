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
=== TEST 1: Issue a notice (but do not error) if trying to read the request body in a subrequest
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        echo_location /b;
    }
    location = /b {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            httpc:connect("127.0.0.1", ngx.var.server_port)

            local res, err = httpc:request{
                path = "/c",
                headers = {
                    ["Content-Type"] = "application/x-www-form-urlencoded",
                }
            }
            if not res then
                ngx.say(err)
            end
            ngx.print(res:read_body())
            httpc:close()
        ';
    }
    location /c {
        echo "OK";
    }
--- request
GET /a
--- response_body
OK
--- no_error_log
[error]
[warn]

