package daemon

import (
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type DaemonServer struct {
	Host   string
	Port   int
	router *mux.Router
}

func New(host string, port int) *DaemonServer {
	r := mux.NewRouter().Host("localhost").Subrouter()
	http.Handle("/", r)
	return &DaemonServer{host, port, r}

}

func (s *DaemonServer) AddService(pathPrefix string, services *DaemonService) {
	r := s.router.PathPrefix(pathPrefix).Subrouter()
	for path, endpoint := range services.Routes {
		r.HandleFunc(path, http.HandlerFunc(endpoint.Handle))
	}
}

func (s *DaemonServer) Run() error {
	handler := cors.Default().Handler(s.router)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), handler)
}
