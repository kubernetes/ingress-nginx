local split = require("util.split")


describe("split", function()
  it("get_last_value", function()
    for _, case in ipairs({
        {"127.0.0.1:26157 : 127.0.0.1:26158", "127.0.0.1:26158"},
        {"127.0.0.1:26157, 127.0.0.1:26158", "127.0.0.1:26158"},
        {"127.0.0.1:26158", "127.0.0.1:26158"},
    }) do
      local last = split.get_last_value(case[1])
      assert.equal(case[2], last)
    end
  end)
end)
