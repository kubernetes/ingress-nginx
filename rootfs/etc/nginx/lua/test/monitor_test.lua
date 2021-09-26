local cjson = require("cjson.safe")

local original_ngx = ngx
local function reset_ngx()
  _G.ngx = original_ngx
end

local function mock_ngx(mock)
  local _ngx = mock
  setmetatable(_ngx, { __index = ngx })
  _G.ngx = _ngx
end

local function mock_ngx_socket_tcp()
  local tcp_mock = {}
  stub(tcp_mock, "connect", true)
  stub(tcp_mock, "send", true)
  stub(tcp_mock, "close", true)

  local socket_mock = {}
  stub(socket_mock, "tcp", tcp_mock)
  mock_ngx({ socket = socket_mock })

  return tcp_mock
end

describe("Monitor", function()
  after_each(function()
    reset_ngx()
    package.loaded["monitor"] = nil
  end)

  it("extended batch size", function()
    mock_ngx({ var = {} })
    local monitor = require("monitor")
    monitor.set_metrics_max_batch_size(20000)

    for i = 1,20000,1 do
      monitor.call()
    end

    assert.equal(20000, #monitor.get_metrics_batch())
  end)

  it("batches metrics", function()
    mock_ngx({ var = {} })
    local monitor = require("monitor")

    for i = 1,10,1 do
      monitor.call()
    end

    assert.equal(10, #monitor.get_metrics_batch())
  end)

  describe("flush", function()
    it("short circuits when premature is true (when worker is shutting down)", function()
      local tcp_mock = mock_ngx_socket_tcp()
      mock_ngx({ var = {} })
      local monitor = require("monitor")

      for i = 1,10,1 do
        monitor.call()
      end
      monitor.flush(true)
      assert.stub(tcp_mock.connect).was_not_called()
    end)

    it("short circuits when there's no metrics batched", function()
      local tcp_mock = mock_ngx_socket_tcp()
      local monitor = require("monitor")

      monitor.flush()
      assert.stub(tcp_mock.connect).was_not_called()
    end)

    it("JSON encodes and sends the batched metrics", function()
      local tcp_mock = mock_ngx_socket_tcp()

      local ngx_var_mock = {
        host = "example.com",
        namespace = "default",
        ingress_name = "example",
        service_name = "http-svc",
        proxy_alternative_upstream_name = "default-http-svc-canary-80",
        location_path = "/",

        request_method = "GET",
        status = "200",
        request_length = "256",
        request_time = "0.04",
        bytes_sent = "512",

        upstream_addr = "10.10.0.1",
        upstream_connect_time = "0.01",
        upstream_response_time = "0.02",
        upstream_response_length = "456",
        upstream_status = "200",
      }
      mock_ngx({ var = ngx_var_mock })
      local monitor = require("monitor")
      monitor.call()

      local ngx_var_mock1 = ngx_var_mock
      ngx_var_mock1.status = "201"
      ngx_var_mock1.request_method = "POST"
      mock_ngx({ var = ngx_var_mock })
      monitor.call()

      monitor.flush()

      local expected_payload = cjson.encode({
        {
          host = "example.com",
          namespace = "default",
          ingress = "example",
          service = "http-svc",
          canary = "default-http-svc-canary-80",
          path = "/",

          method = "GET",
          status = "200",
          requestLength = 256,
          requestTime = 0.04,
          responseLength = 512,

          upstreamLatency = 0.01,
          upstreamResponseTime = 0.02,
          upstreamResponseLength = 456,
        },
        {
          host = "example.com",
          namespace = "default",
          ingress = "example",
          service = "http-svc",
          canary = "default-http-svc-canary-80",
          path = "/",

          method = "POST",
          status = "201",
          requestLength = 256,
          requestTime = 0.04,
          responseLength = 512,

          upstreamLatency = 0.01,
          upstreamResponseTime = 0.02,
          upstreamResponseLength = 456,
        },
      })

      assert.stub(tcp_mock.connect).was_called_with(tcp_mock, "unix:/tmp/prometheus-nginx.socket")
      assert.stub(tcp_mock.send).was_called_with(tcp_mock, expected_payload)
      assert.stub(tcp_mock.close).was_called_with(tcp_mock)
    end)
  end)
end)
