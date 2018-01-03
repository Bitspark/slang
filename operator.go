package slang

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type Operator struct {
	name     string
	basePort *Port
	inPort   *Port
	outPort  *Port
	parent   *Operator
	children map[string]*Operator
	function OFunc
}

type operatorDef struct {
	Name      string   `json:"name"`
	In        *portDef `json:"in"`
	Out       *portDef `json:"out"`
	Operators map[string]struct {
		Class      string                 `json:"class"`
		Properties map[string]interface{} `json:"properties"`
	}
	Connections map[string][]string `json:"connections"`
	valid       bool
}

func (d *operatorDef) validate() error {
	fmt.Println(">>>> operaterDef", d, d.Name)

	if d.Name == "" {
		return errors.New(`operator name may not be empty`)
	}

	if strings.Contains(d.Name, " ") {
		return errors.New(fmt.Sprintf(`operator name may not contain spaces: "%s"`, d.Name))
	}

	if d.In == nil {
		return errors.New(`port in must be defined`)
	}

	var portErr error
	portErr = d.In.validate()
	if portErr != nil {
		return portErr
	}

	portErr = d.Out.validate()
	if portErr != nil {
		return portErr
	}

	d.valid = true
	return nil
}

type OFunc func(in, out *Port)

func MakeOperator(name string, f OFunc, defIn, defOut portDef, par *Operator) (*Operator, error) {
	o := &Operator{}
	o.function = f
	o.parent = par
	o.name = name
	o.children = make(map[string]*Operator)

	if par != nil {
		par.children[o.name] = o
	}

	var err error

	o.inPort, err = MakePort(o, defIn, DIRECTION_IN)
	if err != nil {
		return nil, err
	}

	o.outPort, err = MakePort(o, defOut, DIRECTION_OUT)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func ParseOperator(jsonDef string) (*Operator, error) {
	def := operatorDef{}
	json.Unmarshal([]byte(jsonDef), &def)

	err := def.validate()

	if err != nil {
		return &Operator{}, err
	}

	return &Operator{}, nil
}

func (o *Operator) InPort() *Port {
	return o.inPort
}

func (o *Operator) OutPort() *Port {
	return o.outPort
}

func (o *Operator) Name() string {
	return o.name
}

func (o *Operator) BasePort() *Port {
	return o.basePort
}

func (o *Operator) Parent() *Operator {
	return o.parent
}

func (o *Operator) Child(name string) *Operator {
	c, _ := o.children[name]
	return c
}

func (o *Operator) Start() {
	o.function(o.inPort, o.outPort)
}

func (o *Operator) Compile() bool {
	compiled := o.InPort().Merge()
	compiled = o.OutPort().Merge() || compiled
	return compiled
}
