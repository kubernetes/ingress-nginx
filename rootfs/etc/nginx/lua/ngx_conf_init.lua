local function initialize_ingress(statusport, enablemetrics, ocsp, ingress)
    collectgarbage("collect")
    -- init modules
    local ok, res
    ok, res = pcall(require, "lua_ingress")
    if not ok then
      error("require failed: " .. tostring(res))
    else
      lua_ingress = res
      lua_ingress.set_config(ingress)
    end

    ok, res = pcall(require, "configuration")
    if not ok then
      error("require failed: " .. tostring(res))
    else
      configuration = res
      configuration.prohibited_localhost_port = statusport
    end

    ok, res = pcall(require, "balancer")
    if not ok then
      error("require failed: " .. tostring(res))
    else
      balancer = res
    end

    if enablemetrics then
        ok, res = pcall(require, "monitor")
        if not ok then
            error("require failed: " .. tostring(res))
        else
            monitor = res
        end
    end

    ok, res = pcall(require, "certificate")
    if not ok then
      error("require failed: " .. tostring(res))
    else
      certificate = res
      certificate.is_ocsp_stapling_enabled = ocsp
    end
end

return { initialize_ingress = initialize_ingress }