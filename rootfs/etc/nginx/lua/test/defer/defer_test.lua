package.path = "./rootfs/etc/nginx/lua/?.lua;./rootfs/etc/nginx/lua/test/mocks/?.lua;" .. package.path
_G._TEST = true
local defer = require('defer')

local _ngx = {
    shared = {},
    log = function(...) end,
    get_phase = function() return "timer" end,
}
_G.ngx = _ngx
  
describe("Defer", function()
   describe("to_timer_phase", function()
    it("executes passed callback immediately if called on timer phase", function()
        defer.counter = 0
        defer.to_timer_phase(function() defer.counter = defer.counter + 1 end)
        assert.equal(defer.counter, 1)
    end)
   end)
end)
