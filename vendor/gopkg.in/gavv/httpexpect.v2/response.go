package httpexpect

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ajg/form"
	"github.com/gorilla/websocket"
)

// StatusRange is enum for response status ranges.
type StatusRange int

const (
	// Status1xx defines "Informational" status codes.
	Status1xx StatusRange = 100

	// Status2xx defines "Success" status codes.
	Status2xx StatusRange = 200

	// Status3xx defines "Redirection" status codes.
	Status3xx StatusRange = 300

	// Status4xx defines "Client Error" status codes.
	Status4xx StatusRange = 400

	// Status5xx defines "Server Error" status codes.
	Status5xx StatusRange = 500
)

// Response provides methods to inspect attached http.Response object.
type Response struct {
	config    Config
	chain     chain
	resp      *http.Response
	content   []byte
	cookies   []*http.Cookie
	websocket *websocket.Conn
	rtt       *time.Duration
}

// NewResponse returns a new Response given a reporter used to report
// failures and http.Response to be inspected.
//
// Both reporter and response should not be nil. If response is nil,
// failure is reported.
//
// If rtt is given, it defines response round-trip time to be reported
// by response.RoundTripTime().
func NewResponse(
	reporter Reporter, response *http.Response, rtt ...time.Duration,
) *Response {
	var rttPtr *time.Duration
	if len(rtt) > 0 {
		rttPtr = &rtt[0]
	}
	return makeResponse(responseOpts{
		chain:    makeChain(reporter),
		response: response,
		rtt:      rttPtr,
	})
}

type responseOpts struct {
	config    Config
	chain     chain
	response  *http.Response
	websocket *websocket.Conn
	rtt       *time.Duration
}

func makeResponse(opts responseOpts) *Response {
	var content []byte
	var cookies []*http.Cookie
	if opts.response != nil {
		content = getContent(&opts.chain, opts.response)
		cookies = opts.response.Cookies()
	} else {
		opts.chain.fail("expected non-nil response")
	}
	return &Response{
		config:    opts.config,
		chain:     opts.chain,
		resp:      opts.response,
		content:   content,
		cookies:   cookies,
		websocket: opts.websocket,
		rtt:       opts.rtt,
	}
}

func getContent(chain *chain, resp *http.Response) []byte {
	if resp.Body == nil {
		return []byte{}
	}

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		chain.fail(err.Error())
		return nil
	}

	return content
}

// Raw returns underlying http.Response object.
// This is the value originally passed to NewResponse.
func (r *Response) Raw() *http.Response {
	return r.resp
}

// RoundTripTime returns a new Duration object that may be used to inspect
// the round-trip time.
//
// The returned duration is a time interval starting just before request is
// sent and ending right after response is received (handshake finished for
// WebSocket request), retrieved from a monotonic clock source.
//
// Example:
//  resp := NewResponse(t, response, time.Duration(10000000))
//  resp.RoundTripTime().Lt(10 * time.Millisecond)
func (r *Response) RoundTripTime() *Duration {
	return &Duration{r.chain, r.rtt}
}

// Deprecated: use RoundTripTime instead.
func (r *Response) Duration() *Number {
	if r.rtt == nil {
		return &Number{r.chain, 0}
	}
	return &Number{r.chain, float64(*r.rtt)}
}

// Status succeeds if response contains given status code.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Status(http.StatusOK)
func (r *Response) Status(status int) *Response {
	if r.chain.failed() {
		return r
	}
	r.checkEqual("status", statusCodeText(status), statusCodeText(r.resp.StatusCode))
	return r
}

