local json = require "json"
local ssl = require "ngx.ssl"

local shared_memory = ngx.shared.shared_memory;

local http_host, err = ssl.server_name()
if http_host ~= nil then
    local vhosts_json = shared_memory:get("VHOSTS")
    local vhosts = json.decode(json.decode(vhosts_json))

    local server = vhosts.servers[http_host]
    if (server == nil) then
        server = vhosts.servers["_"]
        if (server == nil) then
            ngx.status = 503
            ngx.exit(ngx.status)
        end
    end
    if server.sslcertificate ~= "" then
        local ok, err = ssl.clear_certs()
        if not ok then
            ngx.log(ngx.ERR, "SSL ["..http_host.."]: failed to clear fallback certificates")
            return ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
        end

        local cert_key_data = server.sslcertificatereal

        if cert_key_data == nil then
            ngx.log(ngx.ERR, "SSL certificate not found in memory")
            return ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
        end

        local pem_cert_chain = string.match(cert_key_data, "%-*BEGIN CERTIFICATE.-END CERTIFICATE%-*")

        local der_cert_chain, err = ssl.cert_pem_to_der(pem_cert_chain)
        if not der_cert_chain then
            ngx.log(ngx.ERR, "failed to convert certificate chain ",
                    "from PEM to DER: ", err)
            return ngx.exit(ngx.ERROR)
        end

        local ok, err = ssl.set_der_cert(der_cert_chain)
        if not ok then
            ngx.log(ngx.ERR, "failed to set DER cert: ", err)
            return ngx.exit(ngx.ERROR)
        end

        local pem_pkey = string.match(cert_key_data, "%-*BEGIN RSA PRIVATE KEY.-END RSA PRIVATE KEY%-*")

        local der_pkey, err = ssl.priv_key_pem_to_der(pem_pkey)
        if not der_pkey then
            ngx.log(ngx.ERR, "failed to convert private key ",
                    "from PEM to DER: ", err)
            return ngx.exit(ngx.ERROR)
        end

        local ok, err = ssl.set_der_priv_key(der_pkey)
        if not ok then
            ngx.log(ngx.ERR, "failed to set DER private key: ", err)
            return ngx.exit(ngx.ERROR)
        end

    end
else
    ngx.log(ngx.ERR, "No SNI not provided from client")
end

