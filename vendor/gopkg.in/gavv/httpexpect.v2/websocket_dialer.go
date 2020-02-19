package httpexpect

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/valyala/fasthttp"
)

// NewWebsocketDialer produces new websocket.Dialer which dials to bound
// http.Handler without creating a real net.Conn.
func NewWebsocketDialer(handler http.Handler) *websocket.Dialer {
	return &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			hc := newHandlerConn()
			hc.runHandler(handler)
			return hc, nil
		},
	}
}

// NewFastWebsocketDialer produces new websocket.Dialer which dials to bound
// fasthttp.RequestHandler without creating a real net.Conn.
func NewFastWebsocketDialer(handler fasthttp.RequestHandler) *websocket.Dialer {
	return &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			hc := newHandlerConn()
			hc.runFastHandler(handler)
			return hc, nil
		},
	}
}

type handlerConn struct {
	net.Conn          // returned from dialer
	backConn net.Conn // passed to the background goroutine

	wg sync.WaitGroup
}

func newHandlerConn() *handlerConn {
	dialConn, backConn := net.Pipe()

	return &handlerConn{
		Conn:     dialConn,
		backConn: backConn,
	}
}

func (hc *handlerConn) Close() error {
	err := hc.Conn.Close()
	hc.wg.Wait() // wait the background goroutine
	return err
}

func (hc *handlerConn) runHandler(handler http.Handler) {
	hc.wg.Add(1)

	go func() {
		defer hc.wg.Done()

		recorder := &hijackRecorder{conn: hc.backConn}

		for {
			req, err := http.ReadRequest(bufio.NewReader(hc.backConn))
			if err != nil {
				return
			}
			handler.ServeHTTP(recorder, req)
		}
	}()
}

func (hc *handlerConn) runFastHandler(handler fasthttp.RequestHandler) {
	hc.wg.Add(1)

	go func() {
		defer hc.wg.Done()

		_ = fasthttp.ServeConn(hc.backConn, handler)
	}()
}

// hijackRecorder it similar to httptest.ResponseRecorder,
// but with Hijack capabilities.
//
// Original idea is stolen from https://github.com/posener/wstest
type hijackRecorder struct {
	httptest.ResponseRecorder
	conn net.Conn
}

// Hijack the connection for caller.
//
// Implements http.Hijacker interface.
func (r *hijackRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	rw := bufio.NewReadWriter(bufio.NewReader(r.conn), bufio.NewWriter(r.conn))
	return r.conn, rw, nil
}

// WriteHeader write HTTP header to the client and closes the connection
//
// Implements http.ResponseWriter interface.
func (r *hijackRecorder) WriteHeader(code int) {
	resp := http.Response{StatusCode: code, Header: r.Header()}
	_ = resp.Write(r.conn)
}
