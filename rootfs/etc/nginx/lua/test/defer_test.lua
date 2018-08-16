_G._TEST = true
local defer = require("util.defer")

describe("Defer", function()
   describe("to_timer_phase", function()
    it("executes passed callback immediately if called on timer phase", function()
        defer.counter = 0
        defer.to_timer_phase(function() defer.counter = defer.counter + 1 end)
        assert.equal(defer.counter, 1)
    end)
   end)
end)
