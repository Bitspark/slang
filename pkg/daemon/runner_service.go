package daemon

import (
	"encoding/json"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Bitspark/slang/pkg/core"
)

var runningInstances = make(map[int64]struct {
	port int
	op   *core.Operator
})
var rnd = rand.New(rand.NewSource(99))

const SuffixPacked = "_packed"

var RunnerService = &Service{map[string]*Endpoint{
	"/": {func(st api.Storage, w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			type runInstructionJSON struct {
				Id     string          `json:"fqn"`
				Props  core.Properties `json:"props"`
				Gens   core.Generics   `json:"gens"`
				Stream bool            `json:"stream"`
			}

			type outJSON struct {
				URL    string `json:"url,omitempty"`
				Handle string `json:"handle,omitempty"`
				Status string `json:"status"`
				Error  *Error `json:"error,omitempty"`
			}

			var data outJSON

			decoder := json.NewDecoder(r.Body)
			var ri runInstructionJSON
			err := decoder.Decode(&ri)
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			port := 50000
			portUsed := true
			for portUsed {
				port++
				portUsed = false
				ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
				if err != nil {
					portUsed = true
				} else {
					ln.Close()
				}
			}

			opId, err := uuid.Parse(ri.Id)

			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			var httpDef *core.OperatorDef
			if ri.Stream {
				httpDef, err = constructHttpStreamEndpoint(st, port, opId, ri.Gens, ri.Props)
			} else {
				httpDef, err = constructHttpEndpoint(st, port, opId, ri.Gens, ri.Props)
			}
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			packagedOperator := strings.Replace(ri.Id+SuffixPacked, ".", string(filepath.Separator), -1) + ".yaml"

			/* TODO
			bytes, _ := yaml.Marshal(httpDef)
			ioutil.WriteFile(
				filepath.Join(st.WorkingDir(), packagedOperator),
				bytes,
				os.ModePerm,
			)
			*/

			op, err := api.Build(*httpDef, nil, nil)
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			handle := rnd.Int63()
			runningInstances[handle] = struct {
				port int
				op   *core.Operator
			}{port, op}

			op.Main().Out().Bufferize()
			op.Start()
			log.Printf("operator %s (port: %d, id: %s) started", packagedOperator, port, strconv.FormatInt(handle, 16))
			op.Main().In().Push(nil) // Start server

			data.Status = "success"
			data.Handle = strconv.FormatInt(handle, 16)
			data.URL = "/instance/" + strconv.FormatInt(handle, 16)

			writeJSON(w, &data)

			go func() {
				oprlt := op.Main().Out().Pull()
				log.Printf("operator %s (port: %d, id: %s) terminated: %v", packagedOperator, port, strconv.FormatInt(handle, 16), oprlt)
			}()
		} else if r.Method == "DELETE" {
			type stopInstructionJSON struct {
				Handle string `json:"handle"`
			}

			type outJSON struct {
				Status string `json:"status"`
				Error  *Error `json:"error,omitempty"`
			}

			var data outJSON

			decoder := json.NewDecoder(r.Body)
			var si stopInstructionJSON
			err := decoder.Decode(&si)
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			handle, _ := strconv.ParseInt(si.Handle, 16, 64)

			if ii, ok := runningInstances[handle]; !ok {
				data = outJSON{Status: "error", Error: &Error{Msg: "Unknown handle", Code: "E000X"}}
				writeJSON(w, &data)
				return
			} else {
				go ii.op.Stop()
				delete(runningInstances, handle)

				data.Status = "success"
				writeJSON(w, &data)
			}
		}
	}},
}}
