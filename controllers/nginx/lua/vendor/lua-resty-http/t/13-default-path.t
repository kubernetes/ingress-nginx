# vim:set ft= ts=4 sw=4 et:

use Test::Nginx::Socket;
use Cwd qw(cwd);

plan tests => repeat_each() * (blocks() * 3);

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
=== TEST 1: request_uri (check the default path)
--- http_config eval: $::HttpConfig
--- config
    location /lua {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            
            local res, err = httpc:request_uri("http://127.0.0.1:"..ngx.var.server_port)

            if res and 200 == res.status then
                ngx.say("OK")
            else
                ngx.say("FAIL")
            end
        ';
    }

    location =/ {
        content_by_lua '
            ngx.print("OK")
        ';
    }
--- request
GET /lua
--- response_body
OK
--- no_error_log
[error]

