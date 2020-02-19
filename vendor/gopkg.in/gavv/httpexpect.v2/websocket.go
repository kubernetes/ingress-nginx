package httpexpect

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

const noDuration = time.Duration(0)

var infiniteTime = time.Time{}

// Websocket provides methods to read from, write into and close WebSocket
// connection.
type Websocket struct {
	config       Config
	chain        chain
	conn         *websocket.Conn
	readTimeout  time.Duration
	writeTimeout time.Duration
	isClosed     bool
}

// NewWebsocket returns a new Websocket given a Config with Reporter and
// Printers, and websocket.Conn to be inspected and handled.
func NewWebsocket(config Config, conn *websocket.Conn) *Websocket {
	return makeWebsocket(config, makeChain(config.Reporter), conn)
}

func makeWebsocket(config Config, chain chain, conn *websocket.Conn) *Websocket {
	return &Websocket{
		config: config,
		chain:  chain,
		conn:   conn,
	}
}

// Raw returns underlying websocket.Conn object.
// This is the value originally passed to NewConnection.
func (c *Websocket) Raw() *websocket.Conn {
	return c.conn
}

// WithReadTimeout sets timeout duration for WebSocket connection reads.
//
// By default no timeout is used.
func (c *Websocket) WithReadTimeout(timeout time.Duration) *Websocket {
	c.readTimeout = timeout
	return c
}

// WithoutReadTimeout removes timeout for WebSocket connection reads.
func (c *Websocket) WithoutReadTimeout() *Websocket {
	c.readTimeout = noDuration
	return c
}

// WithWriteTimeout sets timeout duration for WebSocket connection writes.
//
// By default no timeout is used.
func (c *Websocket) WithWriteTimeout(timeout time.Duration) *Websocket {
	c.writeTimeout = timeout
	return c
}

// WithoutWriteTimeout removes timeout for WebSocket connection writes.
//
// If not used then DefaultWebsocketTimeout will be used.
func (c *Websocket) WithoutWriteTimeout() *Websocket {
	c.writeTimeout = noDuration
	return c
}

// Subprotocol returns a new String object that may be used to inspect
// negotiated protocol for the connection.
func (c *Websocket) Subprotocol() *String {
	s := &String{chain: c.chain}
	if c.conn != nil {
		s.value = c.conn.Subprotocol()
	}
	return s
}

// Expect reads next message from WebSocket connection and
// returns a new WebsocketMessage object to inspect received message.
//
// Example:
//  msg := conn.Expect()
//  msg.JSON().Object().ValueEqual("message", "hi")
func (c *Websocket) Expect() *WebsocketMessage {
	switch {
	case c.chain.failed():
		return makeWebsocketMessage(c.chain)
	case c.conn == nil:
		c.chain.fail("\nunexpected read from failed WebSocket connection")
		return makeWebsocketMessage(c.chain)
	case c.isClosed:
		c.chain.fail("\nunexpected read from closed WebSocket connection")
		return makeWebsocketMessage(c.chain)
	case !c.setReadDeadline():
		return makeWebsocketMessage(c.chain)
	}
	var err error
	m := makeWebsocketMessage(c.chain)
	m.typ, m.content, err = c.conn.ReadMessage()
	if err != nil {
		if cls, ok := err.(*websocket.CloseError); ok {
			m.typ = websocket.CloseMessage
			m.closeCode = cls.Code
			m.content = []byte(cls.Text)
			c.printRead(m.typ, m.content, m.closeCode)
		} else {
			c.chain.fail(
				"\nexpected read WebSocket connection, "+
					"but got failure: %s", err.Error())
			return makeWebsocketMessage(c.chain)
		}
	} else {
		c.printRead(m.typ, m.content, m.closeCode)
	}
	return m
}

func (c *Websocket) setReadDeadline() bool {
	deadline := infiniteTime
	if c.readTimeout != noDuration {
		deadline = time.Now().Add(c.readTimeout)
	}
	if err := c.conn.SetReadDeadline(deadline); err != nil {
		c.chain.fail(
			"\nunexpected failure when setting "+
				"read WebSocket connection deadline: %s", err.Error())
		return false
	}
	return true
}

