package httpserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

type HttpServer struct {
	router *mux.Router
	Port   string
}

func NewHttpServer(port string) *HttpServer {
	return &HttpServer{Port: port}
}

func (h *HttpServer) Start() error {
	return http.ListenAndServe(h.Port, h.router)
}

func (h *HttpServer) AddRoute(f http.HandlerFunc, path string, method []string) {
	h.router.HandleFunc(path, f).Methods(method...)
}
