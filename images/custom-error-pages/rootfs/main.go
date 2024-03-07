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
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// FormatHeader name of the header used to extract the format
	FormatHeader = "X-Format"

	// CodeHeader name of the header used as source of the HTTP status code to return
	CodeHeader = "X-Code"

	// ContentType name of the header that defines the format of the reply
	ContentType = "Content-Type"

	// OriginalURI name of the header with the original URL from NGINX
	OriginalURI = "X-Original-URI"

	// Namespace name of the header that contains information about the Ingress namespace
	Namespace = "X-Namespace"

	// IngressName name of the header that contains the matched Ingress
	IngressName = "X-Ingress-Name"

	// ServiceName name of the header that contains the matched Service in the Ingress
	ServiceName = "X-Service-Name"

	// ServicePort name of the header that contains the matched Service port in the Ingress
	ServicePort = "X-Service-Port"

	// RequestId is a unique ID that identifies the request - same as for backend service
	RequestId = "X-Request-ID"

	// ErrFilesPathEnvVar is the name of the environment variable indicating
	// the location on disk of files served by the handler.
	ErrFilesPathEnvVar = "ERROR_FILES_PATH"

	// DefaultFormatEnvVar is the name of the environment variable indicating
	// the default error MIME type that should be returned if either the
	// client does not specify an Accept header, or the Accept header provided
	// cannot be mapped to a file extension.
	DefaultFormatEnvVar = "DEFAULT_RESPONSE_FORMAT"

	//  IsMetricsExportEnvVar is the name of the environment variable indicating
	// whether or not to export /metrics and /debug/vars.
	IsMetricsExportEnvVar = "IS_METRICS_EXPORT"

	// MetricsPortEnvVar is the name of the environment variable indicating
	// the port on which to export /metrics and /debug/vars.
	MetricsPortEnvVar = "METRICS_PORT"

	// CustomErrorPagesPort is the port on which to listen to serve custom error pages.
	CustomErrorPagesPort = "8080"
)

func init() {
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(requestDuration)
}

func main() {
	listeners := createListeners()
	startListeners(listeners)
}

func createListeners() []Listener {
	errFilesPath := "/www"
	if os.Getenv(ErrFilesPathEnvVar) != "" {
		errFilesPath = os.Getenv(ErrFilesPathEnvVar)
	}

	defaultFormat := "text/html"
	if os.Getenv(DefaultFormatEnvVar) != "" {
		defaultFormat = os.Getenv(DefaultFormatEnvVar)
	}

	isExportMetrics := true
	if os.Getenv(IsMetricsExportEnvVar) != "" {
		val, err := strconv.ParseBool(os.Getenv(IsMetricsExportEnvVar))
		if err == nil {
			isExportMetrics = val
		}
	}

	metricsPort := "8080"
	if os.Getenv(MetricsPortEnvVar) != "" {
		metricsPort = os.Getenv(MetricsPortEnvVar)
	}

	var listeners []Listener

	// MUST use NewServerMux when not exporting /metrics because expvar HTTP handler registers
	// against DefaultServerMux as a consequence of importing it in client_golang/prometheus.
	if !isExportMetrics {
		mux := http.NewServeMux()
		mux.HandleFunc("/", errorHandler(errFilesPath, defaultFormat))
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		listeners = append(listeners, Listener{mux, CustomErrorPagesPort})
		return listeners
	}

	// MUST use DefaultServerMux when exporting /metrics to the public because /debug/vars is
	// only available with DefaultServerMux.
	if metricsPort == CustomErrorPagesPort {
		mux := http.DefaultServeMux
		mux.Handle("/metrics", promhttp.Handler())
		mux.HandleFunc("/", errorHandler(errFilesPath, defaultFormat))
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		listeners = append(listeners, Listener{mux, CustomErrorPagesPort})
		return listeners
	}

	// MUST use DefaultServerMux for /metrics and NewServerMux for custom error pages when you
	// wish to expose /metrics only to internal services, because expvar HTTP handler registers
	// against DefaultServerMux.
	metricsMux := http.DefaultServeMux
	metricsMux.Handle("/metrics", promhttp.Handler())

	errorsMux := http.NewServeMux()
	errorsMux.HandleFunc("/", errorHandler(errFilesPath, defaultFormat))
	errorsMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	listeners = append(listeners, Listener{metricsMux, metricsPort}, Listener{errorsMux, CustomErrorPagesPort})
	return listeners
}

