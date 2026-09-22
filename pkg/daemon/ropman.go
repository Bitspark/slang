package daemon

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"sync"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
)

type runningOperator struct {
	// JSON
	Blueprint uuid.UUID    `json:"blueprint"`
	In        core.TypeDef `json:"in"`
	Out       core.TypeDef `json:"out"`
	Handle    string       `json:"handle"`
	URL       string       `json:"url"`

	op        *core.Operator
	incoming  chan interface{}
	outgoing  chan interface{}
	inStop    chan bool
	outStop   chan bool
	stopOnce  sync.Once
	requestMu sync.Mutex
}

func (rop *runningOperator) Push(data interface{}) {
	select {
	case rop.incoming <- data:
	case <-rop.inStop:
	}
}

func (rop *runningOperator) Pull() (interface{}, bool) {
	for {
		select {
		case odat := <-rop.outgoing:
			return odat, true
		case <-rop.outStop:
			return nil, false
		}
	}
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

func definitionKey(bpid uuid.UUID, gens core.Generics, props core.Properties) string {
	// JSON sorts map keys and preserves property names as well as values.
	key, _ := json.Marshal(struct {
		Blueprint  uuid.UUID
		Generics   core.Generics
		Properties core.Properties
	}{bpid, gens, props})
	return string(key)
}

type runningOperatorManager struct {
	mu                 sync.Mutex
	ropByHandle        map[string]*runningOperator
	handleByDefinition map[string]string
}

var rnd = rand.New(rand.NewSource(99))
var romanager = &runningOperatorManager{
	ropByHandle:        make(map[string]*runningOperator),
	handleByDefinition: make(map[string]string),
}

func (rom *runningOperatorManager) start(op *core.Operator) *runningOperator {
	handle := strconv.FormatInt(rnd.Int63(), 16)
	url := "/run/" + handle + "/"
	ro := &runningOperator{
		op.Id(),
		op.Main().In().Define(),
		op.Main().Out().Define(),
		handle,
		url,
		op,
		make(chan interface{}),
		make(chan interface{}),
		make(chan bool),
		make(chan bool),
		sync.Once{},
		sync.Mutex{},
	}

	op.Main().Out().Bufferize()
	op.Start()

	return ro
}

func (rom *runningOperatorManager) addRopAccess(rop *runningOperator, gens core.Generics, props core.Properties) {
	key := definitionKey(rop.Blueprint, gens, props)
	handle := rop.Handle

	rom.handleByDefinition[key] = handle
	rom.ropByHandle[handle] = rop
}

func (rom *runningOperatorManager) handleInputOutput(ro *runningOperator) {
	op := ro.op

	// Handle incoming data
	go func() {
	loop:
		for {
			select {
			case incoming := <-ro.incoming:
				op.Main().In().Push(incoming)
			case <-ro.inStop:
				break loop
			}
		}
	}()

	// Block until a complete output is available; polling here used to spin and
	// could discard partially received map or stream values.
	go func() {
		for {
			item := op.Main().Out().Pull()
			select {
			case ro.outgoing <- item:
			case <-ro.outStop:
				return
			}
		}
	}()
}

func (rom *runningOperatorManager) Exec(bpid uuid.UUID, gens core.Generics, props core.Properties, st storage.Storage) (*runningOperator, error) {
	rom.mu.Lock()
	defer rom.mu.Unlock()
	return rom.exec(bpid, gens, props, st)
}

func (rom *runningOperatorManager) exec(bpid uuid.UUID, gens core.Generics, props core.Properties, st storage.Storage) (*runningOperator, error) {
	op, err := api.BuildAndCompile(bpid, gens, props, st)

	if err != nil {
		return nil, err
	}

	ro := rom.start(op)
	rom.addRopAccess(ro, gens, props)
	rom.handleInputOutput(ro)

	return ro, nil
}

func (rom *runningOperatorManager) Halt(ro *runningOperator) error {
	ro.stopOnce.Do(func() {
		close(ro.inStop)
		close(ro.outStop)
		ro.op.Stop()
		rom.mu.Lock()
		defer rom.mu.Unlock()
		delete(rom.ropByHandle, ro.Handle)
		for key, handle := range rom.handleByDefinition {
			if handle == ro.Handle {
				delete(rom.handleByDefinition, key)
			}
		}
	})
	return nil
}

func (rom *runningOperatorManager) GetByHandle(handle string) (*runningOperator, error) {
	rom.mu.Lock()
	defer rom.mu.Unlock()
	if ro, ok := rom.ropByHandle[handle]; ok {
		return ro, nil
	}
	return nil, fmt.Errorf("unknown handle value: %s", handle)
}

func (rom *runningOperatorManager) GetOrExec(bpid uuid.UUID, gens core.Generics, props core.Properties, st storage.Storage) (*runningOperator, error) {
	rom.mu.Lock()
	defer rom.mu.Unlock()
	key := definitionKey(bpid, gens, props)
	handle, ok := rom.handleByDefinition[key]
	if ok {
		if rop := rom.ropByHandle[handle]; rop != nil {
			return rop, nil
		}
	}
	return rom.exec(bpid, gens, props, st)
}

func (rom *runningOperatorManager) List() []*runningOperator {
	rom.mu.Lock()
	defer rom.mu.Unlock()
	result := make([]*runningOperator, 0, len(rom.ropByHandle))
	for _, rop := range rom.ropByHandle {
		result = append(result, rop)
	}
	return result
}
