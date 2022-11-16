package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

type outJSON struct {
	Objects []*runningOperator `json:"objects"`
	Status  string             `json:"status"`
	Error   *Error             `json:"error,omitempty"`
}

type RunInstructionJSON struct {
	Blueprint uuid.UUID       `json:"blueprint"`
	Props     core.Properties `json:"props"`
	Gens      core.Generics   `json:"gens"`
	Stream    bool            `json:"stream"`
}

type RunInstructionResponseJSON struct {
	Handle string `json:"handle,omitempty"`
	URL    string `json:"url,omitempty"`
	Status string `json:"status"`
	Error  *Error `json:"error,omitempty"`
}

var RunOperatorService = &Service{map[string]*Endpoint{
	/*
	 *	Get all running operators
	 */
	"/": {func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			resp := outJSON{Objects: funk.Values(runningOperatorManager.ops).([]*runningOperator), Status: "success", Error: nil}
			writeJSON(w, &resp)

		} else if r.Method == "POST" {
			hub := GetHub(r)
			st := GetStorage(r)

			var requ RunInstructionJSON
			var resp RunInstructionResponseJSON

			decoder := json.NewDecoder(r.Body)
			fmt.Println(r.Body)
			err := decoder.Decode(&requ)
			if err != nil {
				resp = RunInstructionResponseJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0001"}}
				writeJSON(w, &resp)
				return
			}

			op, err := api.BuildAndCompile(requ.Blueprint, requ.Gens, requ.Props, st)
			if err != nil {
				resp = RunInstructionResponseJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0002"}}
				writeJSON(w, &resp)
				return
			}

			runOp := runningOperatorManager.Run(op)

			// Move into the background and wait on message from the operator resp. ports
			// and relay them through the `hub`
			go func() {
			loop:
				for {
					select {
					case outgoing := <-runOp.outgoing:
						// I don't know what happens when Root would be a dynamically changing variable.
						// Is root's value bound to the scope or is the reference bound to the scope.
						// I would suspect the latter, which means this is could turn into a race condition.
						hub.broadCastTo(Root, Port, outgoing)
					case <-runOp.outStop:
						break loop
					}
				}
			}()

			resp.Status = "success"
			resp.Handle = runOp.Handle
			resp.URL = runOp.URL

			writeJSON(w, &resp)

		}
	}},
}}
