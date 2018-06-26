package daemon

import (
	"net/http"
	"encoding/json"
	"io"
)

type DaemonService struct {
	Routes map[string]*DaemonEndpoint
}

type DaemonEndpoint struct {
	Handle func(w http.ResponseWriter, r *http.Request)
}

func writeJSON(w io.Writer, dat interface{}) error {
	return json.NewEncoder(w).Encode(dat)
}

type Error struct {
	Msg  string `json:"msg"`
	Code string `json:"code"`
}