func (c *Websocket) printRead(typ int, content []byte, closeCode int) {
	for _, printer := range c.config.Printers {
		if p, ok := printer.(WebsocketPrinter); ok {
			p.WebsocketRead(typ, content, closeCode)
		}
	}
}

// Disconnect closes the underlying WebSocket connection without sending or
// waiting for a close message.
//
// It's okay to call this function multiple times.
//
// It's recommended to always call this function after connection usage is over
// to ensure that no resource leaks will happen.
//
// Example:
//  conn := resp.Connection()
//  defer conn.Disconnect()
func (c *Websocket) Disconnect() *Websocket {
	if c.conn == nil || c.isClosed {
		return c
	}
	c.isClosed = true
	if err := c.conn.Close(); err != nil {
		c.chain.fail("close error when disconnecting webcoket: " + err.Error())
	}
	return c
}

// Close cleanly closes the underlying WebSocket connection
// by sending an empty close message and then waiting (with timeout)
// for the server to close the connection.
//
// WebSocket close code may be optionally specified.
// If not, then "1000 - Normal Closure" will be used.
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// It's okay to call this function multiple times.
//
// Example:
//  conn := resp.Connection()
//  conn.Close(websocket.CloseUnsupportedData)
func (c *Websocket) Close(code ...int) *Websocket {
	switch {
	case c.checkUnusable("Close"):
		return c
	case len(code) > 1:
		c.chain.fail("\nunexpected multiple code arguments passed to Close")
		return c
	}
	return c.CloseWithBytes(nil, code...)
}

// CloseWithBytes cleanly closes the underlying WebSocket connection
// by sending given slice of bytes as a close message and then waiting
// (with timeout) for the server to close the connection.
//
// WebSocket close code may be optionally specified.
// If not, then "1000 - Normal Closure" will be used.
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// It's okay to call this function multiple times.
//
// Example:
//  conn := resp.Connection()
//  conn.CloseWithBytes([]byte("bye!"), websocket.CloseGoingAway)
func (c *Websocket) CloseWithBytes(b []byte, code ...int) *Websocket {
	switch {
	case c.checkUnusable("CloseWithBytes"):
		return c
	case len(code) > 1:
		c.chain.fail(
			"\nunexpected multiple code arguments passed to CloseWithBytes")
		return c
	}

	c.WriteMessage(websocket.CloseMessage, b, code...)

	return c
}

// CloseWithJSON cleanly closes the underlying WebSocket connection
// by sending given object (marshaled using json.Marshal()) as a close message
// and then waiting (with timeout) for the server to close the connection.
//
// WebSocket close code may be optionally specified.
// If not, then "1000 - Normal Closure" will be used.
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// It's okay to call this function multiple times.
//
// Example:
//  type MyJSON struct {
//    Foo int `json:"foo"`
//  }
//
//  conn := resp.Connection()
//  conn.CloseWithJSON(MyJSON{Foo: 123}, websocket.CloseUnsupportedData)
func (c *Websocket) CloseWithJSON(
	object interface{}, code ...int,
) *Websocket {
	switch {
	case c.checkUnusable("CloseWithJSON"):
		return c
	case len(code) > 1:
		c.chain.fail(
			"\nunexpected multiple code arguments passed to CloseWithJSON")
		return c
	}

	b, err := json.Marshal(object)
	if err != nil {
		c.chain.fail(err.Error())
		return c
	}
	return c.CloseWithBytes(b, code...)
}

// CloseWithText cleanly closes the underlying WebSocket connection
// by sending given text as a close message and then waiting (with timeout)
// for the server to close the connection.
//
// WebSocket close code may be optionally specified.
// If not, then "1000 - Normal Closure" will be used.
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// It's okay to call this function multiple times.
//
// Example:
//  conn := resp.Connection()
//  conn.CloseWithText("bye!")
func (c *Websocket) CloseWithText(s string, code ...int) *Websocket {
	switch {
	case c.checkUnusable("CloseWithText"):
		return c
	case len(code) > 1:
		c.chain.fail(
			"\nunexpected multiple code arguments passed to CloseWithText")
		return c
	}
	return c.CloseWithBytes([]byte(s), code...)
}

