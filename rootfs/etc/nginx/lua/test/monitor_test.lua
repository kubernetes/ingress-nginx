package.path = "./rootfs/etc/nginx/lua/?.lua;./rootfs/etc/nginx/lua/test/mocks/?.lua;" .. package.path
_G._TEST = true
local cjson = require('cjson')

local function udp_mock()
    return {
        setpeername = function(...) return true end,
        send = function(payload) return payload end,
        close = function(...) return true end
    }
end

local _ngx = {
    shared = {},
    log = function(...) end,
    socket = {
        udp = udp_mock
    },
    get_phase = function() return "timer" end,
    var = {}
}
_G.ngx = _ngx

describe("Monitor", function()
    local monitor = require("monitor")
    describe("encode_nginx_stats()", function()
        it("successfuly encodes the current stats of nginx to JSON", function()
            local nginx_environment = {
                host = "testshop.com",
                status = "200",
                remote_addr = "10.10.10.10",
                realip_remote_addr = "5.5.5.5",
                remote_user = "admin",
                bytes_sent = "150",
                server_protocol = "HTTP",
                request_method = "GET",
                uri = "/admin",
                request_length = "300",
                request_time = "60",
                proxy_upstream_name = "test-upstream",
                upstream_addr = "2.2.2.2",
                upstream_response_time = "200",
                upstream_status = "220",
                namespace = "test-app-production",
                ingress_name = "web-yml",
                service_name = "test-app",
            }
            ngx.var = nginx_environment

            local encode_nginx_stats = monitor.encode_nginx_stats
            local encoded_json_stats = encode_nginx_stats()
            local decoded_json_stats = cjson.decode(encoded_json_stats)

            local expected_json_stats = {
                host = "testshop.com",
                status = "200",
                remoteAddr = "10.10.10.10",
                realIpAddr = "5.5.5.5",
                remoteUser = "admin",
                bytesSent = 150.0,
                protocol = "HTTP",
                method = "GET",
                uri = "/admin",
                requestLength = 300.0,
                requestTime = 60.0,
                upstreamName = "test-upstream",
                upstreamIP = "2.2.2.2",
                upstreamResponseTime = 200,
                upstreamStatus = "220",
                namespace = "test-app-production",
                ingress = "web-yml",
                service = "test-app",
            }

            assert.are.same(decoded_json_stats,expected_json_stats)
        end)

        it("replaces empty numeric keys with -1 and missing string keys with -", function()
            local nginx_environment = {
                remote_addr = "10.10.10.10",
                realip_remote_addr = "5.5.5.5",
                remote_user = "francisco",
                server_protocol = "HTTP",
                request_method = "GET",
                uri = "/admin",
                request_time = "60",
                proxy_upstream_name = "test-upstream",
                upstream_addr = "2.2.2.2",
                upstream_response_time = "200",
                upstream_status = "220",
                ingress_name = "web-yml",
            }
            ngx.var = nginx_environment

            local encode_nginx_stats = monitor.encode_nginx_stats
            local encoded_json_stats = encode_nginx_stats()
            local decoded_json_stats = cjson.decode(encoded_json_stats)

            local expected_json_stats = {
                host = "-",
                status = "-",
                remoteAddr = "10.10.10.10",
                realIpAddr = "5.5.5.5",
                remoteUser = "francisco",
                bytesSent = -1,
                protocol = "HTTP",
                method = "GET",
                uri = "/admin",
                requestLength = -1,
                requestTime = 60.0,
                upstreamName = "test-upstream",
                upstreamIP = "2.2.2.2",
                upstreamResponseTime = 200,
                upstreamStatus = "220",
                namespace = "-",
                ingress = "web-yml",
                service = "-",
            }
            assert.are.same(decoded_json_stats,expected_json_stats)
        end)
    end)
end)
