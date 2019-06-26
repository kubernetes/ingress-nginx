_G._TEST = true
local cjson = require("cjson")
local configuration = require("configuration")

local unmocked_ngx = _G.ngx
local certificate_data = ngx.shared.certificate_data

function get_backends()
    return {
        {
            name = "my-dummy-backend-1", ["load-balance"] = "sticky",
            endpoints = { { address = "10.183.7.40", port = "8080", maxFails = 0, failTimeout = 0 } },
            sessionAffinityConfig = { name = "cookie", cookieSessionAffinity = { name = "route" } },
        },
        {
            name = "my-dummy-backend-2", ["load-balance"] = "ewma",
            endpoints = {
                { address = "10.184.7.40", port = "7070", maxFails = 3, failTimeout = 2 },
                { address = "10.184.7.41", port = "7070", maxFails = 2, failTimeout = 1 },
            }
        },
        {
            name = "my-dummy-backend-3", ["load-balance"] = "round_robin",
            endpoints = {
                { address = "10.185.7.40", port = "6060", maxFails = 0, failTimeout = 0 },
                { address = "10.185.7.41", port = "6060", maxFails = 2, failTimeout = 1 },
            }
        },
    }
end

function get_mocked_ngx_env()
    local _ngx = {
        status = ngx.HTTP_OK,
        var = {},
        req = {
            read_body = function() end,
            get_body_data = function() return cjson.encode(get_backends()) end,
            get_body_file = function() return nil end,
        },
        log = function(msg) end,
    }
    setmetatable(_ngx, {__index = _G.ngx})
    return _ngx
end