func startListeners(listeners []Listener) {
	var wg sync.WaitGroup

	for _, listener := range listeners {
		wg.Add(1)
		go func(l Listener) {
			defer wg.Done()
			err := http.ListenAndServe(fmt.Sprintf(":%s", l.port), l.mux)
			if err != nil {
				log.Fatal(err)
			}
		}(listener)
	}

	wg.Wait()
}

type Listener struct {
	mux  *http.ServeMux
	port string
}

func errorHandler(path, defaultFormat string) func(http.ResponseWriter, *http.Request) {
	defaultExts, err := mime.ExtensionsByType(defaultFormat)
	if err != nil || len(defaultExts) == 0 {
		panic("couldn't get file extension for default format")
	}
	defaultExt := defaultExts[0]

	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ext := defaultExt

		if os.Getenv("DEBUG") != "" {
			w.Header().Set(FormatHeader, r.Header.Get(FormatHeader))
			w.Header().Set(CodeHeader, r.Header.Get(CodeHeader))
			w.Header().Set(ContentType, r.Header.Get(ContentType))
			w.Header().Set(OriginalURI, r.Header.Get(OriginalURI))
			w.Header().Set(Namespace, r.Header.Get(Namespace))
			w.Header().Set(IngressName, r.Header.Get(IngressName))
			w.Header().Set(ServiceName, r.Header.Get(ServiceName))
			w.Header().Set(ServicePort, r.Header.Get(ServicePort))
			w.Header().Set(RequestId, r.Header.Get(RequestId))
		}

		format := r.Header.Get(FormatHeader)
		if format == "" {
			format = defaultFormat
			log.Printf("format not specified. Using %v", format)
		}

		cext, err := mime.ExtensionsByType(format)
		if err != nil {
			log.Printf("unexpected error reading media type extension: %v. Using %v", err, ext)
			format = defaultFormat
		} else if len(cext) == 0 {
			log.Printf("couldn't get media type extension. Using %v", ext)
		} else {
			ext = cext[0]
		}
		w.Header().Set(ContentType, format)

		errCode := r.Header.Get(CodeHeader)
		code, err := strconv.Atoi(errCode)
		if err != nil {
			code = 404
			log.Printf("unexpected error reading return code: %v. Using %v", err, code)
		}
		w.WriteHeader(code)

		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		// special case for compatibility
		if ext == ".htm" {
			ext = ".html"
		}
		file := fmt.Sprintf("%v/%v%v", path, code, ext)
		f, err := os.Open(file)
		if err != nil {
			log.Printf("unexpected error opening file: %v", err)
			scode := strconv.Itoa(code)
			file := fmt.Sprintf("%v/%cxx%v", path, scode[0], ext)
			f, err := os.Open(file)
			if err != nil {
				log.Printf("unexpected error opening file: %v", err)
				http.NotFound(w, r)
				return
			}
			defer f.Close()
			log.Printf("serving custom error response for code %v and format %v from file %v", code, format, file)
			io.Copy(w, f)
			return
		}
		defer f.Close()
		log.Printf("serving custom error response for code %v and format %v from file %v", code, format, file)
		io.Copy(w, f)

		duration := time.Now().Sub(start).Seconds()

		proto := strconv.Itoa(r.ProtoMajor)
		proto = fmt.Sprintf("%s.%s", proto, strconv.Itoa(r.ProtoMinor))

		requestCount.WithLabelValues(proto).Inc()
		requestDuration.WithLabelValues(proto).Observe(duration)
	}
}
