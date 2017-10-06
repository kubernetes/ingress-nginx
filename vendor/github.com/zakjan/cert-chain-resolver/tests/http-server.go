package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	fs := http.FileServer(http.Dir(os.Args[1]))
	http.Handle("/", fs)

	err := http.ListenAndServe(":http", nil)

	if err != nil {
		fmt.Println(err)
	}
}
