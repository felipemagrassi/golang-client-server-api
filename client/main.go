package main

import (
	"fmt"
	"io"
	"net/http"
)

func main() {
	req, err := http.Get("http://localhost:8081")
	if err != nil {
		panic(err)
	}

	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)

	fmt.Println(string(body))
}
