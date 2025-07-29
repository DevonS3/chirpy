package main

import (
	"net/http"
)

func main() {
	servMux := http.NewServeMux()
	var server http.Server

	server.Handler = servMux
	server.Addr = ":8080"

	server.ListenAndServe()
}
