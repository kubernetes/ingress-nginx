-- Rendered in /etc/nginx/lua/stratio_auth.lua
local ngx = ngx
local http = require "resty.http"
local ck = require "resty.cookie"
local cjson = require "cjson.safe"

local _M = {}

local function create_jwt(oauth2_cookie, userinfo_url, signing_key)

    -- Get User info
    local httpc = http.new()
    local res, err = httpc:request_uri(userinfo_url, {
        method = "GET",
        ssl_verify = false,
        headers = {
        ["Content-Type"] = "application/json",
        ["Cookie"] = "_oauth2_proxy=" .. oauth2_cookie,
        }
    })

    if not res then
        ngx.log(ngx.STDERR, 'Unexpected error obtaining the user information: ', err)
        return 403
    end

    local json_decoder = cjson.decode
    local userinfo, err = json_decoder(res.body)

    -- Create JWT
    local jwt = require "resty.jwt"
    local stratio_jwt = jwt:sign(
        signing_key,
        {
            header = {
                alg="RS256",
                kid="secret",
                typ="JWT"
            },
            payload = {
                iss = "stratio-ingress",
                nbf = os.time(),
                exp = os.time() + 21600,
                cn = userinfo["preferredUsername"],
                groups = userinfo["groups"],
                mail = userinfo["email"],
                tenant = userinfo["tenant"],
                uid = userinfo["preferredUsername"]
            }
        }
    )
    return stratio_jwt
end

local function create_jwt_from_cert(signing_key)

    -- Get User info
    local cert_cn = string.match(ngx.var.ssl_client_s_dn, "CN=([^,]+)")

    -- Create JWT
    local jwt = require "resty.jwt"
    local stratio_jwt = jwt:sign(
        signing_key,
        {
            header = {
                alg="RS256",
                kid="secret",
                typ="JWT"
            },
            payload = {
                iss = "stratio-ingress",
                nbf = os.time(),
                exp = os.time() + 21600,
                cn = cert_cn,
                uid = cert_cn
            }
        }
    )
    return stratio_jwt
end

function _M.create_cookie(userinfo_url, oauth2_cookie_name, stratio_cookie_name, signing_key, verification_key)

    local stratio_jwt = ""

    -- If there's a cert in request (prevoiusly validated by nginx), create its JWT + cookie
    if ngx.var.ssl_client_s_dn then
        -- Create a JWT from the cert's info
        stratio_jwt = create_jwt_from_cert(signing_key)

        -- Add cookie to request
        if ngx.var.http_cookie then
          ngx.req.set_header("Cookie", stratio_cookie_name .. "=" .. stratio_jwt .. ";" .. ngx.var.http_cookie);
        else
          ngx.req.set_header("Cookie", stratio_cookie_name .. "=" .. stratio_jwt);
        end
        return
    end
    
    -- Get request's cookies
    local req_cookie, err = ck:new()
    if not req_cookie then
        ngx.log(ngx.STDERR, 'Unexpected error obtaining the request cookies: ', err)
        ngx.status = ngx.HTTP_UNAUTHORIZED
        ngx.exit(ngx.HTTP_UNAUTHORIZED)
        return
    end

    -- Get oauth2-proxy cookie
    local oauth2_cookie, err = req_cookie:get(oauth2_cookie_name)
    if not oauth2_cookie then
        ngx.log(ngx.STDERR, 'Unexpected error obtaining the _oauth2_proxy cookie')
        ngx.status = ngx.HTTP_UNAUTHORIZED
        ngx.exit(ngx.HTTP_UNAUTHORIZED)
        return
    end

    -- If there's no Stratio cookie in the request, add it
    local stratio_cookie, err = req_cookie:get(stratio_cookie_name)
    if not stratio_cookie then

        ngx.log(ngx.DEBUG, 'Cookie not found in request')
        stratio_jwt = create_jwt(oauth2_cookie, userinfo_url, signing_key)

        ngx.log(ngx.DEBUG, 'Cookie created, adding to response')
        local ok, err = req_cookie:set({
            key = stratio_cookie_name,
            value = stratio_jwt,
            path = "/",
            secure = true,
            httponly = true,
            samesite = "Lax",
            expires = ngx.cookie_time(ngx.time() + 21600),
            max_age = 21600
        })

        if not ok then
            ngx.log(ngx.STDERR, 'Unexpected error setting the Stratio cookie: ', err)
            return 403
        end

        ngx.log(ngx.DEBUG, 'Adding cookie to request')
        ngx.req.set_header("Cookie", stratio_cookie_name .. "=" .. stratio_jwt .. ";" .. ngx.var.http_cookie);
    else

        ngx.log(ngx.DEBUG, 'Cookie found in request, verifying signature, expiration and issuer')
        local jwt = require "resty.jwt"
        local jwt_obj = jwt:verify(verification_key, stratio_cookie, {
            lifetime_grace_period = 120,
            require_exp_claim = true,
            valid_issuers = { "stratio-ingress" }
        })

        if not jwt_obj["verified"] then
            
            ngx.log(ngx.DEBUG, 'Invalid JWT, generating a new one')
            stratio_jwt = create_jwt(oauth2_cookie, userinfo_url, signing_key)
            
            ngx.log(ngx.DEBUG, 'Cookie created, adding to response')
            local ok, err = req_cookie:set({
                key = stratio_cookie_name,
                value = stratio_jwt,
                path = "/",
                secure = true,
                httponly = true,
                samesite = "Lax",
                expires = ngx.cookie_time(ngx.time() + 21600),
                max_age = 21600
            })
    
            if not ok then
                ngx.log(ngx.STDERR, 'Unexpected error setting the Stratio cookie: ', err)
                return 401
            end

            local cookies, err = req_cookie:get_all()
            if not cookies then
                ngx.log(ngx.STDERR, 'Unexpected error getting the request cookies: ', err)
                return
            end

            local mycookiestr = ''
            for k, v in pairs(cookies) do
                if k == stratio_cookie_name then
                    v = stratio_jwt
                end
                mycookiestr = mycookiestr .. k .. "=" .. v .. ";"
            end
            
            ngx.log(ngx.DEBUG, 'Adding cookie to request')
            ngx.req.set_header("Cookie", stratio_cookie_name .. "=" .. stratio_jwt .. ";" .. mycookiestr);
        else
            ngx.log(ngx.DEBUG, 'Valid JWT found, moving on..')
            return
        end
    end
end

return _M
