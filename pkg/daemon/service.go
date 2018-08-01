package daemon

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Bitspark/slang/pkg/api"
)

type Service struct {
	Routes map[string]*Endpoint
}

type Endpoint struct {
	Handle func(e *api.Environ, w http.ResponseWriter, r *http.Request)
}

func writeJSON(w io.Writer, dat interface{}) error {
	return json.NewEncoder(w).Encode(dat)
}

type Error struct {
	Msg  string `json:"msg"`
	Code string `json:"code"`
}

type responseOK struct {
	Data interface{} `json:"data,omitempty"`
}

type responseBad struct {
	Error *Error `json:"error,omitempty"`
}

func sendSuccess(w http.ResponseWriter, resp *responseOK) {
	w.WriteHeader(200)
	err := writeJSON(w, resp)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}
}

func sendFailure(w http.ResponseWriter, resp *responseBad) {
	w.WriteHeader(200) // XXX Use StatusCode 400
	err := writeJSON(w, resp)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}
}
