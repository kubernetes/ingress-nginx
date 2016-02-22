local ssl = require "ngx.ssl"
local ssl_base_directory = "/etc/nginx/nginx-ssl"

local server_name = ssl.server_name()
local addr, addrtyp, err = ssl.raw_server_addr()
local byte = string.byte
local cert_path = ""

ssl.clear_certs()

-- Check for SNI request.
if server_name == nil then
    ngx.log(ngx.INFO, "SNI Not present - performing IP lookup")
    -- Set server name as IP address.
    server_name = string.format("%d.%d.%d.%d", byte(addr, 1), byte(addr, 2), byte(addr, 3), byte(addr, 4))
    ngx.log(ngx.INFO, "IP Address: ", server_name)
end 

-- Set certifcate paths
cert_path = ssl_base_directory .. "/" .. server_name .. ".cert"
key_path = ssl_base_directory .. "/" .. server_name .. ".key"

-- Attempt to retrieve and set certificate for request.
local f = assert(io.open(cert_path))
local cert_data = f:read("*a")
f:close()

local ok, err = ssl.set_der_cert(cert_data)
if not ok then
    ngx.log(ngx.ERR, "failed to set DER cert: ", err)
    return
end

-- Attempt to retrieve and set key for request.
local f = assert(io.open(key_path))
local pkey_data = f:read("*a")
f:close()

local ok, err = ssl.set_der_priv_key(pkey_data)

if not ok then
    ngx.log(ngx.ERR, "failed to set DER key: ", err)
    return
end
