# lua-resty-http

Lua HTTP client cosocket driver for [OpenResty](http://openresty.org/) / [ngx_lua](https://github.com/openresty/lua-nginx-module).

# Status

Production ready.

# Features

* HTTP 1.0 and 1.1
* SSL
* Streaming interface to the response body, for predictable memory usage
* Alternative simple interface for singleshot requests without manual connection step
* Chunked and non-chunked transfer encodings
* Keepalive
* Pipelining
* Trailers


# API

* [new](#name)
* [connect](#connect)
* [set_timeout](#set_timeout)
* [ssl_handshake](#ssl_handshake)
* [set_keepalive](#set_keepalive)
* [get_reused_times](#get_reused_times)
* [close](#close)
* [request](#request)
* [request_uri](#request_uri)
* [request_pipeline](#request_pipeline)
* [Response](#response)
    * [body_reader](#resbody_reader)
    * [read_body](#resread_body)
    * [read_trailes](#resread_trailers)
* [Proxy](#proxy)
    * [proxy_request](#proxy_request)
    * [proxy_response](#proxy_response)
* [Utility](#utility)
    * [parse_uri](#parse_uri)
    * [get_client_body_reader](#get_client_body_reader)


## Synopsis

```` lua
lua_package_path "/path/to/lua-resty-http/lib/?.lua;;";

server {


  location /simpleinterface {
    resolver 8.8.8.8;  # use Google's open DNS server for an example

    content_by_lua '

      -- For simple singleshot requests, use the URI interface.
      local http = require "resty.http"
      local httpc = http.new()
      local res, err = httpc:request_uri("http://example.com/helloworld", {
        method = "POST",
        body = "a=1&b=2",
        headers = {
          ["Content-Type"] = "application/x-www-form-urlencoded",
        }
      })

      if not res then
        ngx.say("failed to request: ", err)
        return
      end

      -- In this simple form, there is no manual connection step, so the body is read
      -- all in one go, including any trailers, and the connection closed or keptalive
      -- for you.

      ngx.status = res.status

      for k,v in pairs(res.headers) do
          --
      end

      ngx.say(res.body)
    ';
  }


  location /genericinterface {
    content_by_lua '

      local http = require "resty.http"
      local httpc = http.new()

      -- The generic form gives us more control. We must connect manually.
      httpc:set_timeout(500)
      httpc:connect("127.0.0.1", 80)

      -- And request using a path, rather than a full URI.
      local res, err = httpc:request{
          path = "/helloworld",
          headers = {
              ["Host"] = "example.com",
          },
      }

      if not res then
        ngx.say("failed to request: ", err)
        return
      end

      -- Now we can use the body_reader iterator, to stream the body according to our desired chunk size.
      local reader = res.body_reader

      repeat
        local chunk, err = reader(8192)
        if err then
          ngx.log(ngx.ERR, err)
          break
        end

        if chunk then
          -- process
        end
      until not chunk

      local ok, err = httpc:set_keepalive()
      if not ok then
        ngx.say("failed to set keepalive: ", err)
        return
      end
    ';
  }
}
````

# Connection

## new

`syntax: httpc = http.new()`

Creates the http object. In case of failures, returns `nil` and a string describing the error.

## connect

`syntax: ok, err = httpc:connect(host, port, options_table?)`

`syntax: ok, err = httpc:connect("unix:/path/to/unix.sock", options_table?)`

Attempts to connect to the web server.

Before actually resolving the host name and connecting to the remote backend, this method will always look up the connection pool for matched idle connections created by previous calls of this method.

An optional Lua table can be specified as the last argument to this method to specify various connect options:

* `pool`
: Specifies a custom name for the connection pool being used. If omitted, then the connection pool name will be generated from the string template `<host>:<port>` or `<unix-socket-path>`.

## set_timeout

`syntax: httpc:set_timeout(time)`

Sets the timeout (in ms) protection for subsequent operations, including the `connect` method.

## ssl_handshake

`syntax: session, err = httpc:ssl_handshake(session, host, verify)`

Performs an SSL handshake on the TCP connection, only availble in ngx_lua > v0.9.11

See docs for [ngx.socket.tcp](https://github.com/openresty/lua-nginx-module#ngxsockettcp) for details.

## set_keepalive

`syntax: ok, err = httpc:set_keepalive(max_idle_timeout, pool_size)`

Attempts to puts the current connection into the ngx_lua cosocket connection pool.

You can specify the max idle timeout (in ms) when the connection is in the pool and the maximal size of the pool every nginx worker process.

Only call this method in the place you would have called the `close` method instead. Calling this method will immediately turn the current http object into the `closed` state. Any subsequent operations other than `connect()` on the current objet will return the `closed` error.

Note that calling this instead of `close` is "safe" in that it will conditionally close depending on the type of request. Specifically, a `1.0` request without `Connection: Keep-Alive` will be closed, as will a `1.1` request with `Connection: Close`.

In case of success, returns `1`. In case of errors, returns `nil, err`. In the case where the conneciton is conditionally closed as described above, returns `2` and the error string `connection must be closed`.

## get_reused_times

`syntax: times, err = httpc:get_reused_times()`

This method returns the (successfully) reused times for the current connection. In case of error, it returns `nil` and a string describing the error.

If the current connection does not come from the built-in connection pool, then this method always returns `0`, that is, the connection has never been reused (yet). If the connection comes from the connection pool, then the return value is always non-zero. So this method can also be used to determine if the current connection comes from the pool.

## close

`syntax: ok, err = http:close()`

Closes the current connection and returns the status.

In case of success, returns `1`. In case of errors, returns `nil` with a string describing the error.


# Requesting

## request

`syntax: res, err = httpc:request(params)`

Returns a `res` table or `nil` and an error message.

The `params` table accepts the following fields:

* `version` The HTTP version number, currently supporting 1.0 or 1.1.
* `method` The HTTP method string.
* `path` The path string.
* `headers` A table of request headers.
* `body` The request body as a string, or an iterator function (see [get_client_body_reader](#get_client_body_reader)).
* `ssl_verify` Verify SSL cert matches hostname

When the request is successful, `res` will contain the following fields:

* `status` The status code.
* `reason` The status reason phrase.
* `headers` A table of headers. Multiple headers with the same field name will be presented as a table of values.
* `has_body` A boolean flag indicating if there is a body to be read. 
* `body_reader` An iterator function for reading the body in a streaming fashion.
* `read_body` A method to read the entire body into a string.
* `read_trailers` A method to merge any trailers underneath the headers, after reading the body.

## request_uri

`syntax: res, err = httpc:request_uri(uri, params)`

The simple interface. Options supplied in the `params` table are the same as in the generic interface, and will override components found in the uri itself.

In this mode, there is no need to connect manually first. The connection is made on your behalf, suiting cases where you simply need to grab a URI without too much hassle.

Additionally there is no ability to stream the response body in this mode. If the request is successful, `res` will contain the following fields:

* `status` The status code.
* `headers` A table of headers.
* `body` The response body as a string.


## request_pipeline

`syntax: responses, err = httpc:request_pipeline(params)`

This method works as per the [request](#request) method above, but `params` is instead a table of param tables. Each request is sent in order, and `responses` is returned as a table of response handles. For example:

```lua
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
    ngx.say(r:read_body())
  end
end
```

Due to the nature of pipelining, no responses are actually read until you attempt to use the response fields (status / headers etc). And since the responses are read off in order, you must read the entire body (and any trailers if you have them), before attempting to read the next response.

Note this doesn't preclude the use of the streaming response body reader. Responses can still be streamed, so long as the entire body is streamed before attempting to access the next response.

Be sure to test at least one field (such as status) before trying to use the others, in case a socket read error has occurred.

# Response

## res.body_reader

The `body_reader` iterator can be used to stream the response body in chunk sizes of your choosing, as follows:

````lua
local reader = res.body_reader

repeat
  local chunk, err = reader(8192)
  if err then
    ngx.log(ngx.ERR, err)
    break
  end

  if chunk then
    -- process
  end
until not chunk
````

If the reader is called with no arguments, the behaviour depends on the type of connection. If the response is encoded as chunked, then the iterator will return the chunks as they arrive. If not, it will simply return the entire body.

Note that the size provided is actually a **maximum** size. So in the chunked transfer case, you may get chunks smaller than the size you ask, as a remainder of the actual HTTP chunks.

## res:read_body

`syntax: body, err = res:read_body()`

Reads the entire body into a local string.


## res:read_trailers

`syntax: res:read_trailers()`

This merges any trailers underneath the `res.headers` table itself. Must be called after reading the body.


# Proxy

There are two convenience methods for when one simply wishes to proxy the current request to the connected upstream, and safely send it downstream to the client, as a reverse proxy. A complete example:

```lua
local http = require "resty.http"
local httpc = http.new()

httpc:set_timeout(500)
local ok, err = httpc:connect(HOST, PORT)

if not ok then
  ngx.log(ngx.ERR, err)
  return
end

httpc:set_timeout(2000)
httpc:proxy_response(httpc:proxy_request())
httpc:set_keepalive()
```


## proxy_request

`syntax: local res, err = httpc:proxy_request(request_body_chunk_size?)`

Performs a request using the current client request arguments, effectively proxying to the connected upstream. The request body will be read in a streaming fashion, according to `request_body_chunk_size` (see [documentation on the client body reader](#get_client_body_reader) below).


## proxy_response

`syntax: httpc:proxy_response(res, chunksize?)`

Sets the current response based on the given `res`. Ensures that hop-by-hop headers are not sent downstream, and will read the response according to `chunksize` (see [documentation on the body reader](#resbody_reader) above).


# Utility

## parse_uri

`syntax: local scheme, host, port, path = unpack(httpc:parse_uri(uri))`

This is a convenience function allowing one to more easily use the generic interface, when the input data is a URI. 


## get_client_body_reader

`syntax: reader, err = httpc:get_client_body_reader(chunksize?, sock?)`

Returns an iterator function which can be used to read the downstream client request body in a streaming fashion. You may also specify an optional default chunksize (default is `65536`), or an already established socket in
place of the client request.

Example:

```lua
local req_reader = httpc:get_client_body_reader()

repeat
  local chunk, err = req_reader(8192)
  if err then
    ngx.log(ngx.ERR, err)
    break
  end

  if chunk then
    -- process
  end
until not chunk
```

This iterator can also be used as the value for the body field in request params, allowing one to stream the request body into a proxied upstream request.

```lua
local client_body_reader, err = httpc:get_client_body_reader()

local res, err = httpc:request{
   path = "/helloworld",
   body = client_body_reader,
}
```

If `sock` is specified, 

# Author

James Hurst <james@pintsized.co.uk>

Originally started life based on https://github.com/bakins/lua-resty-http-simple. Cosocket docs and implementation borrowed from the other lua-resty-* cosocket modules.


# Licence

This module is licensed under the 2-clause BSD license.

Copyright (c) 2013-2016, James Hurst <james@pintsized.co.uk>

All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:

* Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.

* Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
