std = 'ngx_lua'
globals = {
  '_TEST'
}
exclude_files = {'./rootfs/etc/nginx/lua/test/**/*.lua'}
files["rootfs/etc/nginx/lua/lua_ingress.lua"] = {
  ignore = { "122" },
  -- TODO(elvinefendi) figure out why this does not work
  --read_globals = {"math.randomseed"},
}
