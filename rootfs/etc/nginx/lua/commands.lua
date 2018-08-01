local cjson = require("cjson")
local ngx_re = require("ngx.re")

local _M = {}

local function fetch_request_body()
  ngx.req.read_body()
  local body = ngx.req.get_body_data()

  return body
end

function _M.call()
  if ngx.var.request_method ~= "POST" then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.print("Only POST and requests are allowed!")
    return
  end

  if ngx.var.request_uri ~= "/execute" then
    ngx.status = ngx.HTTP_NOT_FOUND
    ngx.print("Not found!")
    return
  end

  local json_text = fetch_request_body()
  if not json_text then
    ngx.log(ngx.ERR, "dynamic-configuration: unable to read valid request body")
    ngx.status = ngx.HTTP_BAD_REQUEST
    return
  end

  local input = cjson.decode(body)

  if input.command == nil then
    ngx.log(ngx.ERR, "invalid JSON payload")
    ngx.status = ngx.HTTP_BAD_REQUEST
  end

  local prog = require "resty.exec".new("/var/lib/shared/nginx/exec.sock")
  prog.argv = ngx_re.split(input.command, " ")
  local data, err = prog()
  if (err) then
    local res = {
      error = err
    }

    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.print(cjson.encode(res))
  end

  ngx.print(cjson.encode(data))
  ngx.status = ngx.HTTP_OK
end

return _M
