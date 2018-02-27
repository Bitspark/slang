package core

import "errors"

type OFunc func(in, out *Port, dels map[string]*Delegate, store interface{})

type Operator struct {
	name      string
	basePort  *Port
	inPort    *Port
	outPort   *Port
	delegates map[string]*Delegate
	parent    *Operator
	children  map[string]*Operator
	function  OFunc
	store     interface{}
}

type Delegate struct {
	op      *Operator
	name    string
	inPort  *Port
	outPort *Port
}

func NewOperator(name string, f OFunc, defIn, defOut PortDef, delegates map[string]*DelegateDef) (*Operator, error) {
	o := &Operator{}
	o.function = f
	o.name = name
	o.children = make(map[string]*Operator)

	var err error

	o.inPort, err = NewPort(o, nil, defIn, DIRECTION_IN, nil)
	if err != nil {
		return nil, err
	}

	o.outPort, err = NewPort(o, nil, defOut, DIRECTION_OUT, nil)
	if err != nil {
		return nil, err
	}

	o.delegates = make(map[string]*Delegate)
	for delName, del := range delegates {
		o.delegates[delName], err = NewDelegate(delName, o, *del)
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

func (o *Operator) In() *Port {
	return o.inPort
}

func (o *Operator) Out() *Port {
	return o.outPort
}

func (o *Operator) Delegate(del string) *Delegate {
	return o.delegates[del]
}

func (o *Operator) Name() string {
	return o.name
}

func (o *Operator) BasePort() *Port {
	return o.basePort
}

func (o *Operator) SetParent(par *Operator) {
	o.parent = par
	if par != nil {
		par.children[o.name] = o
	}
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
		go o.function(o.inPort, o.outPort, o.delegates, o.store)
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

	for _, dlg := range o.delegates {
		dlg.In().Merge()
		dlg.Out().Merge()
	}

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

func (o *Operator) CorrectlyCompiled() error {
	for _, chld := range o.Children() {
		if len(chld.children) != 0 {
			return errors.New(chld.Name() + " not flat")
		}
		if err := chld.In().DirectlyConnected(); err != nil {
			return err
		}
		for _, del := range chld.delegates {
			if err := del.In().DirectlyConnected(); err != nil {
				return err
			}
		}
	}
	return nil
}

// DELEGATE

func NewDelegate(name string, op *Operator, def DelegateDef) (*Delegate, error) {
	if !def.valid {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	del := &Delegate{name: name, op: op}

	var err error
	if del.inPort, err = NewPort(op, del, def.In, DIRECTION_IN, nil); err != nil {
		return nil, err
	}
	if del.outPort, err = NewPort(op, del, def.Out, DIRECTION_OUT, nil); err != nil {
		return nil, err
	}

	return del, nil
}

func (d *Delegate) Name() string {
	return d.name
}

func (d *Delegate) Operator() *Operator {
	return d.op
}

func (d *Delegate) In() *Port {
	return d.inPort
}

func (d *Delegate) Out() *Port {
	return d.outPort
}
