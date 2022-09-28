local original_io_open = io.open

describe("resolv_conf", function()
  before_each(function()
    package.loaded["util.resolv_conf"] = nil
    io.open = original_io_open
  end)

  it("errors when file can not be opened", function()
    io.open = function(...)
      return nil, "file does not exist"
    end

    assert.has_error(function() require("util.resolv_conf") end, "could not open /etc/resolv.conf: file does not exist")
  end)

  it("opens '/etc/resolv.conf' with mode 'r'", function()
    io.open = function(path, mode)
      assert.are.same("/etc/resolv.conf", path)
      assert.are.same("r", mode)

      return original_io_open(path, mode)
    end

    assert.has_no.errors(function() require("util.resolv_conf") end)
  end)

  it("correctly parses resolv.conf", function()
    local conf = [===[
# This is a comment
nameserver 10.96.0.10
nameserver 10.96.0.99
nameserver 2001:4860:4860::8888
search ingress-nginx.svc.cluster.local svc.cluster.local cluster.local
options ndots:5
    ]===]

    helpers.with_resolv_conf(conf, function()
      local resolv_conf = require("util.resolv_conf")
      assert.are.same({
        nameservers = { "10.96.0.10", "10.96.0.99", "[2001:4860:4860::8888]" },
        search = { "ingress-nginx.svc.cluster.local", "svc.cluster.local", "cluster.local" },
        ndots = 5,
      }, resolv_conf)
    end)
  end)

  it("ignores options that it does not understand", function()
    local conf = [===[
nameserver 10.96.0.10
search example.com
options debug
options ndots:3
    ]===]

    helpers.with_resolv_conf(conf, function()
      local resolv_conf = require("util.resolv_conf")
      assert.are.same({
        nameservers = { "10.96.0.10" },
        search = { "example.com" },
        ndots = 3,
      }, resolv_conf)
    end)
  end)
end)
