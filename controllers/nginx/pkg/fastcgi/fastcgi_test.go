/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fastcgi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

type dummyHandler struct {
}

func (d *dummyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	code := req.Header.Get(CodeHeader)
	format := req.Header.Get(FormatHeader)

	if format == "" || format == "*/*" {
		format = "text/html"
	}

	httpCode, err := strconv.Atoi(code)
	if err != nil {
		httpCode = 404
	}

	de := []byte(code)
	w.Header().Set(ContentTypeHeader, format)
	w.WriteHeader(httpCode)
	w.Write(de)
}

func TestErrorHandler(t *testing.T) {
	tt := []struct {
		name      string
		code      int
		format    string
		endpoints string
	}{
		{name: "404 text/html", code: 404, format: "text/html", endpoints: "127.0.0.1:80"},
		{name: "503 text/html", code: 503, format: "text/html"},
		{name: "404 application/json", code: 404, format: "application/json", endpoints: "127.0.0.1:80"},
	}

	server := httptest.NewServer(&dummyHandler{})
	defer server.Close()
	hp := strings.Replace(server.URL, "http://", "", -1)

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", server.URL, nil)
			req.Header.Add(CodeHeader, fmt.Sprintf("%v", tc.code))
			req.Header.Add(FormatHeader, tc.format)
			req.Header.Add(EndpointsHeader, hp)
			if err != nil {
				t.Fatalf("could not created request: %v", err)
			}
			w := httptest.NewRecorder()
			serveError(w, req)

			res := w.Result()
			defer res.Body.Close()

			b, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("could not read response: %v", err)
			}

			if res.StatusCode != tc.code {
				t.Errorf("expected status %v; got %v", tc.code, res.StatusCode)
			}

			ct := res.Header.Get(ContentTypeHeader)
			if ct != tc.format {
				t.Errorf("expected content type %v; got %v", tc.format, ct)
			}

			if len(b) == 0 {
				t.Fatalf("unexpected empty body")
			}

			if string(b) != strconv.Itoa(tc.code) {
				t.Fatalf("body: %v", string(b))
			}
		})
	}
}
