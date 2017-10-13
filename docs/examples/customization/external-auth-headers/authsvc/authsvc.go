package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

// Sample authentication service returning several HTTP headers in response
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.ContainsAny(r.Header.Get("User"), "internal") {
			w.Header().Add("UserID", strconv.Itoa(rand.Int()))
			w.Header().Add("UserRole", "admin")
			w.Header().Add("Other", "not used")
			fmt.Fprint(w, "ok")
		} else {
			rc := http.StatusForbidden
			if c := r.URL.Query().Get("code"); len(c) > 0 {
				c, _ := strconv.Atoi(c)
				if c > 0 && c < 600 {
					rc = c
				}
			}

			w.WriteHeader(rc)
			fmt.Fprint(w, "unauthorized")
		}
	})
	http.ListenAndServe(":8080", nil)
}
