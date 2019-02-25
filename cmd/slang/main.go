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

var verbose bool

func main() {
	flag.Parse()

	slangFile := flag.Arg(0)

	operator, err := buildOperator(slangFile)

	if err != nil {
		log.Fatal(err)
	}

	start(*operator)
}

func buildOperator(slangFile string) (*core.Operator, error) {
	cmpltOpDef, err := readSlangFile(slangFile)

	if err != nil {
		return nil, err
	}

	stor := newRunnerStorage(cmpltOpDef.Blueprints)

	return newOperator(*cmpltOpDef, *stor)
}

func readSlangFile(slangFile string) (*completeOperatorDef, error) {
	b, err := ioutil.ReadFile(slangFile)
	if err != nil {
		return nil, fmt.Errorf("could not read operator file: %s", slangFile)
	}

	d := completeOperatorDef{}
	if err := json.Unmarshal([]byte(b), &d); err != nil {
		return nil, err
	}

	return &d, nil
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

func start(op core.Operator) {
	stdin := bufio.NewReader(os.Stdin)
	stdout := bufio.NewWriter(os.Stdout)
	stderr := bufio.NewWriter(os.Stderr)

	op.Main().Out().Bufferize()
	op.Start()

	inPort := op.Main().In()
	outPort := op.Main().Out()

	// Handle STDIN
	go func() {
		defer op.Stop()

		for {
			idat, eof := readPipeDecodeJSON(stdin, stderr)

			if eof {
				break
			}

			inPort.Push(idat)
		}

	}()

	// Handle STDOUT
	go func() {
		defer os.Exit(0)

		for !op.Stopped() {
			odat := outPort.Pull()
			text, err := json.Marshal(odat)

			if err != nil {
				stderr.WriteString(err.Error() + "\n")
				stderr.Flush()
				continue
			}

			stdout.WriteString(string(text))
		}

	}()

	// Handle SIGTERM (CTRL-C)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		op.Stop()
		os.Exit(1)
	}()

	select {}
}

func readPipeDecodeJSON(rd *bufio.Reader, wrerr *bufio.Writer) (interface{}, bool) {
	var idat interface{}
	for {
		text, err := rd.ReadString('\n')

		if err == io.EOF {
			return nil, true
		}

		text = strings.TrimSpace(text)

		if len(text) == 0 {
			return nil, false
		}

		err = json.Unmarshal([]byte(text), &idat)

		if err != nil {
			wrerr.WriteString(err.Error() + "\n")
			wrerr.Flush()
			continue
		}

		return idat, false
	}
}
