
local main = require("plugins.hello_world.main")

-- The unit tests are run within a timer phase in a headless Nginx process.
-- Since `set_header` and `ngx.var.http_` API are disabled in this phase we have to stub it 
-- to avoid `API disabled in the current context` error.

describe("main", function()
  describe("rewrite", function()
    it("sets x-hello-world header to 1 when user agent is hello", function()
      ngx.var = { http_user_agent = "hello" }
      stub(ngx.req, "set_header")
      main.rewrite()
      assert.stub(ngx.req.set_header).was_called_with("x-hello-world", "1")
    end)

    it("does not set x-hello-world header to 1 when user agent is not hello", function()
      ngx.var = { http_user_agent = "not-hello" }
      stub(ngx.req, "set_header")
      main.rewrite()
      assert.stub(ngx.req.set_header).was_not_called_with("x-hello-world", "1")
    end)
  end)
end)
