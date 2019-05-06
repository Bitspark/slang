package daemon

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Bitspark/slang/pkg/env"

	"github.com/rs/cors"

	"github.com/gorilla/mux"
)

var SlangVersion string

type Server struct {
	Host   string
	Port   int
	router *mux.Router
}

func addContext(ctx context.Context, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewServer(env *env.Environment) *Server {
	r := mux.NewRouter().StrictSlash(true)
	http.Handle("/", r)
	srv := &Server{env.HTTP.Address, env.HTTP.Port, r}
	srv.mountWebServices()
	return srv
}

func (s *Server) mountWebServices() {
	s.AddService("/operator", DefinitionService)
	s.AddService("/run", RunnerService)
	s.AddService("/share", SharingService)
	s.AddOperatorProxy("/instance")
}

func (s *Server) AddService(pathPrefix string, services *Service) {
	r := s.router.PathPrefix(pathPrefix).Subrouter()
	for path, endpoint := range services.Routes {
		(func(endpoint *Endpoint) {
			r.HandleFunc(path, endpoint.Handle)
		})(endpoint)
	}
}

func (s *Server) AddStaticServer(pathPrefix string, directory http.Dir) {
	r := s.router.PathPrefix(pathPrefix)
	r.Handler(http.StripPrefix(pathPrefix, http.FileServer(http.Dir(directory))))
}

func (s *Server) AddOperatorProxy(pathPrefix string) {
	r := s.router.PathPrefix(pathPrefix)
	r.Handler(http.StripPrefix(pathPrefix,
		r.HandlerFunc(proxyRequestToOperator).GetHandler()))
}

func (s *Server) AddRedirect(path string, redirectTo string) {
	r := s.router.Path(path)
	r.Handler(http.RedirectHandler(redirectTo, http.StatusSeeOther))
}

func (s *Server) Run(ctx context.Context) error {
	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE"},
	}).Handler(s.router)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), addContext(ctx, handler))
}
