/*
Copyright 2019 The Kubernetes Authors.

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

package controller

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/api/admission/v1beta1"
)

type testAdmissionHandler struct{}

func (testAdmissionHandler) HandleAdmission(ar *v1beta1.AdmissionReview) {
	ar.Response = &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("this is a test error")
}

type errorWriter struct{}

func (errorWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("this is a test error")
}

func (errorWriter) Header() http.Header {
	return nil
}

func (errorWriter) WriteHeader(statusCode int) {}

func TestServer(t *testing.T) {
	w := httptest.NewRecorder()
	b := bytes.NewBuffer(nil)
	writeAdmissionReview(b, &v1beta1.AdmissionReview{})

	// Happy path
	r := httptest.NewRequest("GET", "http://test.ns.svc", b)
	NewAdmissionControllerServer(testAdmissionHandler{}).ServeHTTP(w, r)
	ar, err := parseAdmissionReview(codecs.UniversalDeserializer(), w.Body)
	if w.Code != http.StatusOK {
		t.Errorf("when the admission review allows the request, the http status should be OK")
	}
	if err != nil {
		t.Errorf("failed to parse admission response when the admission controller returns a value")
	}
	if !ar.Response.Allowed {
		t.Errorf("when the admission review allows the request, the parsed body returns not allowed")
	}

	// Ensure the code does not panic when failing to handle the request
	NewAdmissionControllerServer(testAdmissionHandler{}).ServeHTTP(errorWriter{}, r)

	w = httptest.NewRecorder()
	NewAdmissionControllerServer(testAdmissionHandler{}).ServeHTTP(w, httptest.NewRequest("GET", "http://test.ns.svc", strings.NewReader("invalid-json")))
	if w.Code != http.StatusBadRequest {
		t.Errorf("when the server fails to read the request, the replied status should be bad request")
	}
}

func TestParseAdmissionReview(t *testing.T) {
	ar, err := parseAdmissionReview(codecs.UniversalDeserializer(), errorReader{})
	if ar != nil {
		t.Errorf("when reading from request fails, no AdmissionRewiew should be returned")
	}
	if err == nil {
		t.Errorf("when reading from request fails, an error should be returned")
	}
}
