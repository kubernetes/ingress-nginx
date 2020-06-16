local cjson = require("cjson.safe")

local io = io
local ngx = ngx
local tostring = tostring
local string = string
local table = table
local pairs = pairs

-- this is the Lua representation of Configuration struct in internal/ingress/types.go
local configuration_data = ngx.shared.configuration_data
local certificate_data = ngx.shared.certificate_data
local certificate_servers = ngx.shared.certificate_servers

local EMPTY_UID = "-1"

local _M = {}

function _M.get_backends_data()
  return configuration_data:get("backends")
end

function _M.get_general_data()
  return configuration_data:get("general")
end

function _M.get_raw_backends_last_synced_at()
  local raw_backends_last_synced_at = configuration_data:get("raw_backends_last_synced_at")
  if raw_backends_last_synced_at == nil then
    raw_backends_last_synced_at = 1
  end
  return raw_backends_last_synced_at
end

local function fetch_request_body()
  ngx.req.read_body()
  local body = ngx.req.get_body_data()

  if not body then
    -- request body might've been written to tmp file if body > client_body_buffer_size
    local file_name = ngx.req.get_body_file()
    local file = io.open(file_name, "rb")

    if not file then
      return nil
    end

    body = file:read("*all")
    file:close()
  end

  return body
end

local function get_pem_cert(hostname)
  local uid = certificate_servers:get(hostname)
  if not uid then
    return nil
  end

  return certificate_data:get(uid)
end

local function handle_servers()
  if ngx.var.request_method ~= "POST" then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.print("Only POST requests are allowed!")
    return
  end

  local raw_configuration = fetch_request_body()

  local configuration, err = cjson.decode(raw_configuration)
  if not configuration then
    ngx.log(ngx.ERR, "could not parse configuration: ", err)
    ngx.status = ngx.HTTP_BAD_REQUEST
    return
  end

  local err_buf = {}

  for server, uid in pairs(configuration.servers) do
    if uid == EMPTY_UID then
      -- notice that we do not delete certificate corresponding to this server
      -- this is becase a certificate can be used by multiple servers/hostnames
      certificate_servers:delete(server)
    else
      local success, set_err, forcible = certificate_servers:set(server, uid)
      if not success then
        local err_msg = string.format("error setting certificate for %s: %s\n",
          server, tostring(set_err))
        table.insert(err_buf, err_msg)
      end
      if forcible then
        local msg = string.format("certificate_servers dictionary is full, "
          .. "LRU entry has been removed to store %s", server)
        ngx.log(ngx.WARN, msg)
      end
    end
  end

  for uid, cert in pairs(configuration.certificates) do
    local success, set_err, forcible = certificate_data:set(uid, cert)
    if not success then
      local err_msg = string.format("error setting certificate for %s: %s\n",
        uid, tostring(set_err))
      table.insert(err_buf, err_msg)
    end
    if forcible then
      local msg = string.format("certificate_data dictionary is full, "
        .. "LRU entry has been removed to store %s", uid)
      ngx.log(ngx.WARN, msg)
    end
  end

  if #err_buf > 0 then
    ngx.log(ngx.ERR, table.concat(err_buf))
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    return
  end

  ngx.status = ngx.HTTP_CREATED
end

local function handle_general()
  if ngx.var.request_method == "GET" then
    ngx.status = ngx.HTTP_OK
    ngx.print(_M.get_general_data())
    return
  end

  if ngx.var.request_method ~= "POST" then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.print("Only POST and GET requests are allowed!")
    return
  end

  local config = fetch_request_body()

  local success, err = configuration_data:safe_set("general", config)
  if not success then
    ngx.status = ngx.HTTP_INTERNAL_SERVER_ERROR
    ngx.log(ngx.ERR, "error setting general config: " .. tostring(err))
    return
  end

  ngx.status = ngx.HTTP_CREATED
end

local function handle_certs()
  if ngx.var.request_method ~= "GET" then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.print("Only GET requests are allowed!")
    return
  end

  local query = ngx.req.get_uri_args()
  if not query["hostname"] then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.print("Hostname must be specified.")
    return
  end

  local key = get_pem_cert(query["hostname"])
  if key then
    ngx.status = ngx.HTTP_OK
    ngx.print(key)
    return
  else
    ngx.status = ngx.HTTP_NOT_FOUND
    ngx.print("No key associated with this hostname.")
    return
  end
end


local function handle_backends()
  if ngx.var.request_method == "GET" then
    ngx.status = ngx.HTTP_OK
    ngx.print(_M.get_backends_data())
    return
  end

  local backends = fetch_request_body()
  if not backends then
    ngx.log(ngx.ERR, "dynamic-configuration: unable to read valid request body")
    ngx.status = ngx.HTTP_BAD_REQUEST
    return
  end

  local success, err = configuration_data:set("backends", backends)
  if not success then
    ngx.log(ngx.ERR, "dynamic-configuration: error updating configuration: " .. tostring(err))
    ngx.status = ngx.HTTP_BAD_REQUEST
    return
  end

  ngx.update_time()
  local raw_backends_last_synced_at = ngx.time()
  success, err = configuration_data:set("raw_backends_last_synced_at", raw_backends_last_synced_at)
  if not success then
    ngx.log(ngx.ERR, "dynamic-configuration: error updating when backends sync, " ..
                     "new upstream peers waiting for force syncing: " .. tostring(err))
    ngx.status = ngx.HTTP_BAD_REQUEST
    return
  end

  ngx.status = ngx.HTTP_CREATED
end

function _M.call()
  if ngx.var.request_method ~= "POST" and ngx.var.request_method ~= "GET" then
    ngx.status = ngx.HTTP_BAD_REQUEST
    ngx.print("Only POST and GET requests are allowed!")
    return
  end

  if ngx.var.request_uri == "/configuration/servers" then
    handle_servers()
    return
  end

  if ngx.var.request_uri == "/configuration/general" then
    handle_general()
    return
  end

  if ngx.var.uri == "/configuration/certs" then
    handle_certs()
    return
  end

  if ngx.var.request_uri == "/configuration/backends" then
    handle_backends()
    return
  end

  ngx.status = ngx.HTTP_NOT_FOUND
  ngx.print("Not found!")
end

setmetatable(_M, {__index = { handle_servers = handle_servers }})

return _M
