if ngx.var.request_method == "POST" then
    ngx.req.read_body()
    local backends_json = ngx.req.get_body_data()
    local shared_memory = ngx.shared.shared_memory
    shared_memory:set("CFGS", backends_json, 0)
    ngx.exit(200)
elseif method == "GET" then
    ngx.exit(405)
end