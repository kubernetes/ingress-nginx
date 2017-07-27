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
=== TEST 1: parse_uri returns port 443 for https URIs
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            local parsed = httpc:parse_uri("https://www.google.com/foobar")
            ngx.say(parsed[3])
        ';
    }
--- request
GET /a
--- response_body
443
--- no_error_log
[error]
[warn]

=== TEST 2: parse_uri returns port 80 for http URIs
--- http_config eval: $::HttpConfig
--- config
    location = /a {
        content_by_lua '
            local http = require "resty.http"
            local httpc = http.new()
            local parsed = httpc:parse_uri("http://www.google.com/foobar")
            ngx.say(parsed[3])
        ';
    }
--- request
GET /a
--- response_body
80
--- no_error_log
[error]
[warn]
