package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type RequestRunOp struct {
	Blueprint uuid.UUID       `json:"blueprint"`
	Props     core.Properties `json:"props"`
	Gens      core.Generics   `json:"gens"`
}
type ResponseRunOp struct {
	Object *runningOperator `json:"object"`
	Status string           `json:"status"`
	Error  *Error           `json:"error,omitempty"`
}

func parseProperties(formData url.Values, propDef core.TypeDefMap) (core.Properties, error) {
	/*
		Convert operator properties passed as query parameters into correct type depending of expected slang type.
		Currently only supports primitive types: No Maps and Streams.
	*/

	p := make(core.Properties)

	for pname, ptype := range propDef {
		var pv interface{}
		var err error = nil

		fv := formData.Get(pname)

		switch ptype.Type {
		case "string", "primitive":
			pv = fv
		case "trigger":
			pv = nil
		case "boolean":
			pv, err = strconv.ParseBool(fv)
		case "number":
			pv, err = strconv.ParseFloat(fv, 32)
		case "map", "stream":
			err = fmt.Errorf("setting properties via query params is not supported for *maps* and *streams*")
		}

		if err != nil {
			return nil, err
		}

		p[pname] = pv

	}

	return p, nil
}

var RunnerService = &Service{map[string]*Endpoint{
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

			resp := responseListJSON{Objects: funk.Values(romanager.ropByHandle).([]*runningOperator), Status: "success", Error: nil}
			writeJSON(w, &resp)

		} else if r.Method == "POST" {
			/*
				Start operator
			*/
			//hub := GetHub(r)
			st := GetStorage(r)

			var requ RequestRunOp
			var resp ResponseRunOp

			decoder := json.NewDecoder(r.Body)
			err := decoder.Decode(&requ)
			if err != nil {
				resp = ResponseRunOp{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0001"}}
				w.WriteHeader(400)
				writeJSON(w, &resp)
				return
			}

			op, err := api.BuildAndCompile(requ.Blueprint, requ.Gens, requ.Props, st)
			if err != nil {
				resp = ResponseRunOp{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0002"}}
				w.WriteHeader(400)
				writeJSON(w, &resp)
				return
			}

			rop := romanager.Run(op)
			log.Printf("operator %s (id: %s) started", op.Name(), rop.Handle)

			/*
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
			*/

			resp.Status = "success"
			resp = ResponseRunOp{Object: rop, Status: "success"}

			writeJSON(w, &resp)

		}
	}},

	`/{blueprint:[0-9a-f]{8}-[0-9a-f-]+}/`: {func(w http.ResponseWriter, r *http.Request) {
		st := GetStorage(r)
		bpuuid, err := uuid.Parse(mux.Vars(r)["blueprint"])

		if err != nil {
			resp := ResponseRunOp{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0001"}}
			w.WriteHeader(400)
			writeJSON(w, &resp)
			return
		}

		blueprint, err := st.Load(bpuuid)

		if err != nil {
			resp := ResponseRunOp{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0002"}}
			w.WriteHeader(400)
			writeJSON(w, &resp)
			return
		}

		r.ParseForm()

		if r.Method == "GET" {
			props, err := parseProperties(r.Form, blueprint.PropertyDefs)

			if err != nil {
				resp := ResponseRunOp{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0002"}}
				w.WriteHeader(400)
				writeJSON(w, &resp)
				return
			}

			rop := romanager.GetByProperties(props)
			if rop == nil {
				st := GetStorage(r)
				rop, err = romanager.Exec(blueprint, nil, props, st)
				if err != nil {
					resp := ResponseRunOp{Status: "error", Error: &Error{Msg: err.Error(), Code: "E0003"}}
					w.WriteHeader(400)
					writeJSON(w, &resp)
					return
				}
			}

			rop.Push(nil)
			out := rop.Pull()

			if out != nil {
				fmt.Println("\t<--", out)
				w.WriteHeader(200)
				writeJSON(w, out)
			} else {
				w.WriteHeader(http.StatusNoContent)
			}
			return
		}

	}},

	`/{handle:\w+}/`: {func(w http.ResponseWriter, r *http.Request) {
		handle := mux.Vars(r)["handle"]

		rop, err := romanager.Get(handle)
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

			fmt.Println("\t-->", idat)
			rop.incoming <- idat

		loop:
			for {
				select {
				case odat := <-rop.outgoing:
					fmt.Println("\t<--", odat)
					w.WriteHeader(200)
					writeJSON(w, &odat)
					break loop
				case <-rop.outStop:
					fmt.Println("\toperator stopped")
					w.WriteHeader(http.StatusNoContent)
					break loop
				}
			}

		} else if r.Method == "DELETE" {
			/*
				Stop running operator
			*/
			romanager.Halt(rop)
			log.Printf("operator %s (id: %s) stopped", rop.op.Name(), rop.Handle)
			w.WriteHeader(http.StatusNoContent)
		} else if r.Method == "OPTIONS" {

			w.WriteHeader(http.StatusNoContent)
		}
	}},
}}
