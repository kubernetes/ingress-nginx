local _M = {}

function _M.purge(cache_dir, namespace, name)
  local dir = cache_dir .. "/" .. namespace .. "/" .. name
  local cmd = "/usr/bin/find " .. dir .. " -mindepth 1 -maxdepth 1 | /usr/bin/xargs rm -rf"

  os.execute(cmd)

  ngx.say("OK " .. namespace .. "/" .. name)
  ngx.exit(ngx.OK)
end

return _M
