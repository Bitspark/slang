package daemon

import (
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/Bitspark/slang/pkg/api"
)

type DaemonServer struct {
	Env    *api.Environ
	Host   string
	Port   int
	router *mux.Router
}

func New(host string, port int) *DaemonServer {
	r := mux.NewRouter().Host("localhost").Subrouter()
	http.Handle("/", r)
	return &DaemonServer{api.NewEnviron(), host, port, r}
}

func (s *DaemonServer) AddService(pathPrefix string, services *DaemonService) {
	r := s.router.PathPrefix(pathPrefix).Subrouter()
	for path, endpoint := range services.Routes {
		(func(endpoint *DaemonEndpoint) {
			r.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { endpoint.Handle(s.Env, w, r) }))
		})(endpoint)
	}
}

func (s *DaemonServer) Run() error {
	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE"},
	}).Handler(s.router)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), handler)
}
