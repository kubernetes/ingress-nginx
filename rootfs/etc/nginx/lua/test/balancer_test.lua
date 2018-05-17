local cwd = io.popen("pwd"):read('*l')
package.path = cwd .. "/rootfs/etc/nginx/lua/?.lua;" .. package.path

local balancer, mock_cjson, mock_config, mock_backends, mock_lrucache, lock, mock_lock,
  mock_ngx_balancer, mock_ewma

local function dict_generator(vals)
  local _dict = { __index = {
      get = function(self, key)
        return self._vals[key]
      end,
      set = function(self, key, val)
          self._vals[key] = val
          return true, nil, false
      end,
      delete = function(self, key)
        return self:set(key, nil)
      end,
      flush_all = function(self)
        return
      end,
      _vals = vals
    }
  }
  return setmetatable({_vals = vals}, _dict)
end

local function backend_generator(name, endpoints, lb_alg)
  return {
    name = name,
    endpoints = endpoints,
    ["load-balance"] = lb_alg,
  }
end

local default_endpoints = {
  {address = "000.000.000", port = "8080"},
  {address = "000.000.001", port = "8081"},
}

local default_backends = {
  mock_rr_backend = backend_generator("mock_rr_backend", default_endpoints, "round_robin"),
  mock_ewma_backend = backend_generator("mock_ewma_backend", default_endpoints, "ewma"),
}

local function init()
  mock_cjson = {}
  mock_config = {}
  mock_ngx_balancer = {}
  mock_ewma = {
    sync = function(b) return end
  }
  mock_resty_balancer = {
    sync = function(b) return end,
    after_balance = function () return end
  }
  mock_backends = dict_generator(default_backends)
  mock_lrucache = {
    new = function () return mock_backends end
  }
  lock = {
    lock = function() return end,
    unlock = function() return end
  }
  mock_lock = {
    new = function () return lock end
  }
  _G.ngx = {
    shared = {
      balancer_ewma = dict_generator({}),
      balancer_ewma_last_touched_at = dict_generator({}),
    },
    var = {},
    log = function() return end,
    WARN = "warn",
    INFO = "info",
    ERR = "err",
    HTTP_SERVICE_UNAVAILABLE = 503,
    exit = function(status) return end,
  }
  package.loaded["ngx.balancer"] = mock_ngx_balancer
  package.loaded["resty.lrucache"] = mock_lrucache
  package.loaded["resty.string"] = {}
  package.loaded["resty.sha1"] = {}
  package.loaded["resty.md5"] = {}
  package.loaded["resty.cookie"] = {}
  package.loaded["cjson"] = mock_cjson
  package.loaded["resty.lock"] = mock_lock
  package.loaded["balancer.ewma"] = mock_ewma
  package.loaded["balancer.resty"] = mock_resty_balancer
  package.loaded["configuration"] = mock_config
  balancer = require("balancer")
end

