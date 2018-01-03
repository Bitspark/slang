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

type OperatorDef struct {
	Name        string   `json:"name"`
	In          *PortDef `json:"in"`
	Out         *PortDef `json:"out"`
	Operators   map[string]InstanceDef
	Connections map[string][]string `json:"connections"`
	valid       bool
}

type InstanceDef struct {
	Operator   string                 `json:"operator"`
	Properties map[string]interface{} `json:"properties"`
	In         *PortDef               `json:"in"`
	Out        *PortDef               `json:"out"`
	valid      bool
}

type OFunc func(in, out *Port)

func (d *OperatorDef) Validate() error {
	if d.Name == "" {
		return errors.New(`operator name may not be empty`)
	}

	if strings.Contains(d.Name, " ") {
		return fmt.Errorf(`operator name may not contain spaces: "%s"`, d.Name)
	}

	if d.In == nil || d.Out == nil {
		return errors.New(`ports must be defined`)
	}

	var portErr error
	portErr = d.In.Validate()
	if portErr != nil {
		return portErr
	}

	portErr = d.Out.Validate()
	if portErr != nil {
		return portErr
	}

	d.valid = true
	return nil
}

func (d *InstanceDef) validate() error {
	if d.Operator == "" {
		return errors.New(`operator may not be empty`)
	}

	if strings.Contains(d.Operator, " ") {
		return fmt.Errorf(`operator may not contain spaces: "%s"`, d.Operator)
	}

	if d.In != nil {
		if portErr := d.In.Validate(); portErr != nil {
			return portErr
		}
	}

	if d.Out != nil {
		if portErr := d.Out.Validate(); portErr != nil {
			return portErr
		}
	}
	d.valid = true
	return nil
}

func getOperatorDef(oprClass string) *OperatorDef {
	return &OperatorDef{}
}

func MakeOperator(name string, f OFunc, defIn, defOut PortDef, par *Operator) (*Operator, error) {
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
	def := OperatorDef{}
	json.Unmarshal([]byte(jsonDef), &def)

	err := def.Validate()

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
