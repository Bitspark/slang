package daemon

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type Server struct {
	Env    *api.Environ
	Host   string
	Port   int
	router *mux.Router
}

func New(host string, port int) *Server {
	r := mux.NewRouter().Host("localhost").Subrouter()
	http.Handle("/", r)
	return &Server{api.NewEnviron(), host, port, r}
}

func (s *Server) AddService(pathPrefix string, services *Service) {
	s.AddRedirect(pathPrefix, pathPrefix+"/")
	r := s.router.PathPrefix(pathPrefix).Subrouter()
	for path, endpoint := range services.Routes {
		(func(endpoint *Endpoint) {
			r.HandleFunc(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { endpoint.Handle(s.Env, w, r) }))
		})(endpoint)
	}
}

func (s *Server) AddAppServer(pathPrefix string, directory http.Dir) {
	s.AddRedirect(pathPrefix, pathPrefix+"/")
	r := s.router.PathPrefix(pathPrefix)
	r.Handler(http.StripPrefix(pathPrefix,
		r.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				if m, _ := regexp.Match(`\..{1,4}$`, []byte(r.URL.Path)); m {
					http.ServeFile(w, r, filepath.Join(string(directory), r.URL.Path))
					return
				}
			}
			http.ServeFile(w, r, filepath.Join(string(directory), "index.html"))
		}).GetHandler()))
}

func (s *Server) AddStaticServer(pathPrefix string, directory http.Dir) {
	s.AddRedirect(pathPrefix, pathPrefix+"/")
	r := s.router.PathPrefix(pathPrefix)
	r.Handler(http.StripPrefix(pathPrefix, http.FileServer(directory)))
}

func (s *Server) AddRedirect(path string, redirectTo string) {
	r := s.router.Path(path)
	r.Handler(http.RedirectHandler(redirectTo, http.StatusSeeOther))
}

func (s *Server) Run() error {
	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE"},
	}).Handler(s.router)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.Port), handler)
}
