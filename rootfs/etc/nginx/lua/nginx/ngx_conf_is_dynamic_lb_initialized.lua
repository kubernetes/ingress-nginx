local configuration = require("configuration")
local backend_data = configuration.get_backends_data()
if not backend_data then
    ngx.exit(ngx.HTTP_INTERNAL_SERVER_ERROR)
    return
end

ngx.say("OK")
ngx.exit(ngx.HTTP_OK)