package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/Bitspark/go-funk"
	"github.com/Bitspark/slang/pkg/api"
	"github.com/Bitspark/slang/pkg/core"
	"github.com/Bitspark/slang/pkg/storage"
	"github.com/google/uuid"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

/*** (SlangFile ******/
type completeOperatorDef struct {
	Main string `json:"main" yaml:"main"`

	Args struct {
		Properties core.Properties `json:"properties,omitempty" yaml:"properties,omitempty"`
		Generics   core.Generics   `json:"generics,omitempty" yaml:"generics,omitempty"`
	} `json:"args,omitempty" yaml:"args,omitempty"`

	Blueprints []core.OperatorDef `json:"blueprints" yaml:"blueprints"`

	valid bool
}

func (sf completeOperatorDef) Valid() bool {
	return sf.valid
}

func (sf *completeOperatorDef) Validate() error {
	if sf.Main == "" {
		return fmt.Errorf(`missing main blueprint id`)
	}

	if _, err := uuid.Parse(sf.Main); err != nil {
		return fmt.Errorf(`blueprint id is not a valid UUID v4: "%s" --> "%s"`, sf.Main, err)
	}

	if len(sf.Blueprints) == 0 {
		return fmt.Errorf(`incomplete slang file: no blueprint definitions found`)

	}

	for _, bp := range sf.Blueprints {
		if err := bp.Validate(); err != nil {
			return err
		}
	}

	sf.valid = true
	return nil
}

/*** (Loader *******/
type runnerLoader struct {
	blueprintbyId map[string]core.OperatorDef
}

func (l *runnerLoader) Has(opId uuid.UUID) bool {
	_, ok := l.blueprintbyId[opId.String()]
	return ok
}

func (l *runnerLoader) List() ([]uuid.UUID, error) {
	var uuidList []uuid.UUID

	for _, idOrName := range funk.Keys(l.blueprintbyId).([]string) {
		if id, err := uuid.Parse(idOrName); err == nil {
			uuidList = append(uuidList, id)
		}
	}

	return uuidList, nil
}

func (l *runnerLoader) Load(opId uuid.UUID) (*core.OperatorDef, error) {
	if opDef, ok := l.blueprintbyId[opId.String()]; ok {
		return &opDef, nil
	}
	return nil, fmt.Errorf("unknown operator")
}

func main() {
	cmpltOpDef, err := readCompleteOperatorDef(bufio.NewReader(os.Stdin))

	if err != nil {
		log.Fatal(err)
	}

	operator, err := buildOperator(*cmpltOpDef)

	if err != nil {
		log.Fatal(err)
	}

	if err := start(*operator); err != nil {
		log.Fatal(err)
	}
}

func readCompleteOperatorDef(rd *bufio.Reader) (*completeOperatorDef, error) {
	b, err := ioutil.ReadAll(rd)

	if err != nil {
		return nil, err
	}

	d := completeOperatorDef{}
	if err := json.Unmarshal([]byte(b), &d); err != nil {
		return nil, err
	}

	return &d, nil
}

func buildOperator(cmpltOpDef completeOperatorDef) (*core.Operator, error) {
	stor := newRunnerStorage(cmpltOpDef.Blueprints)

	return newOperator(cmpltOpDef, *stor)
}

func newRunnerStorage(blueprints []core.OperatorDef) *storage.Storage {
	m := make(map[string]core.OperatorDef)

	for _, bp := range blueprints {
		m[bp.Id] = bp
	}

	return storage.NewStorage(nil).AddLoader(&runnerLoader{m})
}

func newOperator(d completeOperatorDef, stor storage.Storage) (*core.Operator, error) {
	if !d.Valid() {
		if err := d.Validate(); err != nil {
			return nil, err
		}
	}

	bpId, _ := uuid.Parse(d.Main)
	return api.BuildAndCompile(bpId, d.Args.Generics, d.Args.Properties, stor)
}

