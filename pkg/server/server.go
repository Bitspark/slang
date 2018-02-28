package server

import (
	"net/http"
	"fmt"
)

type SlangServer struct {
	Port int
}

func New(port int) *SlangServer {
	return &SlangServer{port}
}

func (s *SlangServer) AddEndpoint(path string, e SlangEndpoint) {
	http.HandleFunc(path, e.Handle)
}

func (s *SlangServer) Run() error {
	return http.ListenAndServe(fmt.Sprintf(":%v", s.Port), nil)
}
