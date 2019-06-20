local _M = {}

function _M.rewrite()
  ngx.req.set_header("x-hello-world", "1")
end

function _M.access()
  local ua = ngx.var.http_user_agent

  if ua == "hello" then
    ngx.exit(403)
  end
end

return _M
