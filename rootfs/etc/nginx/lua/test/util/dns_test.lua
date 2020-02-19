local conf = [===[
nameserver 1.2.3.4
nameserver 4.5.6.7
search ingress-nginx.svc.cluster.local svc.cluster.local cluster.local
options ndots:5
]===]

package.loaded["util.resolv_conf"] = nil

helpers.with_resolv_conf(conf, function()
  require("util.resolv_conf")
end)

describe("dns.lookup", function()
  local dns, dns_lookup, spy_ngx_log

  before_each(function()
    spy_ngx_log = spy.on(ngx, "log")
    dns = require("util.dns")
    dns_lookup = dns.lookup
  end)

  after_each(function()
    package.loaded["util.dns"] = nil
  end)

  it("sets correct nameservers", function()
    helpers.mock_resty_dns_new(function(self, options)
      assert.are.same({ nameservers = { "1.2.3.4", "4.5.6.7" }, retrans = 5, timeout = 2000 }, options)
      return nil, ""
    end)
    dns_lookup("example.com")
  end)

  describe("when there's an error", function()
    it("returns host when resolver can not be instantiated", function()
      helpers.mock_resty_dns_new(function(...) return nil, "an error" end)
      assert.are.same({ "example.com" }, dns_lookup("example.com"))
      assert.spy(spy_ngx_log).was_called_with(ngx.ERR, "failed to instantiate the resolver: an error")
    end)

    it("returns host when the query returns nil", function()
      helpers.mock_resty_dns_query(nil, nil, "oops!")
      assert.are.same({ "example.com" }, dns_lookup("example.com"))    
      assert.spy(spy_ngx_log).was_called_with(ngx.ERR, "failed to query the DNS server for ", "example.com", ":\n", "oops!\noops!")
    end)

    it("returns host when the query returns empty answer", function()
      helpers.mock_resty_dns_query(nil, {})
      assert.are.same({ "example.com" }, dns_lookup("example.com"))
      assert.spy(spy_ngx_log).was_called_with(ngx.ERR, "failed to query the DNS server for ", "example.com", ":\n", "no A record resolved\nno AAAA record resolved")
    end)

    it("returns host when there's answer but with error", function()
      helpers.mock_resty_dns_query(nil, { errcode = 1, errstr = "format error" })
      assert.are.same({ "example.com" }, dns_lookup("example.com"))
      assert.spy(spy_ngx_log).was_called_with(ngx.ERR, "failed to query the DNS server for ", "example.com", ":\n", "server returned error code: 1: format error\nserver returned error code: 1: format error")
    end)

    it("retuns host when there's answer but no A/AAAA record in it", function()
      helpers.mock_resty_dns_query(nil, { { name = "example.com", cname = "sub.example.com", ttl = 60 } })
      assert.are.same({ "example.com" }, dns_lookup("example.com"))
      assert.spy(spy_ngx_log).was_called_with(ngx.ERR, "failed to query the DNS server for ", "example.com", ":\n", "no A record resolved\nno AAAA record resolved")
    end)

    it("returns host when the query returns nil and number of dots is not less than configured ndots", function()
      helpers.mock_resty_dns_query(nil, nil, "oops!")
      assert.are.same({ "a.b.c.d.example.com" }, dns_lookup("a.b.c.d.example.com"))    
      assert.spy(spy_ngx_log).was_called_with(ngx.ERR, "failed to query the DNS server for ", "a.b.c.d.example.com", ":\n", "oops!\noops!")
    end)

    it("returns host when the query returns nil for a fully qualified domain", function()
      helpers.mock_resty_dns_query("example.com.", nil, "oops!")
      assert.are.same({ "example.com." }, dns_lookup("example.com."))
      assert.spy(spy_ngx_log).was_called_with(ngx.ERR, "failed to query the DNS server for ", "example.com.", ":\n", "oops!\noops!")
    end)
  end)

  it("returns answer from cache if it exists without doing actual DNS query", function()
    dns._cache:set("example.com", { "192.168.1.1" })
    assert.are.same({ "192.168.1.1" }, dns_lookup("example.com"))
  end)

  it("resolves a fully qualified domain without looking at resolv.conf search and caches result", function()
    helpers.mock_resty_dns_query("example.com.", {
      {
        name = "example.com.",
        address = "192.168.1.1",
        ttl = 3600,
      },
      {
        name = "example.com",
        address = "1.2.3.4",
        ttl = 60,
      }
    })
    assert.are.same({ "192.168.1.1", "1.2.3.4" }, dns_lookup("example.com."))
    assert.are.same({ "192.168.1.1", "1.2.3.4" }, dns._cache:get("example.com."))
  end)

  it("starts with host itself when number of dots is not less than configured ndots", function()
    local host = "a.b.c.d.example.com"
    helpers.mock_resty_dns_query(host, { { name = host, address = "192.168.1.1", ttl = 3600, } } )

    assert.are.same({ "192.168.1.1" }, dns_lookup(host))
    assert.are.same({ "192.168.1.1" }, dns._cache:get(host))
  end)

  it("starts with first search entry when number of dots is less than configured ndots", function()
    local host = "example.com.ingress-nginx.svc.cluster.local"
    helpers.mock_resty_dns_query(host, { { name = host, address = "192.168.1.1", ttl = 3600, } } )

    assert.are.same({ "192.168.1.1" }, dns_lookup(host))
    assert.are.same({ "192.168.1.1" }, dns._cache:get(host))
  end)

  it("it caches with minimal ttl", function()
    helpers.mock_resty_dns_query("example.com.", {
      {
        name = "example.com.",
        address = "192.168.1.1",
        ttl = 3600,
      },
      {
        name = "example.com.",
        address = "1.2.3.4",
        ttl = 60,
      }
    })

    local spy_cache_set = spy.on(dns._cache, "set")

    assert.are.same({ "192.168.1.1", "1.2.3.4" }, dns_lookup("example.com."))
    assert.spy(spy_cache_set).was_called_with(match.is_table(), "example.com.", { "192.168.1.1", "1.2.3.4" }, 60)
  end)
end)
