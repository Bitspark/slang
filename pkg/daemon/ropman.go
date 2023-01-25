package daemon

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

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

	op       *core.Operator
	incoming chan interface{}
	outgoing chan interface{}
	inStop   chan bool
	outStop  chan bool
}

func (rop *runningOperator) Push(data interface{}) {
	rop.incoming <- data
}

func (rop *runningOperator) Pull() interface{} {
	for {
		select {
		case odat := <-rop.outgoing:
			return odat
		case <-rop.outStop:
			return nil
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

type PropertiesHash [16]byte

func hashProperties(p core.Properties) PropertiesHash {
	arrBytes := []byte{}
	for _, item := range p {
		jsonBytes, _ := json.Marshal(item)
		arrBytes = append(arrBytes, jsonBytes...)
	}
	return md5.Sum(arrBytes)
}

type runningOperatorManager struct {
	ropByHandle map[string]*runningOperator
}

var rnd = rand.New(rand.NewSource(99))
var romanager = &runningOperatorManager{
	make(map[string]*runningOperator),
}

var handleByProps = make(map[PropertiesHash]string)

func (rom *runningOperatorManager) setRunningOperator(props core.Properties, rop *runningOperator) {
	propsHash := hashProperties(props)
	handle := rop.Handle

	handleByProps[propsHash] = handle
	rom.ropByHandle[handle] = rop
}

func (rom *runningOperatorManager) GetByProperties(props core.Properties) *runningOperator {
	propsHash := hashProperties(props)

	handle, ok := handleByProps[propsHash]

	if !ok {
		return nil
	}

	rop := rom.ropByHandle[handle]

	return rop
}

func (rom *runningOperatorManager) newRunningOperator(op *core.Operator) *runningOperator {
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
	}

	rom.ropByHandle[handle] = ro

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
			case incoming := <-ro.incoming:
				op.Main().In().Push(incoming)
			case <-ro.inStop:
				break loop
			}
		}
	}()

	// Handle outgoing data

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

func (rom *runningOperatorManager) Exec(blueprint uuid.UUID, gens core.Generics, props core.Properties, st storage.Storage) (*runningOperator, error) {
	op, err := api.BuildAndCompile(blueprint, gens, props, st)

	if err != nil {
		return nil, err
	}

	rop := romanager.Run(op)
	rom.setRunningOperator(props, rop)

	return rop, nil
}

func (rom *runningOperatorManager) Halt(ro *runningOperator) error {
	go ro.op.Stop()
	ro.inStop <- true
	ro.outStop <- true
	delete(rom.ropByHandle, ro.Handle)
	return nil
}

func (rom runningOperatorManager) Get(handle string) (*runningOperator, error) {
	if ro, ok := rom.ropByHandle[handle]; ok {
		return ro, nil
	}
	return nil, fmt.Errorf("unknown handle value: %s", handle)
}
