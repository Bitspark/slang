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
	"io/ioutil"
	"log"
	"os"
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

	// Handle STDIN
	go func() {
		for {
			text, _ := stdin.ReadString('\n')

			var dataIn interface{}
			err := json.Unmarshal([]byte(text), &dataIn)

			if err != nil {
				stderr.WriteString(fmt.Sprint(err))
				continue
			}

			fmt.Printf("<<< %v\n", dataIn)

			op.Main().In().Push(dataIn)
		}
	}()

	// Handle STDOUT
	go func() {
		for {
			dataOut := op.Main().Out().Pull()
			fmt.Printf(">>> %v\n", dataOut)
			text, err := json.Marshal(dataOut)

			if err != nil {
				stderr.WriteString(fmt.Sprint(err))
				continue
			}

			stdout.WriteString(string(text))
		}
	}()

	select {}
}
