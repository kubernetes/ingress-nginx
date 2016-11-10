package = "lua-resty-http"
version = "0.07-0"
source = {
   url = "git://github.com/pintsized/lua-resty-http",
   tag = "v0.07"
}
description = {
   summary = "Lua HTTP client cosocket driver for OpenResty / ngx_lua.",
   detailed = [[
    Features an HTTP 1.0 and 1.1 streaming interface to reading 
    bodies using coroutines, for predictable memory usage in Lua 
    land. Alternative simple interface for singleshot requests 
    without manual connection step. Supports chunked transfer 
    encoding, keepalive, pipelining, and trailers. Headers are 
    treated case insensitively. Probably production ready in most
    cases, though not yet proven in the wild.
    Recommended by the OpenResty maintainer as a long-term 
    replacement for internal requests through ngx.location.capture.
  ]],
   homepage = "https://github.com/pintsized/lua-resty-http",
   license = "2-clause BSD",
   maintainer = "James Hurst <james@pintsized.co.uk>"
}
dependencies = {
   "lua >= 5.1"
}
build = {
   type = "builtin",
   modules = {
      ["resty.http"] = "lib/resty/http.lua",
      ["resty.http_headers"] = "lib/resty/http_headers.lua"
   }
}
