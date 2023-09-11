package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
)

func hello(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["name"]

	if !ok || len(keys[0]) < 1 {
		fmt.Fprintf(w, "Hello world!")
		return
	}

	key := keys[0]
	fmt.Fprintf(w, "Hello "+key+"!")
}

func main() {
	http.HandleFunc("/hello", hello)

	l, err := net.Listen("tcp", "0.0.0.0:9000") //nolint:gosec // Ignore the gosec error since it's a hello server
	if err != nil {
		panic(err)
	}
	if err := fcgi.Serve(l, nil); err != nil {
		panic(err)
	}
}
