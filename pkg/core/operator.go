package core

import (
	"errors"
	"fmt"

	"github.com/Bitspark/slang/pkg/log"
	"github.com/google/uuid"
)

type OFunc func(op *Operator)
type CFunc func(op *Operator, dst, src *Port) error

var MAIN_SERVICE = "main"

type Operator struct {
	name        string
	defId       uuid.UUID
	defMeta     BlueprintMetaDef
	services    map[string]*Service
	delegates   map[string]*Delegate
	basePort    *Port
	parent      *Operator
	children    map[string]*Operator
	function    OFunc
	generics    Generics
	properties  Properties
	connectFunc CFunc
	elementary  uuid.UUID
	stopChannel chan bool
	stopped     bool
}

type Delegate struct {
	operator *Operator
	name     string
	inPort   *Port
	outPort  *Port
}

type Service struct {
	operator *Operator
	name     string
	inPort   *Port
	outPort  *Port
}

func NewOperator(name string, f OFunc, c CFunc, gens Generics, props Properties, def Blueprint) (*Operator, error) {
	props.Clean()

	o := &Operator{}
	o.defMeta = def.Meta
	o.defId = def.Id
	o.function = f
	o.connectFunc = c
	o.name = name
	o.elementary = def.Elementary
	o.generics = gens
	o.properties = props
	o.children = make(map[string]*Operator)

	var err error
	if err := def.PropertyDefs.GenericsSpecified(); err != nil {
		return nil, fmt.Errorf("%s: %s", "properties", err.Error())
	}

	o.services = make(map[string]*Service)
	for serName, ser := range def.ServiceDefs {
		o.services[serName], err = NewService(serName, o, *ser)
		if err != nil {
			return nil, err
		}
	}

	o.delegates = make(map[string]*Delegate)
	for delName, del := range def.DelegateDefs {
		o.delegates[delName], err = NewDelegate(delName, o, *del)
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

func (o *Operator) Id() uuid.UUID {
	return o.defId
}

func (o *Operator) Service(srv string) *Service {
	if s, ok := o.services[srv]; ok {
		return s
	}
	return nil
}

func (o *Operator) Main() *Service {
	return o.services[MAIN_SERVICE]
}

func (o *Operator) Delegate(del string) *Delegate {
	if d, ok := o.delegates[del]; ok {
		return d
	}
	return nil
}

func (o *Operator) Property(prop string) interface{} {
	return o.properties[prop]
}

func (o *Operator) SetProperties(properties Properties) {
	o.properties = properties
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
	o.stopChannel = make(chan bool, 1)
	o.stopped = false

	for _, srv := range o.services {
		srv.outPort.Open()
	}
	for _, dlg := range o.delegates {
		dlg.outPort.Open()
	}

	if o.function != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("%s:%s panicked: %s", o.Id(), o.Name(), r)
					o.Stop()
				}
			}()
			o.function(o)
		}()
	} else {
		for _, c := range o.children {
			c.Start()
		}
	}
}

func (o *Operator) Stop() {
	if o.stopped {
		return
	}

	o.stopChannel <- true
	o.stopped = true

	for _, srv := range o.services {
		srv.inPort.Close()
		srv.outPort.Close()
	}
	for _, dlg := range o.delegates {
		dlg.inPort.Close()
		dlg.outPort.Close()
	}

	for _, c := range o.children {
		c.Stop()
	}

	if o.parent != nil {
		o.parent.Stop()
	}
}

func (o *Operator) WaitForStop() {
	<-o.stopChannel
	o.stopChannel <- true
}

func (o *Operator) CheckStop() bool {
	select {
	case <-o.stopChannel:
		return true
	default:
		return false
	}
}

func (o *Operator) Stopped() bool {
	return o.stopped
}

func (o *Operator) Builtin() bool {
	return o.function != nil
}

func (o *Operator) Compile() (compiled int, depth int) {
	if o.Builtin() {
		return
	}

	// Go down
	for _, c := range o.children {
		cc, cd := c.Compile()
		compiled += cc
		if cd > depth {
			depth = cd
		}
	}

	if o.parent == nil {
		return
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
		c.name = o.name + "#" + c.name
		c.parent = o.parent
		o.parent.children[c.name] = c
	}

	// Remove this operator from its parent
	delete(o.parent.children, o.name)

	return compiled + 1, depth + 1
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

func (o *Operator) defineConnections(def *Blueprint) {
	for _, srv := range o.services {
		srv.outPort.defineConnections(def)
	}

	for _, dlg := range o.delegates {
		dlg.outPort.defineConnections(def)
	}
}

func (o *Operator) Define() (Blueprint, error) {
	var def Blueprint
	def.Id = o.defId
	def.Meta = o.defMeta
	def.ServiceDefs = make(map[string]*ServiceDef)
	def.DelegateDefs = make(map[string]*DelegateDef)
	def.Connections = make(map[string][]string)
	def.InstanceDefs = InstanceDefList{}

	for insName, child := range o.children {
		insDef := &InstanceDef{}
		insDef.Name = insName
		insDef.Operator = child.elementary
		insDef.Generics = child.generics
		insDef.Properties = child.properties
		insDef.Blueprint, _ = child.Define()
		def.InstanceDefs = append(def.InstanceDefs, insDef)
	}

	for _, srv := range o.services {
		srvDef := srv.Define()
		def.ServiceDefs[srv.name] = &srvDef
		srv.inPort.defineConnections(&def)
	}

	for _, dlg := range o.delegates {
		dlgDef := dlg.Define()
		def.DelegateDefs[dlg.name] = &dlgDef
		dlg.inPort.defineConnections(&def)
	}

	nonemptyConns := make(map[string][]string)
	for conn, conns := range def.Connections {
		if len(conns) != 0 {
			nonemptyConns[conn] = conns
		}
	}
	def.Connections = nonemptyConns

	if err := def.Validate(); err != nil {
		return def, err
	}

	return def, nil
}

// SERVICE

func NewService(name string, op *Operator, def ServiceDef) (*Service, error) {
	if !def.valid {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	srv := &Service{name: name, operator: op}

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
	return s.operator
}

func (s *Service) In() *Port {
	return s.inPort
}

func (s *Service) Out() *Port {
	return s.outPort
}

func (s *Service) Define() ServiceDef {
	var def ServiceDef
	def.In = s.inPort.Define()
	def.Out = s.outPort.Define()
	return def
}

// DELEGATE

func NewDelegate(name string, op *Operator, def DelegateDef) (*Delegate, error) {
	if !def.valid {
		err := def.Validate()
		if err != nil {
			return nil, err
		}
	}

	del := &Delegate{name: name, operator: op}

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
	return d.operator
}

func (d *Delegate) In() *Port {
	return d.inPort
}

func (d *Delegate) Out() *Port {
	return d.outPort
}

func (d *Delegate) Define() DelegateDef {
	var def DelegateDef
	def.In = d.inPort.Define()
	def.Out = d.outPort.Define()
	return def
}
