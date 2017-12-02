/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"
)

const (
	// FormatHeader name of the header used to extract the format
	FormatHeader = "X-Format"

	// CodeHeader name of the header used as source of the HTTP statu code to return
	CodeHeader = "X-Code"

	// ContentType name of the header that defines the format of the reply
	ContentType = "Content-Type"
)

func main() {
	path := "/www"
	if os.Getenv("PATH") != "" {
		path = os.Getenv("PATH")
	}
	http.HandleFunc("/", errorHandler(path))
	http.ListenAndServe(fmt.Sprintf(":8080"), nil)
}

func errorHandler(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ext := "html"

		format := r.Header.Get(FormatHeader)
		if format == "" {
			format = "text/html"
			log.Printf("forma not specified. Using %v\n", format)
		}

		mediaType, _, _ := mime.ParseMediaType(format)
		cext, err := mime.ExtensionsByType(mediaType)
		if err != nil {
			log.Printf("unexpected error reading media type extension: %v. Using %v\n", err, ext)
		} else {
			ext = cext[0]
		}
		w.Header().Set(ContentType, format)

		errCode := r.Header.Get(CodeHeader)
		code, err := strconv.Atoi(errCode)
		if err != nil {
			code = 404
			log.Printf("unexpected error reading return code: %v. Using %v\n", err, code)
		}
		w.WriteHeader(code)

		file := fmt.Sprintf("%v/%v%v", path, code, ext)
		f, err := os.Open(file)
		if err != nil {
			log.Printf("unexpected error opening file: %v\n", err)
			scode := strconv.Itoa(code)
			file := fmt.Sprintf("%v/%cxx%v", path, scode[0], ext)
			f, err := os.Open(file)
			if err != nil {
				log.Printf("unexpected error opening file: %v\n", err)
				http.NotFound(w, r)
				return
			}
			defer f.Close()
			log.Printf("serving custom error response for code %v and format %v from file %v\n", code, format, file)
			io.Copy(w, f)
			return
		}
		defer f.Close()
		log.Printf("serving custom error response for code %v and format %v from file %v\n", code, format, file)
		io.Copy(w, f)
	}
}
