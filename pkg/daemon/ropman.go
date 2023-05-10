package daemon

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"github.com/thoas/go-funk"
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

func (rop *runningOperator) Pull() (interface{}, bool) {
	for {
		select {
		case odat := <-rop.outgoing:
			return odat, true
		case <-time.After(500 * time.Millisecond):
			return nil, false
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

type PropertiesHash [16]byte

func hashProperties(p core.Properties) PropertiesHash {
	// hash value differs for any change in serializedProps, even when order of propNames changes --> sort props by names
	propNames := funk.Keys(p).([]string)
	sort.Strings(propNames)
	serializedProps := []byte{}

	for _, pn := range propNames {
		pv, ok := p[pn]

		if !ok {
			continue
		}

		jsonBytes, _ := json.Marshal(pv)
		serializedProps = append(serializedProps, jsonBytes...)
	}
	fmt.Println()
	return md5.Sum(serializedProps)
}

type runningOperatorManager struct {
	ropByHandle   map[string]*runningOperator
	handleByProps map[PropertiesHash]string
}

var rnd = rand.New(rand.NewSource(99))
var romanager = &runningOperatorManager{
	make(map[string]*runningOperator),
	make(map[PropertiesHash]string),
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
	}

	op.Main().Out().Bufferize()
	op.Start()

	return ro
}

func (rom *runningOperatorManager) addRopAccess(rop *runningOperator, props core.Properties) {
	propsHash := hashProperties(props)
	handle := rop.Handle

	rom.handleByProps[propsHash] = handle
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
}

func (rom *runningOperatorManager) Exec(bpid uuid.UUID, gens core.Generics, props core.Properties, st storage.Storage) (*runningOperator, error) {
	op, err := api.BuildAndCompile(bpid, gens, props, st)

	if err != nil {
		return nil, err
	}

	ro := rom.start(op)
	rom.addRopAccess(ro, props)
	rom.handleInputOutput(ro)

	return ro, nil
}

func (rom *runningOperatorManager) Halt(ro *runningOperator) error {
	go ro.op.Stop()
	ro.inStop <- true
	ro.outStop <- true
	delete(rom.ropByHandle, ro.Handle)
	return nil
}

func (rom runningOperatorManager) GetByHandle(handle string) (*runningOperator, error) {
	if ro, ok := rom.ropByHandle[handle]; ok {
		return ro, nil
	}
	return nil, fmt.Errorf("unknown handle value: %s", handle)
}

func (rom *runningOperatorManager) GetByProperties(props core.Properties) *runningOperator {
	propsHash := hashProperties(props)
	handle, ok := rom.handleByProps[propsHash]

	if !ok {
		return nil
	}

	rop := rom.ropByHandle[handle]

	return rop
}