// StatusRange succeeds if response status belongs to given range.
//
// Supported ranges:
//  - Status1xx - Informational
//  - Status2xx - Success
//  - Status3xx - Redirection
//  - Status4xx - Client Error
//  - Status5xx - Server Error
//
// See https://en.wikipedia.org/wiki/List_of_HTTP_status_codes.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.StatusRange(Status2xx)
func (r *Response) StatusRange(rn StatusRange) *Response {
	if r.chain.failed() {
		return r
	}

	status := statusCodeText(r.resp.StatusCode)

	actual := statusRangeText(r.resp.StatusCode)
	expected := statusRangeText(int(rn))

	if actual == "" || actual != expected {
		if actual == "" {
			r.chain.fail("\nexpected status from range:\n %q\n\nbut got:\n %q",
				expected, status)
		} else {
			r.chain.fail(
				"\nexpected status from range:\n %q\n\nbut got:\n %q (%q)",
				expected, actual, status)
		}
	}

	return r
}

func statusCodeText(code int) string {
	if s := http.StatusText(code); s != "" {
		return strconv.Itoa(code) + " " + s
	}
	return strconv.Itoa(code)
}

func statusRangeText(code int) string {
	switch {
	case code >= 100 && code < 200:
		return "1xx Informational"
	case code >= 200 && code < 300:
		return "2xx Success"
	case code >= 300 && code < 400:
		return "3xx Redirection"
	case code >= 400 && code < 500:
		return "4xx Client Error"
	case code >= 500 && code < 600:
		return "5xx Server Error"
	default:
		return ""
	}
}

// Headers returns a new Object that may be used to inspect header map.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Headers().Value("Content-Type").String().Equal("application-json")
func (r *Response) Headers() *Object {
	var value map[string]interface{}
	if !r.chain.failed() {
		value, _ = canonMap(&r.chain, r.resp.Header)
	}
	return &Object{r.chain, value}
}

// Header returns a new String object that may be used to inspect given header.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Header("Content-Type").Equal("application-json")
//  resp.Header("Date").DateTime().Le(time.Now())
func (r *Response) Header(header string) *String {
	value := ""
	if !r.chain.failed() {
		value = r.resp.Header.Get(header)
	}
	return &String{r.chain, value}
}

// Cookies returns a new Array object with all cookie names set by this response.
// Returned Array contains a String value for every cookie name.
//
// Note that this returns only cookies set by Set-Cookie headers of this response.
// It doesn't return session cookies from previous responses, which may be stored
// in a cookie jar.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Cookies().Contains("session")
func (r *Response) Cookies() *Array {
	if r.chain.failed() {
		return &Array{r.chain, nil}
	}
	names := []interface{}{}
	for _, c := range r.cookies {
		names = append(names, c.Name)
	}
	return &Array{r.chain, names}
}

// Cookie returns a new Cookie object that may be used to inspect given cookie
// set by this response.
//
// Note that this returns only cookies set by Set-Cookie headers of this response.
// It doesn't return session cookies from previous responses, which may be stored
// in a cookie jar.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Cookie("session").Domain().Equal("example.com")
func (r *Response) Cookie(name string) *Cookie {
	if r.chain.failed() {
		return &Cookie{r.chain, nil}
	}
	names := []string{}
	for _, c := range r.cookies {
		if c.Name == name {
			return &Cookie{r.chain, c}
		}
		names = append(names, c.Name)
	}
	r.chain.fail("\nexpected response with cookie:\n %q\n\nbut got only cookies:\n%s",
		name, dumpValue(names))
	return &Cookie{r.chain, nil}
}

// Websocket returns Websocket object that can be used to interact with
// WebSocket server.
//
// May be called only if the WithWebsocketUpgrade was called on the request.
// That is responsibility of the caller to explicitly close the websocket after use.
//
// Example:
//  req := NewRequest(config, "GET", "/path")
//  req.WithWebsocketUpgrade()
//  ws := req.Expect().Websocket()
//  defer ws.Disconnect()
func (r *Response) Websocket() *Websocket {
	if !r.chain.failed() && r.websocket == nil {
		r.chain.fail("\nunexpected Websocket call for non-WebSocket response")
	}
	return makeWebsocket(r.config, r.chain, r.websocket)
}

