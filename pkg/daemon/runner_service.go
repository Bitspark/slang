package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

type RunInstructionJSON struct {
	Id     string          `json:"id"`
	Props  core.Properties `json:"props"`
	Gens   core.Generics   `json:"gens"`
	Stream bool            `json:"stream"`
}
type InstanceStateJSON struct {
	URL    string `json:"url,omitempty"`
	Handle string `json:"handle,omitempty"`
	Status string `json:"status"`
	Error  *Error `json:"error,omitempty"`
}

type runningInstance struct {
	Port int            `json:"port"`
	Op   *core.Operator `json:"omit"`
}

var runningInstances = make(map[string]runningInstance)
var rnd = rand.New(rand.NewSource(99))

type httpDefLoader struct {
	httpDef *core.OperatorDef
}

func (l *httpDefLoader) List() ([]uuid.UUID, error) {
	httpDefId, _ := uuid.Parse(l.httpDef.Id)
	return []uuid.UUID{httpDefId}, nil
}

func (l *httpDefLoader) Has(opId uuid.UUID) bool {
	httpDefId, _ := uuid.Parse(l.httpDef.Id)
	return httpDefId == opId
}

func (l *httpDefLoader) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	if !l.Has(opId) {
		return nil, fmt.Errorf("")
	}
	return l.httpDef, nil
}

var InstanceService = &Service{map[string]*Endpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			data := runningInstances
			writeJSON(w, &data)
		}
	}},
}}

var RunnerService = &Service{map[string]*Endpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {
		hub := GetHub(r)
		st := GetStorage(r)
		if r.Method == "POST" {
			var data InstanceStateJSON

			decoder := json.NewDecoder(r.Body)
			var ri RunInstructionJSON
			err := decoder.Decode(&ri)
			if err != nil {
				data = InstanceStateJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
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
				data = InstanceStateJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
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
				data = InstanceStateJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			st.AddBackend(&httpDefLoader{httpDef})
			httpDefId, _ := uuid.Parse(httpDef.Id)
			op, err := api.BuildAndCompile(httpDefId, nil, nil, st)
			if err != nil {
				data = InstanceStateJSON{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			handle := strconv.FormatInt(rnd.Int63(), 16)
			runningInstances[handle] = runningInstance{port, ri.Id, op}

			op.Main().Out().Bufferize()
			op.Start()
			log.Printf("operator %s (port: %d, id: %s) started", op.Name(), port, handle)
			op.Main().In().Push(nil) // Start server
			hub.broadCastTo(Root, "Starting Operator")

			data.Status = "success"
			data.Handle = handle
			data.URL = "/instance/" + handle

			writeJSON(w, &data)

			go func() {
				oprlt := op.Main().Out().Pull()
				hub.broadCastTo(Root, "Stopping Operator")
				log.Printf("operator %s (port: %d, id: %s) terminated: %v", op.Name(), port, handle, oprlt)
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

			if ii, ok := runningInstances[si.Handle]; !ok {
				data = outJSON{Status: "error", Error: &Error{Msg: "Unknown handle", Code: "E000X"}}
				writeJSON(w, &data)
				return
			} else {
				go ii.Op.Stop()
				delete(runningInstances, si.Handle)

				data.Status = "success"
				writeJSON(w, &data)
			}
		}
	}},
}}
