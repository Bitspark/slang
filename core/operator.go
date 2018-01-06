package core

import (
	"errors"
	"fmt"
)

type OFunc func(in, out *Port, store interface{})

type Operator struct {
	name           string
	basePort       *Port
	inPort         *Port
	outPort        *Port
	parent         *Operator
	children       map[string]*Operator
	function       OFunc
	store          interface{}
	anyIdentifiers map[string]*PortDef
}

func NewOperator(name string, f OFunc, defIn, defOut PortDef, par *Operator) (*Operator, error) {
	o := &Operator{}
	o.function = f
	o.parent = par
	o.name = name
	o.children = make(map[string]*Operator)
	o.anyIdentifiers = make(map[string]*PortDef)

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

	o.collectAnys(defIn)
	o.collectAnys(defOut)

	return o, nil
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

func (o *Operator) SpecifyAny(identifier string, portDef PortDef) error {
	if !portDef.Valid() {
		if err := portDef.Validate(); err != nil {
			return err
		}
	}
	if ai, ok := o.anyIdentifiers[identifier]; ok {
		if ai != nil {
			if err := ai.Equals(portDef); err != nil {
				return errors.New(o.Name() + ", identifier " + identifier + ": " + err.Error())
			}
		}
		o.specifyAnyPorts(o.inPort, identifier, portDef)
		o.specifyAnyPorts(o.outPort, identifier, portDef)
		o.anyIdentifiers[identifier] = &portDef
		return nil
	}
	return errors.New(fmt.Sprintf(`unknown identifier "%s"`, identifier))
}

func (o *Operator) PropagateAnys() error {
	if err := o.propagateAnys(o.inPort); err != nil {
		return err
	}
	if err := o.propagateAnys(o.outPort); err != nil {
		return err
	}
	return nil
}

func (o *Operator) AnysSpecified() bool {
	for _, a := range o.anyIdentifiers {
		if a == nil {
			return false
		}
	}
	return true
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

func (o *Operator) collectAnys(portDef PortDef) {
	if portDef.Primitive() {
		return
	}

	if portDef.Type == "any" {
		o.anyIdentifiers[portDef.Any] = nil
	} else if portDef.Type == "stream" {
		o.collectAnys(*portDef.Stream)
	} else if portDef.Type == "map" {
		for _, e := range portDef.Map {
			o.collectAnys(e)
		}
	}
}

func (o *Operator) specifyAnyPorts(p *Port, identifier string, portDef PortDef) error {
	if p.primitive() {
		return nil
	}

	if p.itemType == TYPE_ANY {
		if p.any == identifier {
			np, _ := NewPort(o, portDef, p.direction)

			// Save source port and destinations
			src := p.src
			dsts := p.dests

			// Remove p from source port
			if src != nil {
				delete(src.dests, p)
			}

			*p = *np

			// Set source and add to destinations again
			p.src = src
			p.dests = dsts
			if src != nil {
				src.dests[p] = true
			}
		}
		return nil
	}

	if p.itemType == TYPE_STREAM {
		return o.specifyAnyPorts(p.sub, identifier, portDef)
	} else if p.itemType == TYPE_MAP {
		for _, e := range p.subs {
			if err := o.specifyAnyPorts(e, identifier, portDef); err != nil {
				return err
			}
		}
	}

	return errors.New("specify any ports: unknown type")
}

func (o *Operator) propagateAnys(p *Port) error {
	if p.primitive() {
		for d := range p.dests {
			if d.itemType == TYPE_ANY {
				p.Connect(d)
			}
		}
		return nil
	}

	if p.itemType == TYPE_ANY {
		return nil
	}

	if p.itemType == TYPE_STREAM {
		return o.propagateAnys(p.sub)
	} else if p.itemType == TYPE_MAP {
		for _, e := range p.subs {
			if err := o.propagateAnys(e); err != nil {
				return err
			}
		}
		return nil
	}

	return errors.New("propagate anys: unknown type")
}