func wrErr(err error) {
	if err == nil {
		return
	}
	stderr := bufio.NewWriter(os.Stderr)
	_, _ = stderr.WriteString(err.Error() + "\n")
	_ = stderr.Flush()
}

type SocketPort struct {
	op    core.Operator
	pmap  map[net.Addr]*core.Port
	lnmap map[net.Addr]net.Listener
}

func newSocketPort(op core.Operator) (*SocketPort, error) {
	pmap := make(map[net.Addr]*core.Port)
	lnmap := make(map[net.Addr]net.Listener)

	ln, err := net.Listen("tcp", ":0")

	if err != nil {
		return nil, err
	}

	pmap[ln.Addr()] = op.Main().In()
	lnmap[ln.Addr()] = ln

	mapLnP, err := getListenerForPort(op.Main().Out())

	if err != nil {
		return nil, err
	}

	for ln, p := range mapLnP {
		lnmap[ln.Addr()] = ln
		pmap[ln.Addr()] = p
	}

	return &SocketPort{op, pmap, lnmap}, nil
}

func getListenerForPort(port *core.Port) (map[net.Listener]*core.Port, error) {
	if port.Primitive() {
		ln, err := net.Listen("tcp", ":0")

		if err != nil {
			return nil, err
		}

		return map[net.Listener]*core.Port{ln: port}, nil
	}

	if port.Stream() != nil {
		return getListenerForPort(port.Stream())
	}

	m := make(map[net.Listener]*core.Port)
	for _, pname := range port.MapEntries() {
		n, err := getListenerForPort(port.Map(pname))
		if err != nil {
			return nil, err
		}
		for k, v := range n {
			m[k] = v
		}
	}

	return m, nil
}

func (sp *SocketPort) OnAccept(hndlIn func(op core.Operator, p *core.Port, conn net.Conn), hndlOut func(op core.Operator, p *core.Port, conn net.Conn)) {
	for a, ln := range sp.lnmap {
		go func() {
			for !sp.op.Stopped() {
				conn, err := ln.Accept()
				wrErr(err)
				p, _ := sp.pmap[a]

				if p.Direction() == core.DIRECTION_IN {
					hndlIn(sp.op, p, conn)
				} else {
					hndlOut(sp.op, p, conn)
				}
			}
		}()
	}
}

func (sp *SocketPort) String() string {
	paddr := make(map[string]string)
	for a, p := range sp.pmap {
		paddr[p.StringifyComplete()] = a.String()
	}

	j, _ := json.Marshal(paddr)
	return string(j)
}

func start(op core.Operator) error {
	op.Main().Out().Bufferize()
	op.Start()

	sp, err := newSocketPort(op)
	if err != nil {
		return err
	}

	go sp.OnAccept(handleInput, handleOutput)

	fmt.Println(sp.String())

	// Handle SIGTERM (CTRL-C)
	info := make(chan os.Signal, 1)
	signal.Notify(info, syscall.SIGUSR1)
	go func() {
		select {
		case <-info:
			fmt.Println(sp.String())
		default:
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	op.Stopped()
	return nil
}

func handleInput(op core.Operator, p *core.Port, conn net.Conn) {
	rdconn := bufio.NewReader(conn)
	defer conn.Close()

	for !op.Stopped() {
		msg, err := rdconn.ReadString('\n')

		if err == io.EOF {
			break
		}

		msg = strings.TrimSpace(msg)

		if len(msg) == 0 {
			p.Push(nil)
			continue
		}

		var idat interface{}
		err = json.Unmarshal([]byte(msg), &idat)

		if err != nil {
			wrErr(err)
			continue
		}

		p.Push(idat)
	}

}

func handleOutput(op core.Operator, p *core.Port, conn net.Conn) {
	wrconn := bufio.NewWriter(conn)
	defer conn.Close()
	defer wrconn.Flush()

	for !op.Stopped() {
		odat := p.Pull()
		msg, err := json.Marshal(odat)

		if err != nil {
			wrErr(err)
			continue
		}

		wrconn.WriteString(string(msg) + "\n")
	}
}
