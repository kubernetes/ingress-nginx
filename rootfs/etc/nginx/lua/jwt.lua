local string_len = string.len
local string_sub = string.sub
local string_find = string.find
local table_insert = table.insert
local table_concat = table.concat
local jwt = require "resty.jwt"
local cjson = require "cjson"

local _M = {}

local function find_jwk (jwks, kid)
    for index, value in ipairs(jwks) do
        if value.kid == kid then
            return value
        end
    end

    return nil
end

local function cert_to_pem(x5c)
    local result = {}
    local index = 0
    local x5c_length = string_len(x5c)

    while index < x5c_length do
        local ni = math.min(index + 64, x5c_length)
        table_insert(result, string_sub(x5c, index, ni))
        index = 1 + ni
    end

    local cert = table_concat(result, "\n")
    return "-----BEGIN CERTIFICATE-----\n" .. cert .. "\n-----END CERTIFICATE-----\n"
end

local function validate_headers_and_load(token_type)
   local auth_h = ngx.var.http_Authorization

    if auth_h == nil then
        return false, "Authorization header missing", nil
    end

    local _, _, jwt_token = string_find(auth_h, token_type .. "%s+(.+)")

    if jwt_token == nil then
        return false, "JWT missing from header. token type: " .. token_type, nil
    end

    local jwt_obj = jwt:load_jwt(jwt_token)

    if not jwt_obj.valid then
       return false, "JWT malformed or invalid: " .. jwt_obj.reason, nil
    end

    return true, nil, jwt_obj
end

local function cert_of_kid(jwks, kid)
   if kid == nil then
      return false, "Kid missing", nil
   end

   if jwks.keys == nil then
      return false, "JWKS missing `keys` property", nil
   end

   local jwk = find_jwk(jwks.keys, kid)

   if jwk == nil then
      return false, "JWK for kid: " .. kid .. "not found in JWKS", nil
   end

   local success, pem = pcall(cert_to_pem, jwk.x5c[1])

   if not success then
      return false, "Failed to turn certificate into PEM", nil
   end

   return true, nil, pem
end

function _M.jwk_verify(jwks_json, token_type)
  local token_type = token_type or "Bearer"
  local jwks = cjson.decode(jwks_json)
  local valid, reason, jwt_obj = validate_headers_and_load(token_type)

  if not valid then
    return false, reason
  end

  local c_valid, c_reason, pem = cert_of_kid(jwks, jwt_obj.header.kid)

  if not c_valid then
    return false, reason
  end

  local verified_obj = jwt:verify_jwt_obj(pem, jwt_obj)

  if not verified_obj.verified then
    return false, "JWT Verification failed " .. verified_obj.reason
  end
  return true, nil
end

return _M
