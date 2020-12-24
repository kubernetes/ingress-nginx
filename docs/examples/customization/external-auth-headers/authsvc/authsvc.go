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
	"net/http"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/uuid"
)

// Sample authentication service returning several HTTP headers in response
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.ContainsAny(r.Header.Get("User"), "internal") {
			w.Header().Add("UserID", fmt.Sprintf("%v", uuid.NewUUID()))
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
