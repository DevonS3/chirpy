package main

import (
	"net/http"
)

func main() {
	servMux := http.NewServeMux()
	servMux.Handle("/", http.FileServer(http.Dir(".")))

	var server http.Server

	server.Handler = servMux
	server.Addr = ":8080"

	server.ListenAndServe()

}
