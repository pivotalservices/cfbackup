package mock

import (
	"net/http"
	"net/http/httptest"
)

type HttpServer struct {
	Mux    *http.ServeMux
	Server *httptest.Server
}

func (s *HttpServer) Setup() {
	s.Mux = http.NewServeMux()
	s.Server = httptest.NewServer(s.Mux)
}

func (s *HttpServer) Teardown() {
	s.Server.Close()
}
