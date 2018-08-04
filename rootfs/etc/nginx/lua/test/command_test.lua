_G._TEST = true
local cjson = require("cjson")

describe("Command", function()
    local commands = require("commands")
    describe("basic command", function()
        it("successfuly encodes the current stats of nginx to JSON", function()
            ngx.req.set_body_data('{"command" = "uname -a"}')
            ngx.req.set_method(ngx.HTTP_POST)
            ngx.req.set_uri("/execute", true)

            local command_output = commands.call()
            local decoded_command_output = cjson.decode(command_output)

            local expected_json_stats = {
                stdout = "",
                stderr = "",
                code = 0,
            }
            assert.are.same(decoded_json_stats,decoded_command_output)
        end)
    end)
end)
