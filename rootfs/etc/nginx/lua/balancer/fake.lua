local _M = {}

function _M.new(self, backend)
    local o = {
        fake = true,
        traffic_shaping_policy = backend.trafficShapingPolicy,
        alternative_backends = backend.alternativeBackends,
    }
    setmetatable(o, self)
    self.__index = self
    return o
end

return _M