// Body returns a new String object that may be used to inspect response body.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Body().NotEmpty()
//  resp.Body().Length().Equal(100)
func (r *Response) Body() *String {
	return &String{r.chain, string(r.content)}
}

// NoContent succeeds if response contains empty Content-Type header and
// empty body.
func (r *Response) NoContent() *Response {
	if r.chain.failed() {
		return r
	}

	contentType := r.resp.Header.Get("Content-Type")

	r.checkEqual("\"Content-Type\" header", "", contentType)
	r.checkEqual("body", "", string(r.content))

	return r
}

// ContentType succeeds if response contains Content-Type header with given
// media type and charset.
//
// If charset is omitted, and mediaType is non-empty, Content-Type header
// should contain empty or utf-8 charset.
//
// If charset is omitted, and mediaType is also empty, Content-Type header
// should contain no charset.
func (r *Response) ContentType(mediaType string, charset ...string) *Response {
	r.checkContentType(mediaType, charset...)
	return r
}

// ContentEncoding succeeds if response has exactly given Content-Encoding list.
// Common values are empty, "gzip", "compress", "deflate", "identity" and "br".
func (r *Response) ContentEncoding(encoding ...string) *Response {
	if r.chain.failed() {
		return r
	}
	r.checkEqual("\"Content-Encoding\" header", encoding, r.resp.Header["Content-Encoding"])
	return r
}

// TransferEncoding succeeds if response contains given Transfer-Encoding list.
// Common values are empty, "chunked" and "identity".
func (r *Response) TransferEncoding(encoding ...string) *Response {
	if r.chain.failed() {
		return r
	}
	r.checkEqual("\"Transfer-Encoding\" header", encoding, r.resp.TransferEncoding)
	return r
}

// ContentOpts define parameters for matching the response content parameters.
type ContentOpts struct {
	// The media type Content-Type part, e.g. "application/json"
	MediaType string
	// The character set Content-Type part, e.g. "utf-8"
	Charset string
}

// Text returns a new String object that may be used to inspect response body.
//
// Text succeeds if response contains "text/plain" Content-Type header
// with empty or "utf-8" charset.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Text().Equal("hello, world!")
//  resp.Text(ContentOpts{
//    MediaType: "text/plain",
//  }).Equal("hello, world!")
func (r *Response) Text(opts ...ContentOpts) *String {
	var content string

	if !r.chain.failed() && r.checkContentOpts(opts, "text/plain") {
		content = string(r.content)
	}

	return &String{r.chain, content}
}

// Form returns a new Object that may be used to inspect form contents
// of response.
//
// Form succeeds if response contains "application/x-www-form-urlencoded"
// Content-Type header and if form may be decoded from response body.
// Decoding is performed using https://github.com/ajg/form.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.Form().Value("foo").Equal("bar")
//  resp.Form(ContentOpts{
//    MediaType: "application/x-www-form-urlencoded",
//  }).Value("foo").Equal("bar")
func (r *Response) Form(opts ...ContentOpts) *Object {
	object := r.getForm(opts...)
	return &Object{r.chain, object}
}

func (r *Response) getForm(opts ...ContentOpts) map[string]interface{} {
	if r.chain.failed() {
		return nil
	}

	if !r.checkContentOpts(opts, "application/x-www-form-urlencoded", "") {
		return nil
	}

	decoder := form.NewDecoder(bytes.NewReader(r.content))

	var object map[string]interface{}
	if err := decoder.Decode(&object); err != nil {
		r.chain.fail(err.Error())
		return nil
	}

	return object
}

// JSON returns a new Value object that may be used to inspect JSON contents
// of response.
//
// JSON succeeds if response contains "application/json" Content-Type header
// with empty or "utf-8" charset and if JSON may be decoded from response body.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.JSON().Array().Elements("foo", "bar")
//  resp.JSON(ContentOpts{
//    MediaType: "application/json",
//  }).Array.Elements("foo", "bar")
func (r *Response) JSON(opts ...ContentOpts) *Value {
	value := r.getJSON(opts...)
	return &Value{r.chain, value}
}

