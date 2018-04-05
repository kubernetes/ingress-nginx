-- this is the Lua representation of Configuration struct in internal/ingress/types.go
local configuration_data = ngx.shared.configuration_data

local _M = {}

function _M.get_backends_data()
  return configuration_data:get("backends")
end

function _M.call()
  if ngx.var.request_method ~= "POST" and ngx.var.request_method ~= "GET" then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.print("Only POST and GET requests are allowed!")
    return
  end

  if ngx.var.request_uri ~= "/configuration/backends" then
    ngx.status = ngx.HTTP_NOT_FOUND
    ngx.print("Not found!")
    return
  end

  if ngx.var.request_method == "GET" then
    ngx.status = ngx.HTTP_OK
    ngx.print(_M.get_backends_data())
    return
  end

  ngx.req.read_body()
  local backends = ngx.req.get_body_data()
  if not backends then
    -- response might be have been written to tmp file if size(body) > client_body_buffer_size
    backends = ngx.req.get_body_file()
  end
  if not backends then
    -- no current backends data found
    ngx.log(ngx.ERR, "configuration/backends: empty response body")
    ngx.status = ngx.HTTP_BAD_REQUEST
    return
  end

  local success, err = configuration_data:set("backends", ngx.req.get_body_data())
  if not success then
    ngx.log(ngx.ERR, "configuration/backends: error updating configuration: " .. tostring(err))
    ngx.status = ngx.HTTP_BAD_REQUEST
    return
  end

  ngx.status = ngx.HTTP_CREATED
end

return _M
