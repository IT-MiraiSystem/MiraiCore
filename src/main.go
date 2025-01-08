package main

import (
	"net/http"
)

type HelloHandler struct{}

func (h *HelloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, world!"))
}

func main() {
	handler := &HelloHandler{}
	server := &http.Server{
		Addr:    "0.0.0.0:80",
		Handler: handler,
	}
	server.ListenAndServe()
}
