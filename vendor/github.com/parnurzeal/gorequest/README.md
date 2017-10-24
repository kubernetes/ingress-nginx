GoRequest
=========

GoRequest -- Simplified HTTP client ( inspired by famous SuperAgent lib in Node.js )

![GopherGoRequest](https://raw.githubusercontent.com/parnurzeal/gorequest/gh-pages/images/Gopher_GoRequest_400x300.jpg)

#### "Shooting Requests like a Machine Gun" - Gopher

Sending request would never been fun and easier than this. It comes with lots of feature:

* Get/Post/Put/Head/Delete/Patch/Options
* Set - simple header setting
* JSON - made it simple with JSON string as a parameter
* Multipart-Support - send data and files as multipart request
* Proxy - sending request via proxy
* Timeout - setting timeout for a request
* TLSClientConfig - taking control over tls where at least you can disable security check for https
* RedirectPolicy
* Cookie - setting cookies for your request
* CookieJar - automatic in-memory cookiejar
* BasicAuth - setting basic authentication header
* more to come..

## Installation

```bash
$ go get github.com/parnurzeal/gorequest
```

## Documentation
See [Go Doc](http://godoc.org/github.com/parnurzeal/gorequest) or [Go Walker](http://gowalker.org/github.com/parnurzeal/gorequest) for usage and details.

## Status

[![Drone Build Status](https://drone.io/github.com/jmcvetta/restclient/status.png)](https://drone.io/github.com/parnurzeal/gorequest/latest)
[![Travis Build Status](https://travis-ci.org/parnurzeal/gorequest.svg?branch=master)](https://travis-ci.org/parnurzeal/gorequest)

## Why should you use GoRequest?

GoRequest makes thing much more simple for you, making http client more awesome and fun like SuperAgent + golang style usage.

This is what you normally do for a simple GET without GoRequest:

```go
resp, err := http.Get("http://example.com/")
```

With GoRequest:

```go
request := gorequest.New()
resp, body, errs := request.Get("http://example.com/").End()
```

Or below if you don't want to reuse it for other requests.

```go
resp, body, errs := gorequest.New().Get("http://example.com/").End()
```

How about getting control over HTTP client headers, redirect policy, and etc. Things is getting more complicated in golang. You need to create a Client, setting header in different command, ... to do just only one __GET__

```go
client := &http.Client{
  CheckRedirect: redirectPolicyFunc,
}

req, err := http.NewRequest("GET", "http://example.com", nil)

req.Header.Add("If-None-Match", `W/"wyzzy"`)
resp, err := client.Do(req)
```

Why making things ugly while you can just do as follows:

```go
request := gorequest.New()
resp, body, errs := request.Get("http://example.com").
  RedirectPolicy(redirectPolicyFunc).
  Set("If-None-Match", `W/"wyzzy"`).
  End()
```

__DELETE__, __HEAD__, __POST__, __PUT__, __PATCH__ are now supported and can be used the same way as __GET__:

```go
request := gorequest.New()
resp, body, errs := request.Post("http://example.com").End()
// PUT -> request.Put("http://example.com").End()
// DELETE -> request.Delete("http://example.com").End()
// HEAD -> request.Head("http://example.com").End()
// ANYTHING -> request.CustomMethod("TRACE", "http://example.com").End()
```

### JSON

For a __JSON POST__ with standard libraries, you might need to marshal map data structure to json format, setting header to 'application/json' (and other headers if you need to) and declare http.Client. So, you code become longer and hard to maintain:

```go
m := map[string]interface{}{
  "name": "backy",
  "species": "dog",
}
mJson, _ := json.Marshal(m)
contentReader := bytes.NewReader(mJson)
req, _ := http.NewRequest("POST", "http://example.com", contentReader)
req.Header.Set("Content-Type", "application/json")
req.Header.Set("Notes","GoRequest is coming!")
client := &http.Client{}
resp, _ := client.Do(req)
```

Compared to our GoRequest version, JSON is for sure a default. So, it turns out to be just one simple line!:

```go
request := gorequest.New()
resp, body, errs := request.Post("http://example.com").
  Set("Notes","gorequst is coming!").
  Send(`{"name":"backy", "species":"dog"}`).
  End()
```

Moreover, it also supports struct type. So, you can have a fun __Mix & Match__ sending the different data types for your request:

```go
type BrowserVersionSupport struct {
  Chrome string
  Firefox string
}
ver := BrowserVersionSupport{ Chrome: "37.0.2041.6", Firefox: "30.0" }
request := gorequest.New()
resp, body, errs := request.Post("http://version.com/update").
  Send(ver).
  Send(`{"Safari":"5.1.10"}`).
  End()
```

Not only for Send() but Query() is also supported. Just give it a try! :)

## Callback

Moreover, GoRequest also supports callback function. This gives you much more flexibility on using it. You can use it any way to match your own style!
Let's see a bit of callback example:

```go
func printStatus(resp gorequest.Response, body string, errs []error){
  fmt.Println(resp.Status)
}
gorequest.New().Get("http://example.com").End(printStatus)
```

## Multipart/Form-Data

You can specify the content-type of the request to type `multipart` to send all data as `multipart/form-data`. This feature also allows you to send (multiple) files! Check the examples below!

```go
gorequest.New().Post("http://example.com/").
  Type("multipart").
  Send(`{"query1":"test"}`).
  End()
```

The `SendFile` function accepts `strings` as path to a file, `[]byte` slice or even a `os.File`! You can also combine them to send multiple files with either custom name and/or custom fieldname:

```go
          f, _ := filepath.Abs("./file2.txt")
bytesOfFile, _ := ioutil.ReadFile(f)

gorequest.New().Post("http://example.com/").
  Type("multipart").
  SendFile("./file1.txt").
  SendFile(bytesOfFile, "file2.txt", "my_file_fieldname").
  End()
```

Check the docs for `SendFile` to get more information about the types of arguments.

## Proxy

In the case when you are behind proxy, GoRequest can handle it easily with Proxy func:

```go
request := gorequest.New().Proxy("http://proxy:999")
resp, body, errs := request.Get("http://example-proxy.com").End()
// To reuse same client with no_proxy, use empty string:
resp, body, errs = request.Proxy("").Get("http://example-no-proxy.com").End()
```

## Basic Authentication

To add a basic authentication header:

```go
request := gorequest.New().SetBasicAuth("username", "password")
resp, body, errs := request.Get("http://example-proxy.com").End()
```

## Timeout

Timeout can be set in any time duration using time package:

```go
request := gorequest.New().Timeout(2*time.Millisecond)
resp, body, errs:= request.Get("http://example.com").End()
```

Timeout func defines both dial + read/write timeout to the specified time parameter.

## EndBytes

Thanks to @jaytaylor, we now have EndBytes to use when you want the body as bytes.

The callbacks work the same way as with `End`, except that a byte array is used instead of a string.

```go
resp, bodyBytes, errs := gorequest.New().Get("http://example.com/").EndBytes()
```

## EndStruct

We now have EndStruct to use when you want the body as struct.

The callbacks work the same way as with `End`, except that a struct is used instead of a string.

Supposing the URL **http://example.com/** returns the body `{"hey":"you"}`

```go
heyYou struct {
  Hey string `json:"hey"`
}

var heyYou heyYou

resp, _, errs := gorequest.New().Get("http://example.com/").EndStruct(&heyYou)
```

## Retry

Supposing you need retry 3 times, with 5 seconds between each attempt when gets a BadRequest or a InternalServerError

```go
request := gorequest.New()
resp, body, errs := request.Get("http://example.com/").
                    Retry(3, 5 * time.seconds, http.StatusBadRequest, http.StatusInternalServerError).
                    End()
```

## Handling Redirects

Redirects can be handled with RedirectPolicy which behaves similarly to
net/http Client's [CheckRedirect
function](https://golang.org/pkg/net/http#Client). Simply specify a function
which takes the Request about to be made and a slice of previous Requests in
order of oldest first. When this function returns an error, the Request is not
made.

For example to redirect only to https endpoints:

```go
request := gorequest.New()
resp, body, errs := request.Get("http://example.com/").
                    RedirectPolicy(func(req Request, via []*Request) error {
                      if req.URL.Scheme != "https" {
                        return http.ErrUseLastResponse
                      }
                    }).
                    End()
```

## Debug

For debugging, GoRequest leverages `httputil` to dump details of every request/response. (Thanks to @dafang)

You can just use `SetDebug` or environment variable `GOREQUEST_DEBUG=0|1` to enable/disable debug mode and `SetLogger` to set your own choice of logger.

Thanks to @QuentinPerez, we can see even how gorequest is compared to CURL by using `SetCurlCommand`.

## Noted
As the underlying gorequest is based on http.Client in most use cases, gorequest.New() should be called once and reuse gorequest as much as possible.

## Contributing to GoRequest:

If you find any improvement or issue you want to fix, feel free to send me a pull request with testing.

Thanks to all contributors thus far:


|   Contributors                        |
|---------------------------------------|
| https://github.com/alaingilbert       |
| https://github.com/austinov           |
| https://github.com/coderhaoxin        |
| https://github.com/codegoalie         |
| https://github.com/dafang             |
| https://github.com/davyzhang          |
| https://github.com/dickeyxxx          |
| https://github.com/figlief            |
| https://github.com/fraenky8           |
| https://github.com/franciscocpg       |
| https://github.com/heytitle           |
| https://github.com/hownowstephen      |
| https://github.com/kemadz             |
| https://github.com/killix             |
| https://github.com/jaytaylor          |
| https://github.com/na-ga              |
| https://github.com/piotrmiskiewicz    |
| https://github.com/pencil001          |
| https://github.com/pkopac             |
| https://github.com/quangbuule         |
| https://github.com/QuentinPerez       |
| https://github.com/smallnest          |
| https://github.com/WaveCutz           |
| https://github.com/xild               |
| https://github.com/yangmls            |
| https://github.com/6david9            |


Also, co-maintainer is needed here. If anyone is interested, please email me (parnurzeal at gmail.com)

## Credits

* Renee French - the creator of Gopher mascot
* [Wisi Mongkhonsrisawat](https://www.facebook.com/puairw) for providing an awesome GoRequest's Gopher image :)

## License

GoRequest is MIT License.
