package slang

import (
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
	Name        string              `json:"name"`
	In          *PortDef            `json:"in"`
	Out         *PortDef            `json:"out"`
	Operators   []InstanceDef       `json:"operators"`
	Connections map[string][]string `json:"connections"`
	valid       bool
}

type InstanceDef struct {
	Operator   string                 `json:"operator"`
	Name       string                 `json:"name"`
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

	if err := d.In.Validate(); err != nil {
		return err
	}

	if err := d.Out.Validate(); err != nil {
		return err
	}

	alreadyUsedInsNames := make(map[string]bool)
	for _, insDef := range d.Operators {
		if err := insDef.Validate(); err != nil {
			return err
		}

		if _, ok := alreadyUsedInsNames[insDef.Name]; ok {
			return fmt.Errorf(`Colliding instance names within same parent operator: "%s"`, insDef.Name)
		}

		alreadyUsedInsNames[insDef.Name] = true

	}

	d.valid = true
	return nil
}

func (d *InstanceDef) Validate() error {
	if d.Name == "" {
		return fmt.Errorf(`instance name may not be empty`)
	}

	if strings.Contains(d.Name, " ") {
		return fmt.Errorf(`operator instance name may not contain spaces: "%s"`, d.Name)
	}

	if d.Operator == "" {
		return errors.New(`operator may not be empty`)
	}

	if strings.Contains(d.Operator, " ") {
		return fmt.Errorf(`operator may not contain spaces: "%s"`, d.Operator)
	}

	if d.In != nil {
		if err := d.In.Validate(); err != nil {
			return err
		}
	}

	if d.Out != nil {
		if err := d.Out.Validate(); err != nil {
			return err
		}
	}

	d.valid = true
	return nil
}

func getOperator(insDef InstanceDef, par *Operator) (*Operator, error) {
	return nil, errors.New("Not Implemented")
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

func MakeOperatorDeep(def OperatorDef, par *Operator) (*Operator, error) {
	if !def.valid {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	o, err := MakeOperator(def.Name, nil, *def.In, *def.Out, par)

	if err != nil {
		return nil, err
	}

	for _, childOpInsDef := range def.Operators {
		_, err := getOperator(childOpInsDef, o)

		if err != nil {
			return nil, err
		}
	}

	return o, nil
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

func parseConnection(connStr string, operator *Operator) (*Port, error) {
	if operator == nil {
		return nil, errors.New("operator must not be nil")
	}

	if len(connStr) == 0 {
		return nil, errors.New("empty connection string")
	}

	opSplit := strings.Split(connStr, ":")

	if len(opSplit) != 2 {
		return nil, errors.New("connection string malformed")
	}

	var o *Operator
	if len(opSplit[0]) == 0 {
		o = operator
	} else {
		var ok bool
		o, ok = operator.children[opSplit[0]]
		if !ok {
			return nil, errors.New("unknown operator")
		}
	}

	path := strings.Split(opSplit[1], ".")

	if len(path) == 0 {
		return nil, errors.New("connection string malformed")
	}

	var p *Port
	if path[0] == "in" {
		p = o.inPort
	} else if path[0] == "out" {
		p = o.outPort
	} else {
		return nil, errors.New(fmt.Sprintf("invalid direction: %s", path[1]))
	}

	for p.itemType == TYPE_STREAM {
		p = p.sub
	}

	for i := 1; i < len(path); i++ {
		if p.itemType != TYPE_MAP {
			return nil, errors.New("descending too deep")
		}

		k := path[i]
		var ok bool
		p, ok = p.subs[k]
		if !ok {
			return nil, errors.New(fmt.Sprintf("unknown port: %s", k))
		}

		for p.itemType == TYPE_STREAM {
			p = p.sub
		}
	}

	return p, nil
}