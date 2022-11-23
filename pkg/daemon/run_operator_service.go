package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RunInstructionJSON struct {
	Blueprint uuid.UUID       `json:"blueprint"`
	Props     core.Properties `json:"props"`
	Gens      core.Generics   `json:"gens"`
	Stream    bool            `json:"stream"`
}

var RunOperatorService = &Service{map[string]*Endpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {

		if r.Method == "GET" {
			/*
				Get all running operators
			*/
			type responseListJSON struct {
				Objects []*runningOperator `json:"objects"`
				Status  string             `json:"status"`
				Error   *Error             `json:"error,omitempty"`
			}

			resp := responseListJSON{Objects: funk.Values(runningOperatorManager.ops).([]*runningOperator), Status: "success", Error: nil}
			writeJSON(w, &resp)

		} else if r.Method == "POST" {
			/*
				Start operator
			*/
			type responseSingleJSON struct {
				Object *runningOperator `json:"object"`
				Status string           `json:"status"`
				Error  *Error           `json:"error,omitempty"`
			}

			hub := GetHub(r)
			st := GetStorage(r)

			var requ RunInstructionJSON
			var resp responseSingleJSON

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&requ)
			if err != nil {
				resp = responseSingleJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0001"}}
				w.WriteHeader(400)
				writeJSON(w, &resp)
				return
			}

			op, err := api.BuildAndCompile(requ.Blueprint, requ.Gens, requ.Props, st)
			if err != nil {
				resp = responseSingleJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0002"}}
				w.WriteHeader(400)
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
						fmt.Println("Outgoing data", outgoing.Port, outgoing.Data)
						hub.broadCastTo(Root, Port, outgoing)
					case <-runOp.outStop:
						break loop
					}
				}
			}()

			resp.Status = "success"
			resp = responseSingleJSON{Object: runOp, Status: "success"}

			writeJSON(w, &resp)

		}
	}},

	"/{handle:\\w+}/": {func(w http.ResponseWriter, r *http.Request) {
		handle := mux.Vars(r)["handle"]

		runningOp, err := runningOperatorManager.Get(handle)
		if err != nil {
			w.WriteHeader(404)
			return
		}

		var idat interface{}
		if r.Method == "POST" {
			/*
				Pushing data into running operator in-port
			*/

			r.ParseForm() // TODO why is this required
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)

			// An empty buffer would result into an error that is why we check the length
			// and only than try to encode, because an empty POST is still valid and treated as trigger.
			if buf.Len() > 0 {
				// Unmarshal incoming data into a dataformat that is compatible with our operator
				// TODO find out if json.NewDecoder
				err := json.Unmarshal(buf.Bytes(), &idat)
				if err != nil {
					w.WriteHeader(400)
					return
				}
			}

			fmt.Println("Received data", idat)

			runningOp.incoming <- idat

			fmt.Println("Data pushed")

			w.WriteHeader(http.StatusNoContent)
		} else if r.Method == "DELETE" {
			/*
				Stop running operator
			*/
			runningOperatorManager.Halt(runningOp)

			w.WriteHeader(http.StatusNoContent)
		} else if r.Method == "OPTIONS" {

			w.WriteHeader(http.StatusNoContent)
		}
	}},
}}
