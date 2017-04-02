package main

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "UserID: %s, UserRole: %s", r.Header.Get("UserID"), r.Header.Get("UserRole"))
}

// Sample  "echo" service displaying UserID and UserRole HTTP request headers
func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
