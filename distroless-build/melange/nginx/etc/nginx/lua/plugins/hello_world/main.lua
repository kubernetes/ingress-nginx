local ngx = ngx

local _M = {}

function _M.rewrite()
  local ua = ngx.var.http_user_agent

  if ua == "hello" then
    ngx.req.set_header("x-hello-world", "1")
  end
end

return _M
