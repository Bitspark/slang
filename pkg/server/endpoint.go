package server

import (
	"net/http"
)

type SlangEndpoint interface {
	Handle(w http.ResponseWriter, r *http.Request)
}
