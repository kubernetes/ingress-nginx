if ngx.var.request_method == "POST" then
    ngx.req.read_body()
    local backends_json = ngx.req.get_body_data()
    local shared_memory = ngx.shared.shared_memory
    local success, err, forcible = shared_memory:set("CFGS", backends_json, 0)
    if ( success ) then
        ngx.exit(200)
    end
    ngx.log(ngx.WARN, "Error initializing CFGS: err=" .. tostring(err), ", forcible=" .. tostring(forcible))
    ngx.exit(400)
elseif method == "GET" then
    ngx.exit(405)
end