package daemon

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/Bitspark/slang/pkg/storage"
)

type Service struct {
	Routes map[string]*Endpoint
}

type Endpoint struct {
	Handle func(w http.ResponseWriter, r *http.Request)
}

type contextKey string

const storageKey contextKey = "storage"
const hubKey contextKey = "hub"

func GetStorage(r *http.Request) storage.Storage {
	return *contextGet(r, storageKey).(*storage.Storage)
}

func SetHub(ctx context.Context, h *Hub) context.Context {
	return context.WithValue(ctx, hubKey, h)
}

func GetHub(r *http.Request) *Hub {
	return contextGet(r, hubKey).(*Hub)
}

func SetStorage(ctx context.Context, st *storage.Storage) context.Context {
	return context.WithValue(ctx, storageKey, st)
}

func contextGet(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
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

type ResponseJSON struct {
	Object  interface{}   `json:"object,omitempty"`
	Objects []interface{} `json:"objects,omitempty"`
	Status  string        `json:"status"`
	Error   *Error        `json:"error,omitempty"`
}

func response(w http.ResponseWriter, statusCode int, json interface{}) {
	w.WriteHeader(statusCode)
	if json != nil {
		writeJSON(w, json)
	}
}

func responseError(w http.ResponseWriter, statusCode int, err error, code string) {
	response(w, statusCode, &ResponseJSON{
		Status: "error",
		Error:  &Error{Msg: err.Error(), Code: code},
	})
}

func sendSuccess(w http.ResponseWriter, resp *responseOK) {
	w.WriteHeader(200)
	err := writeJSON(w, resp)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}
}

func sendFailure(w http.ResponseWriter, resp *responseBad) {
	w.WriteHeader(400)
	err := writeJSON(w, resp)
	if err != nil {
		log.Printf("[ERROR] %v\n", err)
	}
}
