local mqtt
local original_ngx = ngx

local function reset_ngx()
    _G.ngx = original_ngx
end

local function reset_mqtt()
    package.loaded["protocols.mqtt"] = nil
    mqtt = require("protocols.mqtt")
end

local function mock_ngx(mock)
    local _ngx = mock
    setmetatable(_ngx, { __index = ngx })
    _G.ngx = _ngx
end

local function fromhex(str)
    return (str:gsub('..', function(c)
        return string.char(tonumber(c, 16))
    end))
end

describe("Protocol MQTT", function()

    after_each(function()
        reset_ngx()
    end)

    describe("MQTT 3.1", function()
        it("Parse MQTT connect package, hex generate from MQTT client(MQTT.fx)", function()
            local function peek(self, size)
                return fromhex("105700064d514973647003c2003c00137061686f393338393030373832393237363030001235343631343636333133343035373637363800203633453245384344363136343346434241304134334331324339453942463342"):sub(1, size)
            end

            local function socket()
                return { peek = peek }
            end

            mock_ngx({ req = { socket = socket }, var = {} })
            reset_mqtt()
            mqtt.parse()
            assert.equal("MQIsdp", ngx.var.mqtt_protocol_name)
            assert.equal(3, ngx.var.mqtt_protocol_version)
            assert.equal("paho938900782927600", ngx.var.mqtt_client_id)
            assert.equal("546146631340576768", ngx.var.mqtt_user_name)
            assert.equal("63E2E8CD61643FCBA0A43C12C9E9BF3B", ngx.var.mqtt_password)
        end)
    end)

    describe("MQTT 3.1.1", function()
        it("Parse MQTT connect package, hex generate from MQTT client(MQTT.fx)", function()
            local function peek(self, size)
                return fromhex("105500044d51545404c2003c00137061686f393338393030373832393237363030001235343631343636333133343035373637363800203633453245384344363136343346434241304134334331324339453942463342"):sub(1, size)
            end

            local function socket()
                return { peek = peek }
            end

            mock_ngx({ req = { socket = socket }, var = {} })
            reset_mqtt()
            mqtt.parse()
            assert.equal("MQTT", ngx.var.mqtt_protocol_name)
            assert.equal(4, ngx.var.mqtt_protocol_version)
            assert.equal("paho938900782927600", ngx.var.mqtt_client_id)
            assert.equal("546146631340576768", ngx.var.mqtt_user_name)
            assert.equal("63E2E8CD61643FCBA0A43C12C9E9BF3B", ngx.var.mqtt_password)
        end)
    end)
end)
