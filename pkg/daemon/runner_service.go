package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

func (rom *_RunningOperatorManager) Halt(ro *runningOperator) error {
	go ro.op.Stop()
	ro.inStop <- true
	ro.outStop <- true
	delete(rom.ops, ro.Handle)

	return nil
}

func (rom _RunningOperatorManager) Get(handle string) (*runningOperator, error) {
	if runningOp, ok := rom.ops[handle]; ok {
		return runningOp, nil
	}
	return nil, fmt.Errorf("unknown handle value: %s", handle)
}