describe("[balancer_test]", function()
  setup(function()
    init()
  end)

  teardown(function()
    local packages = {"ngx.balancer", "resty.lrucache","cjson", "resty.lock", "balancer.ewma","configuration"}
    for i, package_name in ipairs(packages) do
      package.loaded[package_name] = nil
    end
  end)

  describe("balancer.call():", function()
    setup(function()
      mock_ngx_balancer.set_more_tries = function () return end
      mock_ngx_balancer.set_current_peer = function () return end
      mock_ewma.after_balance = function () return end
    end)

    before_each(function()
      _G.ngx.get_phase = nil
      _G.ngx.var = {}
      mock_backends._vals = default_backends
    end)

    describe("phase=log", function()
      before_each(function()
        _G.ngx.get_phase = function() return "log" end
      end)

      it("lb_alg=ewma, ewma_after_balance was called", function()
        _G.ngx.var.proxy_upstream_name = "mock_ewma_backend"

        local ewma_after_balance_spy = spy.on(mock_ewma, "after_balance")

        mock_resty_balancer.is_applicable = function(b) return false end
        assert.has_no_errors(balancer.call)
        assert.spy(ewma_after_balance_spy).was_called()
      end)

      it("lb_alg=round_robin, ewma_after_balance was not called", function()
        _G.ngx.var.proxy_upstream_name = "mock_rr_backend"

        local ewma_after_balance_spy = spy.on(mock_ewma, "after_balance")

        mock_resty_balancer.is_applicable = function(b) return true end
        assert.has_no_errors(balancer.call)
        assert.spy(ewma_after_balance_spy).was_not_called()
      end)
    end)

    describe("phase=balancer", function()
      before_each(function()
        _G.ngx.get_phase = function() return "balancer" end
      end)

      it("lb_alg=round_robin, peer was successfully set", function()
        _G.ngx.var.proxy_upstream_name = "mock_rr_backend"

        local backend_get_spy = spy.on(mock_backends, "get")
        local set_more_tries_spy = spy.on(mock_ngx_balancer, "set_more_tries")
        local set_current_peer_spy = spy.on(mock_ngx_balancer, "set_current_peer")

        mock_resty_balancer.balance = function(b) return {address = "000.000.000", port = "8080"} end
        assert.has_no_errors(balancer.call)
        assert.spy(backend_get_spy).was_called_with(match.is_table(), "mock_rr_backend")
        assert.spy(set_more_tries_spy).was_called_with(1)
        assert.spy(set_current_peer_spy).was_called_with("000.000.000", "8080")

        mock_backends.get:clear()
        mock_ngx_balancer.set_more_tries:clear()
        mock_ngx_balancer.set_current_peer:clear()

        mock_resty_balancer.balance = function(b) return {address = "000.000.001", port = "8081"} end
        assert.has_no_errors(balancer.call)
        assert.spy(backend_get_spy).was_called_with(match.is_table(), "mock_rr_backend")
        assert.spy(set_more_tries_spy).was_called_with(1)
        assert.spy(set_current_peer_spy).was_called_with("000.000.001", "8081")
      end)

      it("lb_alg=ewma, peer was successfully set", function()
        _G.ngx.var.proxy_upstream_name = "mock_ewma_backend"

        mock_ewma.balance = function(b) return {address = "000.000.111", port = "8083"} end

        local backend_get_spy = spy.on(mock_backends, "get")
        local set_more_tries_spy = spy.on(mock_ngx_balancer, "set_more_tries")
        local set_current_peer_spy = spy.on(mock_ngx_balancer, "set_current_peer")

        mock_resty_balancer.is_applicable = function(b) return false end
        assert.has_no_errors(balancer.call)
        assert.spy(backend_get_spy).was_called_with(match.is_table(), "mock_ewma_backend")
        assert.spy(set_more_tries_spy).was_called_with(1)
        assert.spy(set_current_peer_spy).was_called_with("000.000.111", "8083")
      end)

      it("fails when no backend exists", function()
        _G.ngx.var.proxy_upstream_name = "mock_rr_backend"

        mock_backends._vals = {}

        local backend_get_spy = spy.on(mock_backends, "get")
        local set_current_peer_spy = spy.on(mock_ngx_balancer, "set_current_peer")

        assert.has_no_errors(balancer.call)
        assert.are_equal(ngx.status, 503)
        assert.spy(backend_get_spy).was_called_with(match.is_table(), "mock_rr_backend")
        assert.spy(set_current_peer_spy).was_not_called()
      end)
    end)

    describe("not in phase log or balancer", function()
      it("returns errors", function()
        _G.ngx.get_phase = function() return "nope" end
        assert.has_error(balancer.call, "must be called in balancer or log, but was called in: nope")
      end)
    end)
  end)

  describe("balancer.init_worker():", function()
    setup(function()
      _G.ngx.timer = {
        every = function(interval, func) return func() end
      }
      mock_cjson.decode = function(x) return x end
    end)

    before_each(function()
      mock_backends._vals = default_backends
    end)

    describe("sync_backends():", function()
      it("succeeds when no sync is required", function()
        mock_config.get_backends_data = function() return default_backends end

        local backend_set_spy = spy.on(mock_backends, "set")

        assert.has_no_errors(balancer.init_worker)
        assert.spy(backend_set_spy).was_not_called()
      end)

      it("lb_alg=round_robin, updates backend when sync is required", function()
        mock_config.get_backends_data = function() return { default_backends.mock_rr_backend } end
        mock_backends._vals = {}

        local backend_set_spy = spy.on(mock_backends, "set")
        local ewma_flush_spy = spy.on(_G.ngx.shared.balancer_ewma, "flush_all")
        local ewma_lta_flush_spy = spy.on(_G.ngx.shared.balancer_ewma_last_touched_at, "flush_all")

        mock_resty_balancer.balance = function(b) return {address = "000.000.000", port = "8080"} end
        mock_resty_balancer.reinit = function(b) return end
        assert.has_no_errors(balancer.init_worker)
        assert.spy(backend_set_spy)
          .was_called_with(match.is_table(), default_backends.mock_rr_backend.name, match.is_table())
        assert.spy(ewma_flush_spy).was_not_called()
        assert.spy(ewma_lta_flush_spy).was_not_called()
      end)

      it("lb_alg=ewma, updates backend when sync is required", function()
        _G.ngx.var.proxy_upstream_name = "mock_ewma_backend"
        mock_config.get_backends_data = function() return { default_backends.mock_ewma_backend } end
        mock_backends._vals = {}

        local backend_set_spy = spy.on(mock_backends, "set")
        local ewma_sync_spy = spy.on(mock_ewma, "sync")

        mock_resty_balancer.is_applicable = function(b) return false end
        assert.has_no_errors(balancer.init_worker)
        assert.spy(backend_set_spy)
          .was_called_with(match.is_table(), default_backends.mock_ewma_backend.name, match.is_table())
        assert.spy(ewma_sync_spy).was_called()
      end)
    end)
  end)
end)