func (r *Response) getJSON(opts ...ContentOpts) interface{} {
	if r.chain.failed() {
		return nil
	}

	if !r.checkContentOpts(opts, "application/json") {
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(r.content, &value); err != nil {
		r.chain.fail(err.Error())
		return nil
	}

	return value
}

// JSONP returns a new Value object that may be used to inspect JSONP contents
// of response.
//
// JSONP succeeds if response contains "application/javascript" Content-Type
// header with empty or "utf-8" charset and response body of the following form:
//  callback(<valid json>);
// or:
//  callback(<valid json>)
//
// Whitespaces are allowed.
//
// Example:
//  resp := NewResponse(t, response)
//  resp.JSONP("myCallback").Array().Elements("foo", "bar")
//  resp.JSONP("myCallback", ContentOpts{
//    MediaType: "application/javascript",
//  }).Array.Elements("foo", "bar")
func (r *Response) JSONP(callback string, opts ...ContentOpts) *Value {
	value := r.getJSONP(callback, opts...)
	return &Value{r.chain, value}
}

var (
	jsonp = regexp.MustCompile(`^\s*([^\s(]+)\s*\((.*)\)\s*;*\s*$`)
)

func (r *Response) getJSONP(callback string, opts ...ContentOpts) interface{} {
	if r.chain.failed() {
		return nil
	}

	if !r.checkContentOpts(opts, "application/javascript") {
		return nil
	}

	m := jsonp.FindSubmatch(r.content)
	if len(m) != 3 || string(m[1]) != callback {
		r.chain.fail(
			"\nexpected JSONP body in form of:\n \"%s(<valid json>)\"\n\nbut got:\n %q\n",
			callback,
			string(r.content))
		return nil
	}

	var value interface{}
	if err := json.Unmarshal(m[2], &value); err != nil {
		r.chain.fail(err.Error())
		return nil
	}

	return value
}

func (r *Response) checkContentOpts(
	opts []ContentOpts, expectedType string, expectedCharset ...string,
) bool {
	if len(opts) != 0 {
		if opts[0].MediaType != "" {
			expectedType = opts[0].MediaType
		}
		if opts[0].Charset != "" {
			expectedCharset = []string{opts[0].Charset}
		}
	}
	return r.checkContentType(expectedType, expectedCharset...)
}

func (r *Response) checkContentType(expectedType string, expectedCharset ...string) bool {
	if r.chain.failed() {
		return false
	}

	contentType := r.resp.Header.Get("Content-Type")

	if expectedType == "" && len(expectedCharset) == 0 {
		if contentType == "" {
			return true
		}
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		r.chain.fail("\ngot invalid \"Content-Type\" header %q", contentType)
		return false
	}

	if mediaType != expectedType {
		r.chain.fail(
			"\nexpected \"Content-Type\" header with %q media type,"+
				"\nbut got %q", expectedType, mediaType)
		return false
	}

	charset := params["charset"]

	if len(expectedCharset) == 0 {
		if charset != "" && !strings.EqualFold(charset, "utf-8") {
			r.chain.fail(
				"\nexpected \"Content-Type\" header with \"utf-8\" or empty charset,"+
					"\nbut got %q", charset)
			return false
		}
	} else {
		if !strings.EqualFold(charset, expectedCharset[0]) {
			r.chain.fail(
				"\nexpected \"Content-Type\" header with %q charset,"+
					"\nbut got %q", expectedCharset[0], charset)
			return false
		}
	}

	return true
}

func (r *Response) checkEqual(what string, expected, actual interface{}) {
	if !reflect.DeepEqual(expected, actual) {
		r.chain.fail("\nexpected %s equal to:\n%s\n\nbut got:\n%s", what,
			dumpValue(expected), dumpValue(actual))
	}
}
