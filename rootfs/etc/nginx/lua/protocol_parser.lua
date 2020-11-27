local mqtt = require("protocols.mqtt")

local IMPLEMENTATIONS = {
    MQTT = mqtt,
}

local _M = {}

function _M.parse(protocol_name)
    local implementation = IMPLEMENTATIONS[protocol_name]
    if not implementation then
        ngx.log(ngx.ERR, string.format("%s protocol parser is not supported", protocol_name))
        return
    end
    implementation.parse()
end

return _M