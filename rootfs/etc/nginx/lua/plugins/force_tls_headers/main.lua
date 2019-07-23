local _M = {}

-- To be used behind an ELB terminating TLS
function _M.rewrite()
  ngx.var.pass_access_scheme = "https"
  ngx.var.pass_port = 443
end

return _M
