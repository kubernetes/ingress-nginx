local original_ngx = ngx
local function reset_ngx()
  _G.ngx = original_ngx
end

local function mock_ngx(mock)
  local _ngx = mock
  setmetatable(_ngx, { __index = ngx })
  _G.ngx = _ngx
end

describe("lua_ngx_var", function()
  local util = require("util")

  before_each(function()
    mock_ngx({ var = { remote_addr = "192.168.1.1", [1] = "nginx/regexp/1/group/capturing" } })
  end)

  after_each(function()
    reset_ngx()
    package.loaded["monitor"] = nil
  end)

  it("returns value of nginx var by key", function()
    assert.equal("192.168.1.1", util.lua_ngx_var("$remote_addr"))
  end)

  it("returns value of nginx var when key is number", function()
    assert.equal("nginx/regexp/1/group/capturing", util.lua_ngx_var("$1"))
  end)

  it("returns nil when variable is not defined", function()
    assert.equal(nil, util.lua_ngx_var("$foo_bar"))
  end)
end)