// WriteMessage writes to the underlying WebSocket connection a message
// of given type with given content.
// Additionally, WebSocket close code may be specified for close messages.
//
// WebSocket message types are defined in RFC 6455, section 11.8.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// WebSocket close codes are defined in RFC 6455, section 11.7.
// See also https://godoc.org/github.com/gorilla/websocket#pkg-constants
//
// Example:
//  conn := resp.Connection()
//  conn.WriteMessage(websocket.CloseMessage, []byte("Namárië..."))
func (c *Websocket) WriteMessage(
	typ int, content []byte, closeCode ...int,
) *Websocket {
	if c.checkUnusable("WriteMessage") {
		return c
	}

	switch typ {
	case websocket.TextMessage, websocket.BinaryMessage:
		c.printWrite(typ, content, 0)
	case websocket.CloseMessage:
		if len(closeCode) > 1 {
			c.chain.fail("\nunexpected multiple closeCode arguments " +
				"passed to WriteMessage")
			return c
		}

		code := websocket.CloseNormalClosure
		if len(closeCode) > 0 {
			code = closeCode[0]
		}

		c.printWrite(typ, content, code)

		content = websocket.FormatCloseMessage(code, string(content))
	default:
		c.chain.fail("\nunexpected WebSocket message type '%s' "+
			"passed to WriteMessage", wsMessageTypeName(typ))
		return c
	}

	if !c.setWriteDeadline() {
		return c
	}
	if err := c.conn.WriteMessage(typ, content); err != nil {
		c.chain.fail(
			"\nexpected write into WebSocket connection, "+
				"but got failure: %s", err.Error())
	}

	return c
}

// WriteBytesBinary is a shorthand for c.WriteMessage(websocket.BinaryMessage, b).
func (c *Websocket) WriteBytesBinary(b []byte) *Websocket {
	if c.checkUnusable("WriteBytesBinary") {
		return c
	}
	return c.WriteMessage(websocket.BinaryMessage, b)
}

// WriteBytesText is a shorthand for c.WriteMessage(websocket.TextMessage, b).
func (c *Websocket) WriteBytesText(b []byte) *Websocket {
	if c.checkUnusable("WriteBytesText") {
		return c
	}
	return c.WriteMessage(websocket.TextMessage, b)
}

// WriteText is a shorthand for
// c.WriteMessage(websocket.TextMessage, []byte(s)).
func (c *Websocket) WriteText(s string) *Websocket {
	if c.checkUnusable("WriteText") {
		return c
	}
	return c.WriteMessage(websocket.TextMessage, []byte(s))
}

// WriteJSON writes to the underlying WebSocket connection given object,
// marshaled using json.Marshal().
func (c *Websocket) WriteJSON(object interface{}) *Websocket {
	if c.checkUnusable("WriteJSON") {
		return c
	}

	b, err := json.Marshal(object)
	if err != nil {
		c.chain.fail(err.Error())
		return c
	}

	return c.WriteMessage(websocket.TextMessage, b)
}

func (c *Websocket) checkUnusable(where string) bool {
	switch {
	case c.chain.failed():
		return true
	case c.conn == nil:
		c.chain.fail("\nunexpected %s call for failed WebSocket connection",
			where)
		return true
	case c.isClosed:
		c.chain.fail("\nunexpected %s call for closed WebSocket connection",
			where)
		return true
	}
	return false
}

func (c *Websocket) setWriteDeadline() bool {
	deadline := infiniteTime
	if c.writeTimeout != noDuration {
		deadline = time.Now().Add(c.writeTimeout)
	}
	if err := c.conn.SetWriteDeadline(deadline); err != nil {
		c.chain.fail(
			"\nunexpected failure when setting "+
				"write WebSocket connection deadline: %s", err.Error())
		return false
	}
	return true
}

func (c *Websocket) printWrite(typ int, content []byte, closeCode int) {
	for _, printer := range c.config.Printers {
		if p, ok := printer.(WebsocketPrinter); ok {
			p.WebsocketWrite(typ, content, closeCode)
		}
	}
}
