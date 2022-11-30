package daemon

import (
	"encoding/json"
	"fmt"
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
	inQueue  *dataQueue
	outQueue *dataQueue
	incoming chan interface{}
	outgoing chan interface{}
	inStop   chan bool
	outStop  chan bool
}

type dataQueue struct {
	consumed int
	queue    []interface{}
	enqueue  chan interface{}
	dequeue  chan interface{}
}

func newDataQueue() *dataQueue {
	dq := &dataQueue{
		0,
		make([]interface{}, 0),
		make(chan interface{}),
		make(chan interface{}),
	}

	// XXX need to stop this go routines
	go func() {
		for {
			enqueued := <-dq.enqueue
			dq.queue = append(dq.queue, enqueued)
		}
	}()

	go func() {
		for {
			if len(dq.queue) > dq.consumed {
				dq.dequeue <- dq.queue[dq.consumed]
				dq.consumed++
			}
		}
	}()

	return dq
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

type runningOperatorManager struct {
	ops map[string]*runningOperator
}

var rnd = rand.New(rand.NewSource(99))
var romanager = &runningOperatorManager{make(map[string]*runningOperator)}

func (rom *runningOperatorManager) newRunningOperator(op *core.Operator) *runningOperator {
	handle := strconv.FormatInt(rnd.Int63(), 16)
	url := "/run/" + handle + "/"
	ro := &runningOperator{
		op.Id(),
		handle,
		url,
		op,
		newDataQueue(),
		newDataQueue(),
		make(chan interface{}),
		make(chan interface{}),
		make(chan bool),
		make(chan bool),
	}

	rom.ops[handle] = ro

	op.Main().Out().Bufferize()
	op.Start()

	return ro
}

func (rom *runningOperatorManager) Run(op *core.Operator) *runningOperator {
	ro := rom.newRunningOperator(op)

	// Handle incoming data
	go func() {
	loop:
		for {
			select {
			case inData := <-ro.incoming:
				ro.inQueue.enqueue <- inData
			case inData := <-ro.inQueue.dequeue:
				op.Main().In().Push(inData)
			case <-ro.inStop:
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
				// XXX Could it happen, that this goroutine will never finish?
				i := p.Pull()
				fmt.Println("OUT", map[string]interface{}{p.String(): i})
				ro.outQueue.enqueue <- map[string]interface{}{p.String(): i}
				//po := portOutput{ro.Handle, p.String(), i, core.IsEOS(i), core.IsBOS(i), p}
				//ro.outgoing <- po
			}
		}()

	})

	go func() {
		var outgoings = make([]interface{}, 0)
		go func() {
		loop:
			for {
				select {
				case outData := <-ro.outQueue.dequeue:
					fmt.Println("OUTDATA")
					outgoings = append(outgoings, outData)
				case <-ro.outStop:
					break loop
				}
			}
		}()

		go func() {
			for {
				if len(outgoings) > 0 {
					fmt.Println("DEFAULT")
					ro.outgoing <- outgoings
					outgoings = make([]interface{}, 0)
				}
			}
		}()
	}()

	/*
		go func() {
			p := ro.op.Main().Out()

			go func() {
			loop:
				for {
					if p.Closed() {
						break loop
					}
					ro.outgoing <- p.Pull()
				}
			}()

			<-ro.outStop
		}()
	*/

	/*
		op.Main().Out().WalkPrimitivePorts(func(p *core.Port) {
			go func() {
				for {
					if p.Closed() {
						break
					}
					i := p.Pull()

					po := portOutput{ro.Handle, p.String(), i, core.IsEOS(i), core.IsBOS(i), p}
					ro.outgoing <- po
				}
			}()
		})
	*/

	return ro
}

func (rom *runningOperatorManager) Halt(ro *runningOperator) error {
	go ro.op.Stop()
	ro.inStop <- true
	ro.outStop <- true
	delete(rom.ops, ro.Handle)
	return nil
}

func (rom runningOperatorManager) Get(handle string) (*runningOperator, error) {
	if ro, ok := rom.ops[handle]; ok {
		return ro, nil
	}
	return nil, fmt.Errorf("unknown handle value: %s", handle)
}