describe("Configuration", function()
    before_each(function()
        _G.ngx = get_mocked_ngx_env()
    end)

    after_each(function()
        _G.ngx = unmocked_ngx
        package.loaded["configuration"] = nil
        configuration = require("configuration")
    end)

    describe("Backends", function()
        context("Request method is neither GET nor POST", function()
            it("sends 'Only POST and GET requests are allowed!' in the response body", function()
                ngx.var.request_method = "PUT"
                local s = spy.on(ngx, "print")
                assert.has_no.errors(configuration.call)
                assert.spy(s).was_called_with("Only POST and GET requests are allowed!")
            end)

            it("returns a status code of 400", function()
                ngx.var.request_method = "PUT"
                assert.has_no.errors(configuration.call)
                assert.equal(ngx.status, ngx.HTTP_BAD_REQUEST)
            end)
        end)

        context("GET request to /configuration/backends", function()
            before_each(function()
                ngx.var.request_method = "GET"
                ngx.var.request_uri = "/configuration/backends"
            end)

            it("returns the current configured backends on the response body", function()
                -- Encoding backends since comparing tables fail due to reference comparison
                local encoded_backends = cjson.encode(get_backends())
                ngx.shared.configuration_data:set("backends", encoded_backends)
                local s = spy.on(ngx, "print")
                assert.has_no.errors(configuration.call)
                assert.spy(s).was_called_with(encoded_backends)
            end)

            it("returns a status of 200", function()
                assert.has_no.errors(configuration.call)
                assert.equal(ngx.status, ngx.HTTP_OK)
            end)
        end)

        context("POST request to /configuration/backends", function()
            before_each(function()
                ngx.var.request_method = "POST"
                ngx.var.request_uri = "/configuration/backends"
            end)

            it("stores the posted backends on the shared dictionary", function()
                -- Encoding backends since comparing tables fail due to reference comparison
                assert.has_no.errors(configuration.call)
                assert.equal(ngx.shared.configuration_data:get("backends"), cjson.encode(get_backends()))
            end)

            context("Failed to read request body", function()
                local mocked_get_body_data = ngx.req.get_body_data
                before_each(function()
                    ngx.req.get_body_data = function() return nil end
                end)

                teardown(function()
                    ngx.req.get_body_data = mocked_get_body_data
                end)

                it("returns a status of 400", function()
                    _G.io.open = function(filename, extension) return false end
                    assert.has_no.errors(configuration.call)
                    assert.equal(ngx.status, ngx.HTTP_BAD_REQUEST)
                end)

                it("logs 'dynamic-configuration: unable to read valid request body to stderr'", function()
                    local s = spy.on(ngx, "log")
                    assert.has_no.errors(configuration.call)
                    assert.spy(s).was_called_with(ngx.ERR, "dynamic-configuration: unable to read valid request body")
                end)
            end)

            context("Failed to set the new backends to the configuration dictionary", function()
                local resty_configuration_data_set = ngx.shared.configuration_data.set
                before_each(function()
                    ngx.shared.configuration_data.set = function(key, value) return false, "" end
                end)

                teardown(function()
                    ngx.shared.configuration_data.set = resty_configuration_data_set
                end)

                it("returns a status of 400", function()
                    assert.has_no.errors(configuration.call)
                    assert.equal(ngx.status, ngx.HTTP_BAD_REQUEST)
                end)

                it("logs 'dynamic-configuration: error updating configuration:' to stderr", function()
                    local s = spy.on(ngx, "log")
                    assert.has_no.errors(configuration.call)
                    assert.spy(s).was_called_with(ngx.ERR, "dynamic-configuration: error updating configuration: ")
                end)
            end)

            context("Succeeded to update backends configuration", function()
                it("returns a status of 201", function()
                    assert.has_no.errors(configuration.call)
                    assert.equal(ngx.status, ngx.HTTP_CREATED)
                end)
            end)
        end)
    end)

    describe("handle_servers()", function()
        it("should not accept non POST methods", function()
            ngx.var.request_method = "GET"
            
            local s = spy.on(ngx, "print")
            assert.has_no.errors(configuration.handle_servers)
            assert.spy(s).was_called_with("Only POST requests are allowed!")
            assert.same(ngx.status, ngx.HTTP_BAD_REQUEST)
        end)

        it("should ignore servers that don't have hostname or pemCertKey set", function()
            ngx.var.request_method = "POST"
            local mock_servers = cjson.encode({
                {
                    hostname = "hostname",
                    sslCert = {}
                },
                {
                    sslCert = {
                        pemCertKey = "pemCertKey"
                    }
                }
            })
            ngx.req.get_body_data = function() return mock_servers end

            local s = spy.on(ngx, "log")
            assert.has_no.errors(configuration.handle_servers)
            assert.spy(s).was_called_with(ngx.WARN, "hostname or pemCertKey are not present")
            assert.same(ngx.status, ngx.HTTP_CREATED)
        end)

        it("should successfully update certificates and keys for each host", function()
            ngx.var.request_method = "POST"
            local mock_servers = cjson.encode({
                {
                    hostname = "hostname",
                    sslCert = {
                        pemCertKey = "pemCertKey"
                    }
                }
            })
            ngx.req.get_body_data = function() return mock_servers end

            assert.has_no.errors(configuration.handle_servers)
            assert.same(certificate_data:get("hostname"), "pemCertKey")
            assert.same(ngx.status, ngx.HTTP_CREATED)
        end)

        it("should log an err and set status to Internal Server Error when a certificate cannot be set", function()
            ngx.var.request_method = "POST"
            ngx.shared.certificate_data.set = function(self, data) return false, "error", nil end
            local mock_servers = cjson.encode({
                {
                    hostname = "hostname",
                    sslCert = {
                        pemCertKey = "pemCertKey"
                    }
                },
                {
                    hostname = "hostname2",
                    sslCert = {
                        pemCertKey = "pemCertKey2"
                    }
                }
            })
            ngx.req.get_body_data = function() return mock_servers end

            local s = spy.on(ngx, "log")
            assert.has_no.errors(configuration.handle_servers)
            assert.spy(s).was_called_with(ngx.ERR, 
                "error setting certificate for hostname: error\nerror setting certificate for hostname2: error\n")
            assert.same(ngx.status, ngx.HTTP_INTERNAL_SERVER_ERROR)
        end)

        it("logs a warning when entry is forcibly stored", function()
            local stored_entries = {}

            ngx.var.request_method = "POST"
            ngx.shared.certificate_data.set = function(self, key, value)
              stored_entries[key] = value
              return true, nil, true
            end
            local mock_servers = cjson.encode({
                {
                    hostname = "hostname",
                    sslCert = {
                        pemCertKey = "pemCertKey"
                    }
                },
                {
                    hostname = "hostname2",
                    sslCert = {
                        pemCertKey = "pemCertKey2"
                    }
                }
            })
            ngx.req.get_body_data = function() return mock_servers end

            local s1 = spy.on(ngx, "log")
            assert.has_no.errors(configuration.handle_servers)
            assert.spy(s1).was_called_with(ngx.WARN, "certificate_data dictionary is full, LRU entry has been removed to store hostname")
            assert.equal("pemCertKey", stored_entries["hostname"])
            assert.equal("pemCertKey2", stored_entries["hostname2"])
            assert.same(ngx.HTTP_CREATED, ngx.status)
        end)
    end)
end)
