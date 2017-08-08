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
	"math/rand"
	"net"
	"net/http"
	"net/http/fcgi"
	"strconv"
	"strings"
)

const (
	// CodeHeader name of the header that indicates the expected response
	// status code
	CodeHeader = "X-Code"
	// FormatHeader name of the header with the expected Content-Type to be
	// sent to the client
	FormatHeader = "X-Format"
	// EndpointsHeader comma separated header that contains the list of
	// endpoints for the default backend
	EndpointsHeader = "X-Endpoints"
	// ContentTypeHeader returns information about the type of the returned body
	ContentTypeHeader = "Content-Type"
)

// ServeError creates a fastcgi handler to serve the custom error pages
func ServeError(l net.Listener) error {
	return fcgi.Serve(l, handler())
}

func handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/", serveError)
	return r
}

func serveError(w http.ResponseWriter, req *http.Request) {
	code := req.Header.Get(CodeHeader)
	format := req.Header.Get(FormatHeader)

	if format == "" || format == "*/*" {
		format = "text/html"
	}

	httpCode, err := strconv.Atoi(code)
	if err != nil {
		httpCode = 404
	}

	de := []byte(fmt.Sprintf("default backend - %v", httpCode))

	w.Header().Set(ContentTypeHeader, format)
	w.WriteHeader(httpCode)

	eh := req.Header.Get(EndpointsHeader)

	if eh == "" {
		w.Write(de)
		return
	}

	eps := strings.Split(eh, ",")

	// TODO: add retries in case of errors
	ep := eps[rand.Intn(len(eps))]
	req, err = http.NewRequest("GET", fmt.Sprintf("http://%v/", ep), nil)
	if err != nil {
		w.Write(de)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.Write(de)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		w.Write(de)
		return
	}
	defer resp.Body.Close()
	w.Write(b)
}
