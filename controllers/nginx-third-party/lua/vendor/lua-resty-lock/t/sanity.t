# vim:set ft= ts=4 sw=4 et:

use Test::Nginx::Socket::Lua;
use Cwd qw(cwd);

repeat_each(2);

plan tests => repeat_each() * (blocks() * 3);

my $pwd = cwd();

our $HttpConfig = qq{
    lua_package_path "$pwd/lib/?.lua;;";
    lua_package_cpath "/usr/local/openresty-debug/lualib/?.so;/usr/local/openresty/lualib/?.so;;";
    lua_shared_dict cache_locks 100k;
};

$ENV{TEST_NGINX_RESOLVER} = '8.8.8.8';
$ENV{TEST_NGINX_REDIS_PORT} ||= 6379;

no_long_string();
#no_diff();

run_tests();

__DATA__

=== TEST 1: lock is subject to garbage collection
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lock = require "resty.lock"
            for i = 1, 2 do
                collectgarbage("collect")
                local lock = lock:new("cache_locks")
                local elapsed, err = lock:lock("foo")
                ngx.say("lock: ", elapsed, ", ", err)
            end
            collectgarbage("collect")
        ';
    }
--- request
GET /t
--- response_body
lock: 0, nil
lock: 0, nil

--- no_error_log
[error]



=== TEST 2: serial lock and unlock
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lock = require "resty.lock"
            for i = 1, 2 do
                local lock = lock:new("cache_locks")
                local elapsed, err = lock:lock("foo")
                ngx.say("lock: ", elapsed, ", ", err)
                local ok, err = lock:unlock()
                if not ok then
                    ngx.say("failed to unlock: ", err)
                end
                ngx.say("unlock: ", ok)
            end
        ';
    }
--- request
GET /t
--- response_body
lock: 0, nil
unlock: 1
lock: 0, nil
unlock: 1

--- no_error_log
[error]



=== TEST 3: timed out locks
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lock = require "resty.lock"
            for i = 1, 2 do
                local lock1 = lock:new("cache_locks", { timeout = 0.01 })
                local lock2 = lock:new("cache_locks", { timeout = 0.01 })

                local elapsed, err = lock1:lock("foo")
                ngx.say("lock 1: lock: ", elapsed, ", ", err)

                local elapsed, err = lock2:lock("foo")
                ngx.say("lock 2: lock: ", elapsed, ", ", err)

                local ok, err = lock1:unlock()
                ngx.say("lock 1: unlock: ", ok, ", ", err)

                local ok, err = lock2:unlock()
                ngx.say("lock 2: unlock: ", ok, ", ", err)
            end
        ';
    }
--- request
GET /t
--- response_body
lock 1: lock: 0, nil
lock 2: lock: nil, timeout
lock 1: unlock: 1, nil
lock 2: unlock: nil, unlocked
lock 1: lock: 0, nil
lock 2: lock: nil, timeout
lock 1: unlock: 1, nil
lock 2: unlock: nil, unlocked

--- no_error_log
[error]



=== TEST 4: waited locks
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local resty_lock = require "resty.lock"
            local key = "blah"
            local t, err = ngx.thread.spawn(function ()
                local lock = resty_lock:new("cache_locks")
                local elapsed, err = lock:lock(key)
                ngx.say("sub thread: lock: ", elapsed, " ", err)
                ngx.sleep(0.1)
                ngx.say("sub thread: unlock: ", lock:unlock(key))
            end)

            local lock = resty_lock:new("cache_locks")
            local elapsed, err = lock:lock(key)
            ngx.say("main thread: lock: ", elapsed, " ", err)
            ngx.say("main thread: unlock: ", lock:unlock())
        ';
    }
--- request
GET /t
--- response_body_like chop
^sub thread: lock: 0 nil
sub thread: unlock: 1
main thread: lock: 0.12[6-9] nil
main thread: unlock: 1
$
--- no_error_log
[error]



=== TEST 5: waited locks (custom step)
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local resty_lock = require "resty.lock"
            local key = "blah"
            local t, err = ngx.thread.spawn(function ()
                local lock = resty_lock:new("cache_locks")
                local elapsed, err = lock:lock(key)
                ngx.say("sub thread: lock: ", elapsed, " ", err)
                ngx.sleep(0.1)
                ngx.say("sub thread: unlock: ", lock:unlock(key))
            end)

            local lock = resty_lock:new("cache_locks", { step = 0.01 })
            local elapsed, err = lock:lock(key)
            ngx.say("main thread: lock: ", elapsed, " ", err)
            ngx.say("main thread: unlock: ", lock:unlock())
        ';
    }
--- request
GET /t
--- response_body_like chop
^sub thread: lock: 0 nil
sub thread: unlock: 1
main thread: lock: 0.1[4-5]\d* nil
main thread: unlock: 1
$
--- no_error_log
[error]



=== TEST 6: waited locks (custom ratio)
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local resty_lock = require "resty.lock"
            local key = "blah"
            local t, err = ngx.thread.spawn(function ()
                local lock = resty_lock:new("cache_locks")
                local elapsed, err = lock:lock(key)
                ngx.say("sub thread: lock: ", elapsed, " ", err)
                ngx.sleep(0.1)
                ngx.say("sub thread: unlock: ", lock:unlock(key))
            end)

            local lock = resty_lock:new("cache_locks", { ratio = 3 })
            local elapsed, err = lock:lock(key)
            ngx.say("main thread: lock: ", elapsed, " ", err)
            ngx.say("main thread: unlock: ", lock:unlock())
        ';
    }
