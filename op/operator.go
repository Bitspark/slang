package op

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type OFunc func(in, out *Port, store interface{})

type Operator struct {
	name     string
	basePort *Port
	inPort   *Port
	outPort  *Port
	parent   *Operator
	children map[string]*Operator
	function OFunc
	store    interface{}
}

type OperatorDef struct {
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

func ParseOperatorDef(defStr string) OperatorDef {
	def := OperatorDef{}
	json.Unmarshal([]byte(defStr), &def)
	return def
}

func NewOperator(name string, f OFunc, defIn, defOut PortDef, par *Operator) (*Operator, error) {
	o := &Operator{}
	o.function = f
	o.parent = par
	o.name = name
	o.children = make(map[string]*Operator)

	if par != nil {
		par.children[o.name] = o
	}

	var err error

	o.inPort, err = NewPort(o, defIn, DIRECTION_IN)
	if err != nil {
		return nil, err
	}

	o.outPort, err = NewPort(o, defOut, DIRECTION_OUT)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func (d OperatorDef) Valid() bool {
	return d.valid
}

func (d InstanceDef) Valid() bool {
	return d.valid
}

func (d *OperatorDef) Validate() error {
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

func (o *Operator) In() *Port {
	return o.inPort
}

func (o *Operator) Out() *Port {
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

func (o *Operator) Children() map[string]*Operator {
	return o.children
}

func (o *Operator) Child(name string) *Operator {
	c, _ := o.children[name]
	return c
}

func (o *Operator) Start() {
	if o.function != nil {
		go o.function(o.inPort, o.outPort, o.store)
	} else {
		for _, c := range o.children {
			c.Start()
		}
	}
}

func (o *Operator) Stop() {
}

func (o *Operator) SetStore(store interface{}) {
	o.store = store
}

func (o *Operator) Builtin() bool {
	return o.function != nil
}

func (o *Operator) Compile() int {
	if o.Builtin() {
		return 0
	}

	compiled := 0

	// Go down
	for _, c := range o.children {
		compiled += c.Compile()
	}

	if o.parent == nil {
		return compiled
	}

	// Remove in and out port
	o.In().Merge()
	o.Out().Merge()

	// Move children to parent and rename instances
	for _, c := range o.children {
		c.name = o.name + "." + c.name
		c.parent = o.parent
		o.parent.children[c.name] = c
	}

	// Remove this operator from its parent
	delete(o.parent.children, o.name)

	return compiled + 1
}
