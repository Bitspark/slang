package main

import (
	"bufio"
	"encoding/json"
	"flag"
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
	"sync"
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

func newRunnerStorage(blueprints []core.OperatorDef) *storage.Storage {
	m := make(map[string]core.OperatorDef)

	for _, bp := range blueprints {
		m[bp.Id] = bp
	}

	return storage.NewStorage(nil).AddLoader(&runnerLoader{m})
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

var mgntAddr string

func main() {
	flag.StringVar(&mgntAddr, "mgnt-addr", "", "REQUIRED")
	flag.Parse()

	if mgntAddr == "" {
		log.Fatal("address for receiving management commands")
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}

}

type wrkCmds struct {
	op    *core.Operator
	sp    *SocketPort
	ready chan bool
}

func newWrkCmds() api.Cmds {
	return &wrkCmds{nil, nil, make(chan bool, 1)}
}

func (w *wrkCmds) Hello() (string, error) {
	if w.op != nil {
		return w.op.Id().String(), nil
	}
	return "", nil
}

func (w *wrkCmds) Init(a string) (string, error) {
	fmt.Println("--> /init", w.op)

	if w.op != nil {
		return "", nil
	}

	cmpltOpDef := completeOperatorDef{}
	if err := json.Unmarshal([]byte(a), &cmpltOpDef); err != nil {
		fmt.Println("--> 1)", err)
		return "", err
	}

	op, err := buildOperator(cmpltOpDef)

	fmt.Println("--> 2)", err)

	if err != nil {
		return "", err
	}

	w.op = op

	sp, err := newSocketPort(op, false, true)
	fmt.Println("--> 3)", err)
	if err != nil {
		return "", err
	}

	w.sp = sp

	fmt.Println("--> ready")
	w.ready <- true

	return w.PrtCfg()
}

func (w *wrkCmds) PrtCfg() (string, error) {

	fmt.Println("--> 0) /ports")

	if w.op == nil {
		fmt.Println("--> 1) no op")
		return "", fmt.Errorf("runner is not initialized: provide valid operator")
	}

	// todo timeout to prevent infinite loop
	for w.sp == nil {
		fmt.Println("--> 2) no op")
	}

	fmt.Println("---> portscfg", w.sp.String())

	return w.sp.String(), nil
}

func (w *wrkCmds) Action() error {
	<-w.ready

	fmt.Println("---> starting")

	op := w.op
	sp := w.sp

	fmt.Println("---> starting operator")

	op.Main().Out().Bufferize()
	op.Start()

	go sp.OnInput(hndlInput)
	go sp.OnOutput(hndlOutput)

	// Handle SIGTERM (CTRL-C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	op.Stopped()
	return nil
}

func run() error {
	wrkr := api.NewWorkerConnHandler(mgntAddr)
	err := wrkr.Begin(newWrkCmds)

	if err != nil {
		return err
	}

	return nil
}

func readCompleteOperatorDef(rd *bufio.Reader) (*completeOperatorDef, error) {
	b, err := ioutil.ReadAll(rd)
	fmt.Println("--->", err)

	if err != nil {
		return nil, err
	}

	d := completeOperatorDef{}
	if err := json.Unmarshal([]byte(b), &d); err != nil {
		return nil, err
	}

	return &d, nil
}

func buildOperator(d completeOperatorDef) (*core.Operator, error) {
	if !d.Valid() {
		if err := d.Validate(); err != nil {
			return nil, err
		}
	}

	stor := newRunnerStorage(d.Blueprints)

	bpId, _ := uuid.Parse(d.Main)
	return api.BuildAndCompile(bpId, d.Args.Generics, d.Args.Properties, *stor)
}

func wrBuf(buf *bufio.Writer, msg string) error {
	msg = strings.TrimSpace(msg)
	_, err := buf.WriteString(msg + "\n")
	return err
}

func rdBuf(buf *bufio.Reader) (string, error) {
	msg, err := buf.ReadString('\n')
	if err != nil {
		return msg, err
	}
	msg = strings.TrimSpace(msg)
	return msg, nil
}

func eof(e error) bool {
	return e == io.EOF
}

type SocketPort struct {
	op    *core.Operator
	pmap  map[net.Addr]*core.Port
	lnmap map[net.Addr]net.Listener
}

func newSocketPort(op *core.Operator, inested bool, onested bool) (*SocketPort, error) {
	pmap := make(map[net.Addr]*core.Port)
	lnmap := make(map[net.Addr]net.Listener)

	var port *core.Port

	port = op.Main().In()

	if inested {
		mapLnP, err := getListenerForPrimitivePorts(port)

		if err != nil {
			return nil, err
		}

		for ln, p := range mapLnP {
			lnmap[ln.Addr()] = ln
			pmap[ln.Addr()] = p
		}

	} else {
		ln, err := net.Listen("tcp", ":0")

		if err != nil {
			return nil, err
		}

		pmap[ln.Addr()] = port
		lnmap[ln.Addr()] = ln
	}

	port = op.Main().Out()

	if onested {
		mapLnP, err := getListenerForPrimitivePorts(port)

		if err != nil {
			return nil, err
		}

		for ln, p := range mapLnP {
			lnmap[ln.Addr()] = ln
			pmap[ln.Addr()] = p
		}

	} else {
		ln, err := net.Listen("tcp", ":0")

		if err != nil {
			return nil, err
		}

		pmap[ln.Addr()] = port
		lnmap[ln.Addr()] = ln
	}

	return &SocketPort{op, pmap, lnmap}, nil
}

func getListenerForPrimitivePorts(port *core.Port) (map[net.Listener]*core.Port, error) {
	if port.Primitive() {
		ln, err := net.Listen("tcp", ":0")

		if err != nil {
			return nil, err
		}

		return map[net.Listener]*core.Port{ln: port}, nil
	}

	if port.Stream() != nil {
		return getListenerForPrimitivePorts(port.Stream())
	}

	m := make(map[net.Listener]*core.Port)
	for _, pname := range port.MapEntries() {
		n, err := getListenerForPrimitivePorts(port.Map(pname))
		if err != nil {
			return nil, err
		}
		for k, v := range n {
			m[k] = v
		}
	}

	return m, nil
}

func (sp *SocketPort) OnInput(hndl func(op *core.Operator, p *core.Port, conn net.Conn, wg *sync.WaitGroup)) {
	for a, p := range sp.pmap {

		if p.Direction() != core.DIRECTION_IN {
			continue
		}
		ln := sp.lnmap[a]
		op := sp.op

		go func() {
			var wg sync.WaitGroup
			wg.Add(1)

			for !op.Stopped() {
				conn, err := ln.Accept()

				if err != nil {
					continue
				}

				go hndl(sp.op, p, conn, &wg)

				wg.Wait()
			}
		}()
	}
}

func (sp *SocketPort) OnOutput(hndl func(op *core.Operator, p *core.Port, conn net.Conn)) {
	for a, p := range sp.pmap {

		if p.Direction() != core.DIRECTION_OUT {
			continue
		}
		ln := sp.lnmap[a]
		op := sp.op

		go func() {
			for !op.Stopped() {
				conn, err := ln.Accept()

				if err != nil {
					continue
				}

				go hndl(sp.op, p, conn)
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

func hndlInput(op *core.Operator, p *core.Port, conn net.Conn, wg *sync.WaitGroup) {
	rdconn := bufio.NewReader(conn)

	for !op.Stopped() {
		msg, err := rdBuf(rdconn)

		if eof(err) {
			break
		}

		if len(msg) == 0 {
			p.Push(nil)
			continue
		}

		var idat interface{}
		err = json.Unmarshal([]byte(msg), &idat)

		if err != nil {
			continue
		}

		p.Push(idat)
	}

	wg.Done()

}

func hndlOutput(op *core.Operator, p *core.Port, conn net.Conn) {
	wrconn := bufio.NewWriter(conn)
	defer conn.Close()
	defer wrconn.Flush()

	for !op.Stopped() {
		odat := p.Pull()
		msg, err := json.Marshal(odat)

		if err != nil {
			continue
		}

		if err = wrBuf(wrconn, string(msg)); err != nil {
			break
		}
	}
}
