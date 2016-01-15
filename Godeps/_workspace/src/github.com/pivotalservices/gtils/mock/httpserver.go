package mock

import (
	"net/http"
	"net/http/httptest"
)

//HTTPServer - a fake http server object
type HTTPServer struct {
	Mux    *http.ServeMux
	Server *httptest.Server
}

//Setup -- allows us to setup our fake server
func (s *HTTPServer) Setup() {
	s.Mux = http.NewServeMux()
	s.Server = httptest.NewServer(s.Mux)
}

//Teardown -- allows for teardown of fake server
func (s *HTTPServer) Teardown() {
	s.Server.Close()
}
