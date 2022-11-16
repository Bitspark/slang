package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/Bitspark/slang/pkg/core"
	"github.com/google/uuid"
)

type runningOperator struct {
	// JSON
	Blueprint uuid.UUID `json:"blueprint"`
	Handle    string    `json:"handle"`
	URL       string    `json:"url"`

	op       *core.Operator
	incoming chan interface{}
	outgoing chan portOutput
	inStop   chan bool
	outStop  chan bool
}

type portOutput struct {
	// JSON
	Handle string      `json:"handle"`
	Port   string      `json:"port"`
	Data   interface{} `json:"data"`
	IsEOS  bool        `json:"isEOS"`
	IsBOS  bool        `json:"isBOS"`

	port *core.Port
}

func (pm *portOutput) String() string {
	j, _ := json.Marshal(pm)
	return string(j)
}

type _RunningOperatorManager struct {
	ops map[string]*runningOperator
}

var rnd = rand.New(rand.NewSource(99))
var runningOperatorManager = &_RunningOperatorManager{make(map[string]*runningOperator)}

func (rom *_RunningOperatorManager) newRunningOperator(op *core.Operator) *runningOperator {
	handle := strconv.FormatInt(rnd.Int63(), 16)
	url := "/run/" + handle + "/"
	runningOp := &runningOperator{op.Id(), handle, url, op, make(chan interface{}, 0), make(chan portOutput, 0), make(chan bool, 0), make(chan bool, 0)}
	rom.ops[handle] = runningOp
	op.Main().Out().Bufferize()
	op.Start()
	log.Printf("operator %s (id: %s) started", op.Name(), handle)
	return runningOp
}

func (rom *_RunningOperatorManager) Run(op *core.Operator) *runningOperator {
	runningOp := rom.newRunningOperator(op)

	// Handle incoming data
	go func() {
	loop:
		for {
			select {
			case incoming := <-runningOp.incoming:
				fmt.Println("Push data", incoming)
				op.Main().In().Push(incoming)
			case <-runningOp.inStop:
				break loop
			}
		}
	}()

	// Handle outgoing data
	op.Main().Out().WalkPrimitivePorts(func(p *core.Port) {
		go func() {
			for {
				if p.Closed() {
					break
				}
				i := p.Pull()

				po := portOutput{runningOp.Handle, p.String(), i, core.IsEOS(i), core.IsBOS(i), p}
				runningOp.outgoing <- po
			}
		}()
	})
	return runningOp
}

func (rom *_RunningOperatorManager) Halt(handle string) error {
	// `Halt` to me suggest that there is a way to resume operations
	// which is not the case.
	ro, err := runningOperatorManager.Get(handle)

	if err != nil {
		return err
	}

	go ro.op.Stop()
	ro.inStop <- true
	ro.outStop <- true
	delete(rom.ops, handle)

	return nil
}

func (rom _RunningOperatorManager) Get(handle string) (*runningOperator, error) {
	if runningOp, ok := rom.ops[handle]; ok {
		return runningOp, nil
	}
	return nil, fmt.Errorf("unknown handle value: %s", handle)
}

var RunnerService = &Service{map[string]*Endpoint{
	/*
	 * Start and Stop operator
	 */

	"/": {Handle: func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
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

			if err := runningOperatorManager.Halt(si.Handle); err == nil {
				data.Status = "success"
			} else {
				data = outJSON{Status: "error", Error: &Error{Msg: "Unknown handle", Code: "E000X"}}
			}

			writeJSON(w, &data)
		}
	}},
}}
