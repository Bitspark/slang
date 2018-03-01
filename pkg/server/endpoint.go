package server

import (
	"net/http"
	"encoding/json"
	"io"
	"log"
	"github.com/Bitspark/slang/pkg/builtin"
)

type SlangEndpoint struct {
	Handle func(w http.ResponseWriter, r *http.Request)
}

func readJSON(r io.Reader) *map[string]interface{} {
	dec := json.NewDecoder(r)
	dat := map[string]interface{}{}
	if err := dec.Decode(&dat); err != nil {
		log.Fatal(err)
	}
	return &dat
}

func writeJSON(w io.Writer, dat *map[string]interface{}) {
	json.NewEncoder(w).Encode(dat)
}

var ListBuiltinNames = &SlangEndpoint{func(w http.ResponseWriter, r *http.Request) {
	datOut := &map[string]interface{}{"objects": builtin.GetBuiltinNames()}
	writeJSON(w, datOut)
}}
