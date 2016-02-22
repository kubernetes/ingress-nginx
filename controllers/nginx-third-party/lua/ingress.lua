local _M = {}

local cjson = require "cjson"
local trie = require "trie"
local http = require "resty.http"
local cache = require "resty.dns.cache"
local os = require "os"

local encode = cjson.encode
local decode = cjson.decode

local table_concat = table.concat

local trie_get = trie.get
local match = string.match
local gsub = string.gsub
local lower = string.lower


-- we "cache" the config local to each worker
local ingressConfig = nil

local cluster_domain = "cluster.local"

local def_backend = nil

local custom_error = nil

local dns_cache_options = nil

function get_ingressConfig(ngx)
    local d = ngx.shared["ingress"]
    local value, flags, stale = d:get_stale("ingressConfig")
    if not value then
        -- nothing we can do
        return nil, "config not set"
    end
    ingressConfig = decode(value)
    return ingressConfig, nil
end

function worker_cache_config(ngx)
    local _, err = get_ingressConfig(ngx)
    if err then
        ngx.log(ngx.ERR, "unable to get ingressConfig: ", err)
        return
    end
end

function _M.content(ngx)
    local host = ngx.var.host

    -- strip off any port
    local h = match(host, "^(.+):?")
    if h then
        host = h
    end

    host = lower(host)

    local config, err = get_ingressConfig(ngx)
    if err then
        ngx.log(ngx.ERR, "unable to get config: ", err)
        return ngx.exit(503)
    end

    -- this assumes we only allow exact host matches
    local paths = config[host]
    if not paths then
        ngx.log(ngx.ERR, "No server for host "..host.." returning 404")
        if custom_error then
            openCustomErrorURL(404, custom_error)
            return
        else 
            openURL(404, def_backend)
            return
        end
    end

    local backend = trie_get(paths, ngx.var.uri)

    if not backend then
        ngx.log(ngx.ERR, "No server for host "..host.." and path "..ngx.var.uri.." returning 404")
        if custom_error then
            openCustomErrorURL(404, custom_error)
            return
        else 
            openURL(404, def_backend)
            return
        end
    end

    local address = backend.host
    ngx.var.upstream_port = backend.port or 80

    if dns_cache_options then
        local dns = cache.new(dns_cache_options)
        local answer, err, stale = dns:query(address, { qtype = 1 })
        if err or (not answer) then
            if stale then
                answer = stale
            else
                answer = nil
            end
        end
        if answer and answer[1] then
            local ans = answer[1]
            if ans.address then
                address = ans.address
            end
        else
            ngx.log(ngx.ERR, "dns failed for ", address, " with ", err, " => ", encode(answer or ""))
        end
    end

    ngx.var.upstream_host = address
    return
end

function _M.init_worker(ngx)
end

function _M.init(ngx, options)
    -- ngx.log(ngx.OK, "options: "..encode(options))
    def_backend = options.def_backend
    custom_error = options.custom_error

    -- try to create a dns cache
    local resolvers = options.resolvers
    if resolvers then
        cache.init_cache(512)
        local servers = trie.strsplit(" ", resolvers)
        dns_cache_options =
            {
                dict = "dns_cache",
                negative_ttl = nil,
                max_stale = 900,
                normalise_ttl = false,
                resolver  = {
                    nameservers = {servers[1]}
                }
            }
    end

end

-- dump config. This is the raw config (including trie) for now
function _M.config(ngx)
    ngx.header.content_type = "application/json"
    local config = {
        ingress = ingressConfig
    }
    local val = encode(config)
    ngx.print(val)
end

function _M.update_ingress(ngx)
    ngx.header.content_type = "application/json"

    if ngx.req.get_method() ~= "POST" then
        ngx.print(encode({
            message = "only POST request"
        }))
        ngx.exit(400)
        return
    end

    ngx.req.read_body()
    local data = ngx.req.get_body_data()
    local val = decode(data)

    if not val then
        ngx.log(ngx.ERR, "failed to decode body")
        return
    end

    config = {}

    for _, ingress in ipairs(val) do
        local namespace = ingress.metadata.namespace

        local spec = ingress.spec
        -- we do not allow default ingress backends right now.
        for _, rule in ipairs(spec.rules) do
            local host = rule.host
            local paths = config[host]
            if not paths then
                paths = trie.new()
                config[host] = paths
            end
            rule.http = rule.http or { paths = {}}
            for _, path in ipairs(rule.http.paths) do
                local hostname = table_concat(
                    {
                        path.backend.serviceName,
                        namespace,
                        "svc",
                        cluster_domain
                    }, ".")
                local backend = {
                    host = hostname,
                    port = path.backend.servicePort
                }

                paths:add(path.path, backend)
            end
        end
    end

    local d = ngx.shared["ingress"]
    local ok, err, _ = d:set("ingressConfig", encode(config))
    if not ok then
        ngx.log(ngx.ERR, "Error: "..err)
        local res = encode({
            message = "Error updating Ingress rules: "..err
        })
        ngx.print(res)
        return ngx.exit(500)
    end

    ingressConfig = config

    local res = encode({
        message = "Ingress rules updated"
    })
    ngx.print(res)
end

return _M
