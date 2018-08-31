package daemon

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"gopkg.in/yaml.v2"
)

var runningInstances = make(map[int64]struct {
	port int
	op   *core.Operator
})
var rnd = rand.New(rand.NewSource(99))

const SuffixPacked = "_packed"

var RunnerService = &Service{map[string]*Endpoint{
	"/": {func(e *api.Environ, w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			type runInstructionJSON struct {
				Fqn    string          `json:"fqn"`
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
				for _, ri := range runningInstances {
					if ri.port == port {
						portUsed = true
						break
					}
				}
			}

			var httpDef *core.OperatorDef
			if ri.Stream {
				httpDef, err = constructHttpStreamEndpoint(e, port, ri.Fqn, ri.Gens, ri.Props)
			} else {
				httpDef, err = constructHttpEndpoint(e, port, ri.Fqn, ri.Gens, ri.Props)
			}
			if err != nil {
				data = outJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			packagedOperator := strings.Replace(ri.Fqn+SuffixPacked, ".", string(filepath.Separator), -1) + ".yaml"

			bytes, _ := yaml.Marshal(httpDef)
			ioutil.WriteFile(
				filepath.Join(e.WorkingDir(), packagedOperator),
				bytes,
				os.ModePerm,
			)

			op, err := e.BuildAndCompileOperator(packagedOperator, nil, nil)
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

			op.Start()
			op.Main().In().Push(nil) // Start server

			data.Status = "success"
			data.Handle = strconv.FormatInt(handle, 16)
			data.URL = "http://localhost:" + strconv.Itoa(port)

			port++

			writeJSON(w, &data)
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
