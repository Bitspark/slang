package core

import "errors"

type OFunc func(services map[string]*Service, dels map[string]*Delegate, store interface{})
type CFunc func(dest, src *Port) error

var DEFAULT_SERVICE = "main"

type Operator struct {
	name        string
	services    map[string]*Service
	delegates   map[string]*Delegate
	basePort    *Port
	parent      *Operator
	children    map[string]*Operator
	function    OFunc
	store       interface{}
	connectFunc CFunc
}

type Delegate struct {
	op      *Operator
	name    string
	inPort  *Port
	outPort *Port
}

type Service struct {
	op      *Operator
	name    string
	inPort  *Port
	outPort *Port
}

func NewOperator(name string, f OFunc, c CFunc, services map[string]*ServiceDef, delegates map[string]*DelegateDef) (*Operator, error) {
	o := &Operator{}
	o.function = f
	o.connectFunc = c
	o.name = name
	o.children = make(map[string]*Operator)

	var err error

	o.services = make(map[string]*Service)
	for serName, ser := range services {
		o.services[serName], err = NewService(serName, o, *ser)
		if err != nil {
			return nil, err
		}
	}

	o.delegates = make(map[string]*Delegate)
	for delName, del := range delegates {
		if delName == DEFAULT_SERVICE {
			return nil, errors.New("delegate must not be named " + DEFAULT_SERVICE + " (reserved for default service)")
		}
		o.delegates[delName], err = NewDelegate(delName, o, *del)
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

func (o *Operator) Service(srv string) *Service {
	if s, ok := o.services[srv]; ok {
		return s
	}
	return nil
}

func (o *Operator) DefaultService() *Service {
	return o.services[DEFAULT_SERVICE]
}

func (o *Operator) Delegate(del string) *Delegate {
	if d, ok := o.delegates[del]; ok {
		return d
	}
	return nil
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
		go o.function(o.services, o.delegates, o.store)
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
	for _, srv := range o.services {
		srv.In().Merge()
		srv.Out().Merge()
	}

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
		for _, srv := range chld.services {
			if err := srv.In().DirectlyConnected(); err != nil {
				return err
			}
		}
		for _, del := range chld.delegates {
			if err := del.In().DirectlyConnected(); err != nil {
				return err
			}
		}
	}
	return nil
}

// SERVICE

func NewService(name string, op *Operator, def ServiceDef) (*Service, error) {
	if !def.valid {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	srv := &Service{name: name, op: op}

	var err error
	if srv.inPort, err = NewPort(srv, nil, def.In, DIRECTION_IN); err != nil {
		return nil, err
	}
	if srv.outPort, err = NewPort(srv, nil, def.Out, DIRECTION_OUT); err != nil {
		return nil, err
	}

	return srv, nil
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Operator() *Operator {
	return s.op
}

func (s *Service) In() *Port {
	return s.inPort
}

func (s *Service) Out() *Port {
	return s.outPort
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
	if del.inPort, err = NewPort(nil, del, def.In, DIRECTION_IN); err != nil {
		return nil, err
	}
	if del.outPort, err = NewPort(nil, del, def.Out, DIRECTION_OUT); err != nil {
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
