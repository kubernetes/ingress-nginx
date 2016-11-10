package = "lua-resty-http"
version = "0.09-0"
source = {
   url = "git://github.com/pintsized/lua-resty-http",
   tag = "v0.09"
}
description = {
   summary = "Lua HTTP client cosocket driver for OpenResty / ngx_lua.",
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
