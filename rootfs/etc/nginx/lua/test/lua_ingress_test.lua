describe("lua_ingress", function()
  it("patches math.randomseed to not be called more than once per worker", function()
    local s = spy.on(ngx, "log")

    math.randomseed()
    assert.spy(s).was_called_with(ngx.WARN,
      string.format("ignoring math.randomseed() since PRNG is already seeded for worker %d", ngx.worker.pid()))
  end)
end)
