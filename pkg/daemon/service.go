package daemon

import (
	"net/http"
	"encoding/json"
	"io"
	"github.com/Bitspark/slang/pkg/api"
)

type DaemonService struct {
	Routes map[string]*DaemonEndpoint
}

type DaemonEndpoint struct {
	Handle func(e *api.Environ, w http.ResponseWriter, r *http.Request)
}

func writeJSON(w io.Writer, dat interface{}) error {
	return json.NewEncoder(w).Encode(dat)
}

type Error struct {
	Msg  string `json:"msg"`
	Code string `json:"code"`
}
