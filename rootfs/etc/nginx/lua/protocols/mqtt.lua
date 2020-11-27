-- Decodes Mqtt CONNECT messages from nginx stream preread buffer
-- Refer to https://github.com/netty/netty/blob/e5951d46fc89db507ba7d2968d2ede26378f0b04/codec-mqtt/src/main/java/io/netty/handler/codec/mqtt/MqttDecoder.java
-- Support MQTT 3.1 & 3.1.1

local ngx = ngx

local CONNECT = 1
-- 4096 is preread buffer size, 2 bytes for fixed_header
local MAX_CONNECT_MESSAGE_LENGTH = 4096 - 2

local function decode_msb_lsb(message, byte_cursor)
    local msb_size = string.byte(message, byte_cursor)
    local lsb_size = string.byte(message, byte_cursor + 1)
    return bit.bor(bit.lshift(msb_size, 8), lsb_size)
end

local function decode_string(message, byte_cursor)
    local size = decode_msb_lsb(message, byte_cursor)
    local next_cursor = byte_cursor + 2 + size
    return next_cursor, message:sub(byte_cursor + 2, next_cursor - 1)
end

local function decode_fixed_header(sock)
    local raw_fixed_header = assert(sock:peek(6))
    local byte = string.byte(raw_fixed_header)
    local message_type = bit.rshift(byte, 4)
    local dup_flag = bit.band(byte, 0x08) == 0x08
    local qos_level = bit.rshift(bit.band(byte, 0x06), 1)
    local retain = bit.band(byte, 0x01) ~= 0

    local remaining_length = 0
    local remaining_length_bytes = 0
    local multiplier = 1
    local digit
    for i = 2, 6 do
        remaining_length_bytes = remaining_length_bytes + 1
        digit = string.byte(raw_fixed_header, i)
        remaining_length = remaining_length + bit.band(digit, 127) * multiplier
        multiplier = multiplier * 128
        if (bit.band(digit, 128) == 0) then
            break
        end
    end
    if remaining_length_bytes > 4 then
        ngx.log(ngx.ERR, "MQTT protocol Remaining Length exceed 4 bytes")
        return {}
    end
    return { message_type = message_type, dup_flag = dup_flag, qos_level = qos_level, retain = retain, remaining_length = remaining_length }
end

local function decode_connect_variable_header(message)
    local byte_cursor, proto_name = decode_string(message, 1)

    local proto_version = string.byte(message, byte_cursor)
    byte_cursor = byte_cursor + 1

    local flag_byte = string.byte(message, byte_cursor)
    byte_cursor = byte_cursor + 1

    local keep_alive = decode_msb_lsb(message, byte_cursor)
    byte_cursor = byte_cursor + 2

    local has_user_name = bit.band(flag_byte, 0x80) == 0x80
    local has_password = bit.band(flag_byte, 0x40) == 0x40
    local will_retain = bit.band(flag_byte, 0x20) == 0x20
    local will_qos = bit.rshift(bit.band(flag_byte, 0x18), 3)
    local has_will_flag = bit.band(flag_byte, 0x04) == 0x04
    local clean_session = bit.band(flag_byte, 0x02) == 0x02

    return byte_cursor, { proto_name = proto_name, proto_version = proto_version, keep_alive = keep_alive, has_user_name = has_user_name, has_password = has_password, will_retain = will_retain, will_qos = will_qos, has_will_flag = has_will_flag, clean_session = clean_session }
end

local function decode_connect_payload(message, variable_header, start)
    local byte_cursor, client_id = decode_string(message, start)

    local will_topic, will_message
    if variable_header.has_will_flag then
        byte_cursor, will_topic = decode_string(message, byte_cursor)
        byte_cursor, will_message = decode_string(message, byte_cursor)
    end

    local user_name
    if variable_header.has_user_name then
        byte_cursor, user_name = decode_string(message, byte_cursor)
    end

    local password
    if variable_header.has_password then
        byte_cursor, password = decode_string(message, byte_cursor)
    end

    return { client_id = client_id, will_topic = will_topic, will_message = will_message, user_name = user_name, password = password }
end

local _M = {}

function _M.parse()
    local sock = assert(ngx.req.socket())
    local fixed_header = decode_fixed_header(sock)
    if fixed_header.message_type ~= CONNECT then
        ngx.log(ngx.WARN, "Expect MQTT CONNECT Message, Get " .. fixed_header.message_type)
        return
    end
    if fixed_header.remaining_length > MAX_CONNECT_MESSAGE_LENGTH then
        ngx.log(ngx.ERR, "Connect Message Too Large, Expect Below " .. MAX_CONNECT_MESSAGE_LENGTH .. ", Get" .. fixed_header.remaining_length)
        return
    end

    local message = assert(sock:peek(fixed_header.remaining_length + 2)):sub(3)

    local byte_cursor, variable_header = decode_connect_variable_header(message)
    ngx.var.mqtt_protocol_name = variable_header.proto_name
    ngx.var.mqtt_protocol_version = variable_header.proto_version

    local payload = decode_connect_payload(message, variable_header, byte_cursor)
    ngx.var.mqtt_client_id = payload.client_id
    if variable_header.has_user_name then
        ngx.var.mqtt_user_name = payload.user_name
    end
    if variable_header.has_password then
        ngx.var.mqtt_password = payload.password
    end
end

return _M
