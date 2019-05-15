package daemon

import (
	"bytes"
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
	"github.com/gorilla/mux"
)

type RunInstruction struct {
	Id     uuid.UUID       `json:"id"`
	Props  core.Properties `json:"props"`
	Gens   core.Generics   `json:"gens"`
	Stream bool            `json:"stream"`
}
type InstanceState struct {
	URL    string `json:"url,omitempty"`
	Handle string `json:"handle,omitempty"`
	Status string `json:"status"`
	Error  *Error `json:"error,omitempty"`
}

type runningInstance struct {
	// todo remove .Port
	Port     int              `json:"port"`
	OpId     uuid.UUID        `json:"id"`
	Op       *core.Operator   `json:"Op"`
	Incoming chan interface{} `json:"-"`
	Outgoing chan portMessage `json:"-"`
}

type portMessage struct {
	Port *core.Port
	Msg  interface{}
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

var RunningInstanceService = &Service{map[string]*Endpoint{
	"/{handle:\\w+}/": {func(w http.ResponseWriter, r *http.Request) {
		handle := mux.Vars(r)["handle"]

		runningIns, ok := runningInstances[handle]
		if !ok {
			w.WriteHeader(404)
			return
		}

		if r.Method == "POST" {
			r.ParseForm()
			buf := new(bytes.Buffer)
			buf.ReadFrom(r.Body)

			var idat interface{}
			err := json.Unmarshal(buf.Bytes(), &idat)

			if err != nil {
				w.WriteHeader(400)
				return
			}

			runningIns.Incoming <- idat

			writeJSON(w, &runningIns)
		}

	}},
}}

var RunnerService = &Service{map[string]*Endpoint{
	"/": {func(w http.ResponseWriter, r *http.Request) {
		hub := GetHub(r)
		st := GetStorage(r)
		if r.Method == "POST" {
			var data InstanceState

			decoder := json.NewDecoder(r.Body)
			var ri RunInstruction
			err := decoder.Decode(&ri)
			if err != nil {
				data = InstanceState{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
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

			opId := ri.Id

			if err != nil {
				data = InstanceState{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
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
				data = InstanceState{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			st.AddBackend(&httpDefLoader{httpDef})
			httpDefId, _ := uuid.Parse(httpDef.Id)
			op, err := api.BuildAndCompile(httpDefId, nil, nil, st)
			if err != nil {
				data = InstanceState{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			handle := strconv.FormatInt(rnd.Int63(), 16)
			runningInstances[handle] = runningInstance{port, ri.Id, op, make(chan interface{}, 0), make(chan portMessage, 0)}

			op.Main().Out().Bufferize()
			op.Start()
			log.Printf("operator %s (port: %d, id: %s) started", op.Name(), port, handle)
			op.Main().In().Push(nil) // Start server
			hub.broadCastTo(Root, "Starting Operator")

			data.Status = "success"
			data.Handle = handle
			data.URL = "/instance/" + handle

			writeJSON(w, &data)

			op.Main().Out().WalkPrimitivePorts(func(p *core.Port) {
				i := p.Pull()

				if core.IsMarker(i) {
					return
				}

				hub.broadCastTo(Root, fmt.Sprintf("%s:%v", p.Name(), i))
			})

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
	"/ws/": {func(w http.ResponseWriter, r *http.Request) {
		hub := GetHub(r)
		st := GetStorage(r)
		if r.Method == "POST" {
			var data InstanceState

			decoder := json.NewDecoder(r.Body)
			var ri RunInstruction
			err := decoder.Decode(&ri)
			if err != nil {
				data = InstanceState{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			opId := ri.Id
			op, err := api.BuildAndCompile(opId, ri.Gens, ri.Props, st)
			if err != nil {
				data = InstanceState{Status: "error", Error: &Error{Msg: err.Error(), Code: "E000X"}}
				writeJSON(w, &data)
				return
			}

			// --- Running operator Instance
			handle := strconv.FormatInt(rnd.Int63(), 16)
			runningIns := runningInstance{0, ri.Id, op, make(chan interface{}, 0), make(chan portMessage, 0)}
			runningInstances[handle] = runningIns

			op.Main().Out().Bufferize()
			op.Start()
			go func() {
				log.Printf("operator %s (id: %s) started", op.Name(), handle)

				for {
					select {
					case incoming := <-runningIns.Incoming:
						op.Main().In().Push(incoming)
					}
				}
			}()

			go func() {
				op.Main().Out().WalkPrimitivePorts(func(p *core.Port) {
					i := p.Pull()

					fmt.Println("// %v", i)
					fmt.Println("// outgoing %v", i)

					if !core.IsMarker(i) {
						runningIns.Outgoing <- portMessage{p, i}
					}
				})
			}()
			// --- Running operator Instance [END]

			// --- Handle Incoming and Outgoing data
			go func() {
				for {
					select {
					case outgoing := <-runningIns.Outgoing:
						hub.broadCastTo(Root, fmt.Sprintf("%v", outgoing))
					}
				}
			}()
			// --- Handle Incoming and Outgoing data [END]

			data.Status = "success"
			data.Handle = handle
			data.URL = "/instance/" + handle + "/"

			writeJSON(w, &data)

		}
	}},
}}
