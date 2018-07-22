_G._TEST = true
local cjson = require("cjson")

describe("Monitor", function()
    local monitor = require("monitor")
    describe("encode_nginx_stats()", function()
        it("successfuly encodes the current stats of nginx to JSON", function()
            local nginx_environment = {
                host = "testshop.com",
                status = "200",
                bytes_sent = "150",
                server_protocol = "HTTP",
                request_method = "GET",
                location_path = "/admin",
                request_length = "300",
                request_time = "210",
                proxy_upstream_name = "test-upstream",
                upstream_addr = "2.2.2.2",
                upstream_response_time = "200",
                upstream_response_length = "150",
                upstream_connect_time = "1",
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
                responseLength = 150.0,
                method = "GET",
                path = "/admin",
                requestLength = 300.0,
                requestTime = 210.0,
                endpoint = "2.2.2.2",
                upstreamResponseTime = 200,
                upstreamStatus = "220",
                upstreamLatency = 1.0,
                upstreamResponseLength = 150.0,
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
                location_path = "/admin",
                request_time = "202",
                proxy_upstream_name = "test-upstream",
                upstream_addr = "2.2.2.2",
                upstream_response_time = "201",
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
                responseLength = -1,
                method = "GET",
                path = "/admin",
                requestLength = -1,
                requestTime = 202.0,
                endpoint = "2.2.2.2",
                upstreamStatus = "220",
                namespace = "-",
                ingress = "web-yml",
                upstreamLatency = -1,
                upstreamResponseTime = 201,
                upstreamResponseLength = -1,
                responseLength = -1,
                service = "-",
            }
            assert.are.same(decoded_json_stats,expected_json_stats)
        end)
    end)
end)