--- request
GET /t
--- response_body_like chop
^sub thread: lock: 0 nil
sub thread: unlock: 1
main thread: lock: 0.1[2]\d* nil
main thread: unlock: 1
$
--- no_error_log
[error]



=== TEST 7: waited locks (custom max step)
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local resty_lock = require "resty.lock"
            local key = "blah"
            local t, err = ngx.thread.spawn(function ()
                local lock = resty_lock:new("cache_locks")
                local elapsed, err = lock:lock(key)
                ngx.say("sub thread: lock: ", elapsed, " ", err)
                ngx.sleep(0.1)
                ngx.say("sub thread: unlock: ", lock:unlock(key))
            end)

            local lock = resty_lock:new("cache_locks", { max_step = 0.05 })
            local elapsed, err = lock:lock(key)
            ngx.say("main thread: lock: ", elapsed, " ", err)
            ngx.say("main thread: unlock: ", lock:unlock())
        ';
    }
--- request
GET /t
--- response_body_like chop
^sub thread: lock: 0 nil
sub thread: unlock: 1
main thread: lock: 0.11[2-4]\d* nil
main thread: unlock: 1
$
--- no_error_log
[error]



=== TEST 8: lock expired by itself
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local resty_lock = require "resty.lock"
            local key = "blah"
            local t, err = ngx.thread.spawn(function ()
                local lock = resty_lock:new("cache_locks", { exptime = 0.1 })
                local elapsed, err = lock:lock(key)
                ngx.say("sub thread: lock: ", elapsed, " ", err)
                ngx.sleep(0.1)
                -- ngx.say("sub thread: unlock: ", lock:unlock(key))
            end)

            local lock = resty_lock:new("cache_locks", { max_step = 0.05 })
            local elapsed, err = lock:lock(key)
            ngx.say("main thread: lock: ", elapsed, " ", err)
            ngx.say("main thread: unlock: ", lock:unlock())
        ';
    }
--- request
GET /t
--- response_body_like chop
^sub thread: lock: 0 nil
main thread: lock: 0.11[2-4]\d* nil
main thread: unlock: 1
$
--- no_error_log
[error]



=== TEST 9: ref & unref (1 at most)
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lock = require "resty.lock"
            local memo = lock.memo
            local ref = lock.ref_obj("foo")
            ngx.say(#memo)
            lock.unref_obj(ref)
            ngx.say(#memo)
            ref = lock.ref_obj("bar")
            ngx.say(#memo)
            lock.unref_obj(ref)
            ngx.say(#memo)
        ';
    }
--- request
GET /t
--- response_body
1
0
1
0

--- no_error_log
[error]



=== TEST 10: ref & unref (2 at most)
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lock = require "resty.lock"
            local memo = lock.memo

            for i = 1, 2 do
                local refs = {}

                refs[1] = lock.ref_obj("foo")
                ngx.say(#memo)

                refs[2] = lock.ref_obj("bar")
                ngx.say(#memo)

                lock.unref_obj(refs[1])
                ngx.say(#memo)

                lock.unref_obj(refs[2])
                ngx.say(#memo)
            end
        ';
    }
--- request
GET /t
--- response_body
1
2
2
2
2
2
1
1

--- no_error_log
[error]



=== TEST 11: lock on a nil key
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lock = require "resty.lock"
            local lock = lock:new("cache_locks")
            local elapsed, err = lock:lock(nil)
            if elapsed then
                ngx.say("lock: ", elapsed, ", ", err)
                local ok, err = lock:unlock()
                if not ok then
                    ngx.say("failed to unlock: ", err)
                end
            else
                ngx.say("failed to lock: ", err)
            end
        ';
    }
--- request
GET /t
--- response_body
failed to lock: nil key

--- no_error_log
[error]



=== TEST 12: same shdict, multple locks
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lock = require "resty.lock"
            local memo = lock.memo
            local lock1 = lock:new("cache_locks", { timeout = 0.01 })
            for i = 1, 3 do
                lock1:lock("lock_key")
                lock1:unlock()
                collectgarbage("collect")
            end

            local lock2 = lock:new("cache_locks", { timeout = 0.01 })
            local lock3 = lock:new("cache_locks", { timeout = 0.01 })
            lock2:lock("lock_key")
            lock3:lock("lock_key")
            collectgarbage("collect")

            ngx.say(#memo)

            lock2:unlock()
            lock3:unlock()
            collectgarbage("collect")
        ';
    }
--- request
GET /t
--- response_body
4
--- no_error_log
[error]



=== TEST 13: timed out locks (0 timeout)
--- http_config eval: $::HttpConfig
--- config
    location = /t {
        content_by_lua '
            local lock = require "resty.lock"
            for i = 1, 2 do
                local lock1 = lock:new("cache_locks", { timeout = 0 })
                local lock2 = lock:new("cache_locks", { timeout = 0 })

                local elapsed, err = lock1:lock("foo")
                ngx.say("lock 1: lock: ", elapsed, ", ", err)

                local elapsed, err = lock2:lock("foo")
                ngx.say("lock 2: lock: ", elapsed, ", ", err)

                local ok, err = lock1:unlock()
                ngx.say("lock 1: unlock: ", ok, ", ", err)

                local ok, err = lock2:unlock()
                ngx.say("lock 2: unlock: ", ok, ", ", err)
            end
        ';
    }
--- request
GET /t
--- response_body
lock 1: lock: 0, nil
lock 2: lock: nil, timeout
lock 1: unlock: 1, nil
lock 2: unlock: nil, unlocked
lock 1: lock: 0, nil
lock 2: lock: nil, timeout
lock 1: unlock: 1, nil
lock 2: unlock: nil, unlocked

--- no_error_log
[error]

