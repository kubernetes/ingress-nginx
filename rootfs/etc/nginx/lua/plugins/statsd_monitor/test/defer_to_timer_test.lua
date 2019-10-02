_G._TEST = true

local defer_to_timer = require("plugins.statsd_monitor.defer_to_timer")

describe("defer_to_timer", function()
  local f = function() end

  describe("flush_queue", function()
    it("resets the queue", function()
      assert.has_no.errors(function() defer_to_timer.enqueue(f) end)
      assert.equal(1, #defer_to_timer.get_queue())

      assert.has_no.errors(function() defer_to_timer.flush_queue() end)
      assert.equal(0, #defer_to_timer.get_queue())
    end)

    it("runs the functions in queue", function()
      local calls = {}
      local f1 = function() table.insert(calls, "f1") end
      local f2 = function() table.insert(calls, "f2") end
      assert.has_no.errors(function() defer_to_timer.enqueue(f1) end)
      assert.has_no.errors(function() defer_to_timer.enqueue(f2) end)

      assert.has_no.errors(function() defer_to_timer.flush_queue() end)
      assert.are.same({ "f1", "f2" }, calls)
    end)
  end)

  describe("enqueue", function()
    before_each(function()
      assert.has_no.errors(function() defer_to_timer.flush_queue() end)
      assert.equal(0, #defer_to_timer.get_queue())
    end)

    it("appends the function to queue", function()
      assert.has_no.errors(function() defer_to_timer.enqueue(f) end)
      assert.has_no.errors(function() defer_to_timer.enqueue(f) end)
      assert.equal(2, #defer_to_timer.get_queue())
    end)

    it("returns an error when queue is full", function()
      for _ = 1,defer_to_timer.MAX_QUEUE_SIZE do
        assert.has_no.errors(function() defer_to_timer.enqueue(f) end)
      end
      assert.equal(defer_to_timer.MAX_QUEUE_SIZE, #defer_to_timer.get_queue())

      local err = defer_to_timer.enqueue(f)
      assert.equal("deferred timer queue full", err)
    end)
  end)
end)
